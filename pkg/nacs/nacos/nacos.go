package nacos

import (
	"context"
	"net/url"
	"strconv"
	"strings"

	"github.com/cocktail828/go-tools/pkg/nacs"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/model"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/pkg/errors"
)

var _ nacs.Registry = &nacosClient{}
var _ nacs.Configor = &nacosClient{}

type nacosClient struct {
	inGroup      string // group
	inService    string // app
	inVersion    string // version
	namingClient *refNamingClient
	configClient *refConfigClient
}

// ServiceName returns the service name in nacos format: $service@$version
// It is used to identify the service in nacos.
func (r *nacosClient) ServiceName() string { return nacs.Compose(r.inService, r.inVersion) }

// ConfigID returns the config ID in nacos format: $service.$version
// It is used to identify the config in nacos.
func (r *nacosClient) ConfigID() string { return r.inService + "_" + r.inVersion }

// nacos://$user:$password@$host:$port/$namespace/$group/$service/$version
func NewNacosClient(u *url.URL) (*nacosClient, error) {
	var username, password string
	if u.User != nil {
		username = u.User.Username()
		password, _ = u.User.Password()
	}

	parts := strings.Split(u.Path, "/")
	if len(parts) < 5 || parts[1] == "" || parts[2] == "" || parts[3] == "" || parts[4] == "" {
		return nil, errors.New("nacos path format error: expect nacos://$user:$password@$host:$port/$namespace/$group/$service/$version")
	}
	namespace, group, service, version := parts[1], parts[2], parts[3], parts[4]

	cc := constant.NewClientConfig(
		constant.WithNamespaceId(namespace),
		constant.WithNotLoadCacheAtStart(true),
		constant.WithAppName(service),
		constant.WithUsername(username),
		constant.WithPassword(password),
		constant.WithTimeoutMs(3000),
		constant.WithUpdateThreadNum(2),
	)

	port, err := strconv.Atoi(u.Port())
	if err != nil {
		return nil, err
	}

	sc := []constant.ServerConfig{{
		IpAddr: u.Hostname(),
		Port:   uint64(port),
	}}

	namingClient, err := clients.NewNamingClient(vo.NacosClientParam{
		ClientConfig:  cc,
		ServerConfigs: sc,
	})
	if err != nil {
		return nil, err
	}

	configClient, err := clients.NewConfigClient(vo.NacosClientParam{
		ClientConfig:  cc,
		ServerConfigs: sc,
	})
	if err != nil {
		return nil, err
	}

	return &nacosClient{
		inGroup:      group,
		inService:    service,
		inVersion:    version,
		namingClient: &refNamingClient{INamingClient: namingClient},
		configClient: &refConfigClient{IConfigClient: configClient},
	}, nil
}

// Share will create a new client with different group, service, version
// But shared the same instance of namingClient and configClient
// Notice, namingClient and configClient will be close only when the last 'nacosClient' is closed
func (r *nacosClient) Share(group, service, version string) *nacosClient {
	if group == "" {
		group = constant.DEFAULT_GROUP
	}

	return &nacosClient{
		inGroup:      group,
		inService:    service,
		inVersion:    version,
		namingClient: r.namingClient.Share(),
		configClient: r.configClient.Share(),
	}
}

func (r *nacosClient) Close() error {
	r.namingClient.Close()
	r.configClient.Close()
	return nil
}

func (r *nacosClient) Register(inst nacs.Instance) (context.CancelFunc, error) {
	success, err := r.namingClient.RegisterInstance(vo.RegisterInstanceParam{
		Ip:          inst.Host,
		Port:        uint64(inst.Port),
		Weight:      100, // 权重
		Enable:      true,
		Healthy:     true,
		Metadata:    inst.Metadata,
		ServiceName: r.ServiceName(),
		GroupName:   r.inGroup,
		Ephemeral:   true, // ephemeral instance
	})
	if err != nil {
		return nil, err
	}

	if !success {
		return nil, errors.Errorf("failed to register service instance: %+v", inst)
	}

	return func() {
		r.DeRegister(nacs.Instance{
			Service: r.ServiceName(),
			Host:    inst.Host,
			Port:    inst.Port,
		})
	}, nil
}

func (r *nacosClient) DeRegister(inst nacs.Instance) error {
	success, err := r.namingClient.DeregisterInstance(vo.DeregisterInstanceParam{
		Ip:          inst.Host,
		Port:        uint64(inst.Port),
		ServiceName: r.ServiceName(),
		GroupName:   r.inGroup,
		Ephemeral:   true,
	})
	if err != nil {
		return err
	}

	if !success {
		return errors.Errorf("failed to deregister service instance: %+v", inst)
	}

	return nil
}

func toInstance(inst model.Instance) nacs.Instance {
	// nacos combine group and service name with '@@'
	_, svc, found := strings.Cut(inst.ServiceName, "@@")
	if !found {
		svc = inst.ServiceName
	}

	return nacs.Instance{
		Enable:   inst.Enable,
		Healthy:  inst.Healthy,
		Service:  svc,
		Host:     inst.Ip,
		Port:     uint(inst.Port),
		Metadata: inst.Metadata,
	}
}

func (r *nacosClient) Discover() ([]nacs.Instance, error) {
	param := vo.SelectInstancesParam{
		ServiceName: r.ServiceName(),
		GroupName:   r.inGroup,
		HealthyOnly: true, // 只返回健康实例
	}

	instances, err := r.namingClient.SelectInstances(param)
	if err != nil {
		return nil, err
	}

	result := make([]nacs.Instance, 0, len(instances))
	for _, inst := range instances {
		result = append(result, toInstance(inst))
	}

	return result, nil
}

func (r *nacosClient) Watch(callback func([]nacs.Instance, error)) (context.CancelFunc, error) {
	param := vo.SubscribeParam{
		ServiceName: r.ServiceName(),
		GroupName:   r.inGroup,
		SubscribeCallback: func(instances []model.Instance, err error) {
			if err != nil {
				callback(nil, err)
				return
			}

			result := make([]nacs.Instance, 0, len(instances))
			for _, inst := range instances {
				result = append(result, toInstance(inst))
			}

			callback(result, nil)
		},
	}

	return func() {
		r.namingClient.Unsubscribe(&param)
	}, r.namingClient.Subscribe(&param)
}

func (r *nacosClient) Load() ([]byte, error) {
	id := r.ConfigID()
	if id == "" {
		return nil, errors.New("nacos: data ID cannot be empty")
	}

	content, err := r.configClient.GetConfig(vo.ConfigParam{
		DataId: id,
		Group:  r.inGroup,
	})
	if err != nil {
		return nil, err
	}

	return []byte(content), nil
}

func (r *nacosClient) Monitor(cb func(name string, payload []byte, err error)) (context.CancelFunc, error) {
	if cb == nil {
		cb = func(name string, payload []byte, err error) {}
	}

	id := r.ConfigID()
	if id == "" {
		return nil, errors.New("nacos: data ID cannot be empty")
	}

	if err := r.configClient.ListenConfig(vo.ConfigParam{
		DataId:   id,
		Group:    r.inGroup,
		OnChange: func(namespace, group, dataId, data string) { cb(dataId, []byte(data), nil) },
	}); err != nil {
		return nil, err
	}

	return func() {
		r.configClient.CancelListenConfig(vo.ConfigParam{
			DataId: id,
			Group:  r.inGroup,
		})
	}, nil
}
