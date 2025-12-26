package nacos

import (
	"context"
	"net/url"
	"strings"

	"github.com/cocktail828/go-tools/pkg/nacs"
	"github.com/nacos-group/nacos-sdk-go/v2/model"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/pkg/errors"
)

var _ nacs.Registry = &NacosClient{}
var _ nacs.Configor = &NacosClient{}

type NacosClient struct {
	inGroup    string // group
	inService  string // service
	inVersion  string // version
	baseClient *BaseNacosClient
}

// ServiceName returns the service name in nacos format: $service@$version
// It is used to identify the service in nacos.
func (c *NacosClient) ServiceName() string { return c.inService + "@" + c.inVersion }

// ConfigID returns the config ID in nacos format: $service.$version
// It is used to identify the config in nacos.
func (c *NacosClient) ConfigID() string { return c.inService + "_" + c.inVersion }

// nacos://$user:$password@$host:$port?group=$group&service=$service&version=$version
func NewNacosClient(u *url.URL) (*NacosClient, error) {
	baseClient, err := NewBaseNacosClient(u)
	if err != nil {
		return nil, err
	}

	query := u.Query()
	return baseClient.Spawn(query.Get("group"), query.Get("service"), query.Get("version"))
}

// Ancestor returns the base nacos client
func (c *NacosClient) Ancestor() *BaseNacosClient { return c.baseClient }

// Share will create a new client with different group, service, version
// But shared the same instance of namingClient and configClient
// Notice, namingClient and configClient will be close only when the last 'NacosClient' is closed
func (c *NacosClient) Share(group, service, version string) (*NacosClient, error) {
	return c.baseClient.Spawn(group, service, version)
}

func (c *NacosClient) Close() error {
	return c.baseClient.Close()
}

func (c *NacosClient) Register(host string, port uint, meta map[string]string) (context.CancelFunc, error) {
	success, err := c.baseClient.namingClient.RegisterInstance(vo.RegisterInstanceParam{
		Ip:          host,
		Port:        uint64(port),
		Weight:      100, // weight
		Enable:      true,
		Healthy:     true,
		Metadata:    meta,
		ServiceName: c.ServiceName(),
		GroupName:   c.inGroup,
		Ephemeral:   true, // ephemeral instance
	})
	if err != nil {
		return nil, err
	}

	if !success {
		return nil, errors.Errorf("failed to register service instance: %s:%d", host, port)
	}

	return func() { c.DeRegister(host, port) }, nil
}

func (c *NacosClient) DeRegister(host string, port uint) error {
	success, err := c.baseClient.namingClient.DeregisterInstance(vo.DeregisterInstanceParam{
		Ip:          host,
		Port:        uint64(port),
		ServiceName: c.ServiceName(),
		GroupName:   c.inGroup,
		Ephemeral:   true,
	})
	if err != nil {
		return err
	}

	if !success {
		return errors.Errorf("failed to deregister service instance: %s:%d", host, port)
	}

	return nil
}

func toInstances(insts []model.Instance) []nacs.Instance {
	result := make([]nacs.Instance, 0, len(insts))
	for _, inst := range insts {
		if !inst.Enable || !inst.Healthy {
			continue
		}

		// nacos combine group and service name with '@@'
		_, svc, found := strings.Cut(inst.ServiceName, "@@")
		if !found {
			svc = inst.ServiceName
		}

		result = append(result, nacs.Instance{
			Name: svc,
			Host: inst.Ip,
			Port: uint(inst.Port),
			Meta: inst.Metadata,
		})
	}

	return result
}

func (c *NacosClient) Discover() ([]nacs.Instance, error) {
	param := vo.SelectInstancesParam{
		ServiceName: c.ServiceName(),
		GroupName:   c.inGroup,
		HealthyOnly: true, // 只返回健康实例
	}

	instances, err := c.baseClient.namingClient.SelectInstances(param)
	if err != nil {
		return nil, err
	}

	return toInstances(instances), nil
}

func (c *NacosClient) Watch(callback func([]nacs.Instance, error)) (context.CancelFunc, error) {
	param := vo.SubscribeParam{
		ServiceName: c.ServiceName(),
		GroupName:   c.inGroup,
		SubscribeCallback: func(instances []model.Instance, err error) {
			if err != nil {
				callback(nil, err)
				return
			}

			callback(toInstances(instances), nil)
		},
	}

	return func() {
		c.baseClient.namingClient.Unsubscribe(&param)
	}, c.baseClient.namingClient.Subscribe(&param)
}

func (c *NacosClient) Load() ([]byte, error) {
	id := c.ConfigID()
	if id == "" {
		return nil, errors.New("nacos: data ID cannot be empty")
	}

	content, err := c.baseClient.configClient.GetConfig(vo.ConfigParam{
		DataId: id,
		Group:  c.inGroup,
	})
	if err != nil {
		return nil, err
	}

	return []byte(content), nil
}

func (c *NacosClient) Monitor(cb func(name string, payload []byte, err error)) (context.CancelFunc, error) {
	if cb == nil {
		cb = func(name string, payload []byte, err error) {}
	}

	id := c.ConfigID()
	if id == "" {
		return nil, errors.New("nacos: data ID cannot be empty")
	}

	if err := c.baseClient.configClient.ListenConfig(vo.ConfigParam{
		DataId:   id,
		Group:    c.inGroup,
		OnChange: func(namespace, group, dataId, data string) { cb(dataId, []byte(data), nil) },
	}); err != nil {
		return nil, err
	}

	return func() {
		c.baseClient.configClient.CancelListenConfig(vo.ConfigParam{
			DataId: id,
			Group:  c.inGroup,
		})
	}, nil
}
