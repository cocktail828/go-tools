package nacos

import (
	"context"
	"fmt"
	"net"
	"strconv"

	"github.com/cocktail828/go-tools/pkg/nacs/configuration"
	"github.com/cocktail828/go-tools/pkg/nacs/naming"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/model"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/pkg/errors"
)

var _ naming.Registry = &nacosClient{}
var _ configuration.Configor = &nacosClient{}

type nacosClient struct {
	namingClient naming_client.INamingClient
	configClient config_client.IConfigClient
}

func NewNacosClient(namespaceID string, addrs []naming.Endpoint) (*nacosClient, error) {
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

func (r *nacosClient) Register(inst naming.Instance) error {
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

func (r *nacosClient) DeRegister(inst naming.Instance) error {
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

func toInstance(inst model.Instance) naming.Instance {
	return naming.Instance{
		Enable:   inst.Enable,
		Cluster:  inst.ClusterName,
		Healthy:  inst.Healthy,
		Name:     inst.ServiceName,
		Address:  net.JoinHostPort(inst.Ip, fmt.Sprintf("%v", inst.Port)),
		Metadata: inst.Metadata,
	}
}

func (r *nacosClient) Discover(svc naming.Service) ([]naming.Instance, error) {
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

	result := make([]naming.Instance, 0, len(instances))
	for _, inst := range instances {
		result = append(result, toInstance(inst))
	}

	return result, nil
}

func (r *nacosClient) Watch(svc naming.Service, callback func([]naming.Instance, error)) (context.CancelFunc, error) {
	param := vo.SubscribeParam{
		ServiceName: svc.Name,
		GroupName:   svc.Group,
		SubscribeCallback: func(instances []model.Instance, err error) {
			if err != nil {
				callback(nil, err)
				return
			}

			result := make([]naming.Instance, 0, len(instances))
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

func normalize(cfg *configuration.Config) {
	if cfg.Group == "" {
		cfg.Group = "DEFAULT_GROUP" // 默认分组
	}
}

func (r *nacosClient) Get(cfg configuration.Config) ([]byte, error) {
	normalize(&cfg)
	content, err := r.configClient.GetConfig(vo.ConfigParam{
		DataId: cfg.ID,
		Group:  cfg.Group,
	})
	if err != nil {
		return nil, err
	}

	return []byte(content), nil
}

func (r *nacosClient) Set(cfg configuration.Config, payload []byte) error {
	normalize(&cfg)
	_, err := r.configClient.PublishConfig(vo.ConfigParam{
		DataId:  cfg.ID,
		Group:   cfg.Group,
		Content: string(payload),
	})
	return err
}

func (r *nacosClient) Delete(cfg configuration.Config) error {
	normalize(&cfg)
	_, err := r.configClient.DeleteConfig(vo.ConfigParam{
		DataId: cfg.ID,
		Group:  cfg.Group,
	})
	return err
}

func (r *nacosClient) Monitor(cfg configuration.Config, listener configuration.Listener) (context.CancelFunc, error) {
	normalize(&cfg)
	err := r.configClient.ListenConfig(vo.ConfigParam{
		DataId: cfg.ID,
		Group:  cfg.Group,
		OnChange: func(namespace, group, dataId, data string) {
			listener(configuration.Config{
				ID:    dataId,
				Group: group,
			}, []byte(data), nil)
		},
	})
	if err != nil {
		return nil, err
	}

	return func() {
		r.configClient.CancelListenConfig(vo.ConfigParam{
			DataId: cfg.ID,
			Group:  cfg.Group,
		})
	}, nil
}
