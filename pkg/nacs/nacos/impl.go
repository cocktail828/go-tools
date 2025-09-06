package nacos

import (
	"context"
	"fmt"
	"net"
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
	namingClient naming_client.INamingClient
	configClient config_client.IConfigClient
}

func NewNacosClient(namespaceID string, addrs []nacs.Endpoint) (*nacosClient, error) {
	sc := make([]constant.ServerConfig, 0, len(addrs))
	for _, ep := range addrs {
		sc = append(sc, constant.ServerConfig{
			IpAddr: ep.IP,
			Port:   uint64(ep.Port), // Nacos 默认端口
		})
	}

	cc := constant.ClientConfig{
		NamespaceId:         namespaceID, // 命名空间 ID
		TimeoutMs:           5000,        // 请求超时时间
		NotLoadCacheAtStart: true,        // 启动时不加载缓存
		LogDir:              "/tmp/nacos/log",
		CacheDir:            "/tmp/nacos/cache",
		LogLevel:            "info",
	}

	namingClient, err := clients.NewNamingClient(vo.NacosClientParam{
		ClientConfig:  &cc,
		ServerConfigs: sc,
	})
	if err != nil {
		return nil, err
	}

	configClient, err := clients.NewConfigClient(vo.NacosClientParam{
		ClientConfig:  &cc,
		ServerConfigs: sc,
	})
	if err != nil {
		return nil, err
	}

	return &nacosClient{
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

func (r *nacosClient) Register(inst nacs.Instance) error {
	host, port, err := splitHostPort(inst.Address)
	if err != nil {
		return err
	}

	success, err := r.namingClient.RegisterInstance(vo.RegisterInstanceParam{
		Ip:          host,
		Port:        uint64(port),
		Weight:      1, // 权重
		Enable:      true,
		Healthy:     true,
		Metadata:    inst.Metadata,
		ClusterName: inst.Cluster,
		ServiceName: inst.Name,
		GroupName:   inst.Group,
		Ephemeral:   true, // 临时实例
	})
	if err != nil {
		return err
	}

	if !success {
		return errors.Errorf("failed to register service instance: %+v", inst)
	}

	return nil
}

func (r *nacosClient) DeRegister(inst nacs.Instance) error {
	host, port, err := splitHostPort(inst.Address)
	if err != nil {
		return err
	}

	success, err := r.namingClient.DeregisterInstance(vo.DeregisterInstanceParam{
		Ip:          host,
		Port:        uint64(port),
		Cluster:     inst.Cluster,
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
		Cluster:  inst.ClusterName,
		Healthy:  inst.Healthy,
		Name:     inst.ServiceName,
		Address:  net.JoinHostPort(inst.Ip, fmt.Sprintf("%v", inst.Port)),
		Metadata: inst.Metadata,
	}
}

func (r *nacosClient) Discover(svc nacs.Service) ([]nacs.Instance, error) {
	param := vo.SelectInstancesParam{
		ServiceName: svc.Name,
		GroupName:   svc.Group,
		HealthyOnly: true, // 只返回健康实例
	}

	if svc.Cluster != "" {
		param.Clusters = append(param.Clusters, svc.Cluster)
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

func (r *nacosClient) Watch(svc nacs.Service, callback func([]nacs.Instance, error)) (context.CancelFunc, error) {
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
				result = append(result, toInstance(inst))
			}

			callback(result, nil)
		},
	}

	if svc.Cluster != "" {
		param.Clusters = append(param.Clusters, svc.Cluster)
	}

	return func() {
		r.namingClient.Unsubscribe(&param)
	}, r.namingClient.Subscribe(&param)
}

type GetOpt struct {
	ID    string
	Group string
}

func (o *GetOpt) Apply() {
	if o.Group == "" {
		o.Group = "DEFAULT_GROUP" // 默认分组
	}
}

func (r *nacosClient) Get(opts ...nacs.GetOpt) ([]byte, error) {
	var gopt *GetOpt
	for _, o := range opts {
		if o == nil {
			continue
		}
		if ic, ok := o.(*GetOpt); ok {
			gopt = ic
		}
	}
	gopt.Apply()

	content, err := r.configClient.GetConfig(vo.ConfigParam{
		DataId: gopt.ID,
		Group:  gopt.Group,
	})
	if err != nil {
		return nil, err
	}

	return []byte(content), nil
}

type MonitorOpt struct {
	ID    string
	Group string
}

func (o *MonitorOpt) Apply() {
	if o.Group == "" {
		o.Group = "DEFAULT_GROUP" // 默认分组
	}
}

func (r *nacosClient) Monitor(cb nacs.OnChange, opts ...nacs.MonitorOpt) (context.CancelFunc, error) {
	var mopt *MonitorOpt
	for _, o := range opts {
		if o == nil {
			continue
		}
		if ic, ok := o.(*MonitorOpt); ok {
			mopt = ic
		}
	}
	mopt.Apply()

	if err := r.configClient.ListenConfig(vo.ConfigParam{
		DataId:   mopt.ID,
		Group:    mopt.Group,
		OnChange: func(namespace, group, dataId, data string) { cb(nil) },
	}); err != nil {
		return nil, err
	}

	return func() {
		r.configClient.CancelListenConfig(vo.ConfigParam{
			DataId: mopt.ID,
			Group:  mopt.Group,
		})
	}, nil
}
