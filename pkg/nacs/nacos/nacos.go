package nacos

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strconv"

	"github.com/cocktail828/go-tools/pkg/nacs"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/model"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/pkg/errors"
)

var _ nacs.Registry = &nacosClient{}
var _ nacs.Configor = &nacosClient{}

type nacosClient struct {
	ID    string // config ID
	Group string // config group

	namingClient naming_client.INamingClient
	configClient config_client.IConfigClient
}

// nacos://$user:$password@$host:$port/$namespace
func NewNacosClient(uri string) (*nacosClient, error) {
	u, err := url.ParseRequestURI(uri)
	if err != nil {
		return nil, err
	}

	var username, password string
	if u.User != nil {
		username = u.User.Username()
		password, _ = u.User.Password()
	}

	cc := constant.NewClientConfig(
		constant.WithNamespaceId(u.Query().Get("namespace")),
		constant.WithNotLoadCacheAtStart(true),
		constant.WithAppName(u.Query().Get("appname")),
		constant.WithLogDir("./nacos/log"),
		constant.WithCacheDir("./nacos/cache"),
		constant.WithLogLevel("info"),
		constant.WithUsername(username),
		constant.WithPassword(password),
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

	group := u.Query().Get("group")
	if group == "" {
		group = constant.DEFAULT_GROUP
	}

	return &nacosClient{
		ID:           u.Query().Get("id"),
		Group:        group,
		namingClient: namingClient,
		configClient: configClient,
	}, nil
}

func (r *nacosClient) Close() error {
	r.namingClient.CloseClient()
	r.configClient.CloseClient()
	return nil
}

func splitHostPort(addr string) (string, int, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return "", 0, err
	}

	_p, err := strconv.Atoi(port)
	if err != nil {
		return "", 0, errors.Errorf("invalid address: %v", addr)
	}
	return host, _p, nil
}

func (r *nacosClient) Register(inst nacs.RegisterInstance) (context.CancelFunc, error) {
	host, port, err := splitHostPort(inst.Address)
	if err != nil {
		return nil, err
	}

	success, err := r.namingClient.RegisterInstance(vo.RegisterInstanceParam{
		Ip:          host,
		Port:        uint64(port),
		Weight:      1, // 权重
		Enable:      true,
		Healthy:     true,
		Metadata:    inst.Metadata,
		ServiceName: inst.Name,
		GroupName:   inst.Group,
		Ephemeral:   true, // 临时实例
	})
	if err != nil {
		return nil, err
	}

	if !success {
		return nil, errors.Errorf("failed to register service instance: %+v", inst)
	}

	return func() {
		r.DeRegister(nacs.DeRegisterInstance{
			Group:   inst.Group,
			Name:    inst.Name,
			Address: inst.Address,
		})
	}, nil
}

func (r *nacosClient) DeRegister(inst nacs.DeRegisterInstance) error {
	host, port, err := splitHostPort(inst.Address)
	if err != nil {
		return err
	}

	success, err := r.namingClient.DeregisterInstance(vo.DeregisterInstanceParam{
		Ip:          host,
		Port:        uint64(port),
		ServiceName: inst.Name,
		GroupName:   inst.Group,
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
	return nacs.Instance{
		Enable:   inst.Enable,
		Healthy:  inst.Healthy,
		Name:     inst.ServiceName,
		Address:  net.JoinHostPort(inst.Ip, fmt.Sprintf("%v", inst.Port)),
		Metadata: inst.Metadata,
	}
}

func (r *nacosClient) Discover(svc nacs.Service) ([]nacs.Instance, error) {
	if svc.Group == "" {
		svc.Group = constant.DEFAULT_GROUP
	}

	param := vo.SelectInstancesParam{
		ServiceName: svc.Name,
		GroupName:   svc.Group,
		HealthyOnly: true, // 只返回健康实例
	}

	instances, err := r.namingClient.SelectInstances(param)
	if err != nil {
		return nil, err
	}

	result := make([]nacs.Instance, 0, len(instances))
	for _, inst := range instances {
		val := toInstance(inst)
		val.Group = svc.Group
		val.Name = svc.Name
		result = append(result, val)
	}

	return result, nil
}

func (r *nacosClient) Watch(svc nacs.Service, callback func([]nacs.Instance, error)) (context.CancelFunc, error) {
	if svc.Group == "" {
		svc.Group = constant.DEFAULT_GROUP
	}

	param := vo.SubscribeParam{
		ServiceName: svc.Name,
		GroupName:   svc.Group,
		SubscribeCallback: func(instances []model.Instance, err error) {
			if err != nil {
				callback(nil, err)
				return
			}

			result := make([]nacs.Instance, 0, len(instances))
			for _, inst := range instances {
				val := toInstance(inst)
				val.Group = svc.Group
				val.Name = svc.Name
				result = append(result, val)
			}

			callback(result, nil)
		},
	}

	return func() {
		r.namingClient.Unsubscribe(&param)
	}, r.namingClient.Subscribe(&param)
}

func (r *nacosClient) Load() ([]byte, error) {
	if r.ID == "" {
		return nil, errors.New("nacos: data ID cannot be empty")
	}

	content, err := r.configClient.GetConfig(vo.ConfigParam{
		DataId: r.ID,
		Group:  r.Group,
	})
	if err != nil {
		return nil, err
	}

	return []byte(content), nil
}

func (r *nacosClient) Monitor(cb nacs.OnChange) (context.CancelFunc, error) {
	if cb == nil {
		cb = func(err error, args ...any) {}
	}

	if err := r.configClient.ListenConfig(vo.ConfigParam{
		DataId:   r.ID,
		Group:    r.Group,
		OnChange: func(namespace, group, dataId, data string) { cb(nil, namespace, group, dataId, data) },
	}); err != nil {
		return nil, err
	}

	return func() {
		r.configClient.CancelListenConfig(vo.ConfigParam{
			DataId: r.ID,
			Group:  r.Group,
		})
	}, nil
}
