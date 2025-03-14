package nacos

import (
	"context"

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

func (r *nacosClient) Register(svc nacs.Service) error {
	success, err := r.namingClient.RegisterInstance(vo.RegisterInstanceParam{
		Ip:          svc.IP,
		Port:        uint64(svc.Port),
		Weight:      1, // 权重
		Enable:      true,
		Healthy:     true,
		Metadata:    svc.Metadata,
		ClusterName: svc.Cluster,
		ServiceName: svc.Name,
		GroupName:   svc.Group,
		Ephemeral:   true, // 临时实例
	})
	if err != nil {
		return err
	}

	if !success {
		return errors.Errorf("failed to register svc: %+v", svc)
	}

	return nil
}

func (r *nacosClient) Deregister(svc nacs.Service) error {
	success, err := r.namingClient.DeregisterInstance(vo.DeregisterInstanceParam{
		Ip:          svc.IP,
		Port:        uint64(svc.Port),
		Cluster:     svc.Cluster,
		ServiceName: svc.Name,
		GroupName:   svc.Group,
		Ephemeral:   true,
	})
	if err != nil {
		return err
	}

	if !success {
		return errors.Errorf("failed to deregister svc: %+v", svc)
	}

	return nil
}

func (r *nacosClient) Discover(sf nacs.ServiceFilter) ([]nacs.Service, error) {
	param := vo.SelectInstancesParam{
		ServiceName: sf.Name,
		GroupName:   sf.Group,
		HealthyOnly: true, // 只返回健康实例
	}

	if sf.Cluster != "" {
		param.Clusters = append(param.Clusters, sf.Cluster)
	}

	instances, err := r.namingClient.SelectInstances(param)
	if err != nil {
		return nil, err
	}

	result := make([]nacs.Service, 0, len(instances))
	for _, svc := range instances {
		result = append(result, nacs.Service{
			ID:       svc.InstanceId,
			Name:     svc.ServiceName,
			IP:       svc.Ip,
			Port:     int(svc.Port),
			Metadata: svc.Metadata,
		})
	}

	return result, nil
}

func (r *nacosClient) Watch(sf nacs.ServiceFilter, callback func([]nacs.Service, error)) error {
	param := vo.SubscribeParam{
		ServiceName: sf.Name,
		GroupName:   sf.Group,
		SubscribeCallback: func(services []model.Instance, err error) {
			if err != nil {
				callback(nil, err)
				return
			}

			instances := make([]nacs.Service, 0, len(services))
			for _, svc := range services {
				instances = append(instances, nacs.Service{
					ID:       svc.InstanceId,
					Name:     svc.ServiceName,
					IP:       svc.Ip,
					Port:     int(svc.Port),
					Metadata: svc.Metadata,
				})
			}

			callback(instances, nil)
		},
	}

	if sf.Cluster != "" {
		param.Clusters = append(param.Clusters, sf.Cluster)
	}

	return r.namingClient.Subscribe(&param)
}

func normalize(cfg *nacs.Config) {
	if cfg.Group == "" {
		cfg.Group = "DEFAULT_GROUP" // 默认分组
	}
}

func (r *nacosClient) GetConfig(cfg nacs.Config) ([]byte, error) {
	normalize(&cfg)
	content, err := r.configClient.GetConfig(vo.ConfigParam{
		DataId: cfg.Fname,
		Group:  cfg.Group,
	})
	if err != nil {
		return nil, err
	}

	return []byte(content), nil
}

func (r *nacosClient) SetConfig(cfg nacs.Config, payload []byte) error {
	normalize(&cfg)
	_, err := r.configClient.PublishConfig(vo.ConfigParam{
		DataId:  cfg.Fname,
		Group:   cfg.Group,
		Content: string(payload),
	})
	return err
}

func (r *nacosClient) DeleteConfig(cfg nacs.Config) error {
	normalize(&cfg)
	_, err := r.configClient.DeleteConfig(vo.ConfigParam{
		DataId: cfg.Fname,
		Group:  cfg.Group,
	})
	return err
}

func (r *nacosClient) WatchConfig(cfg nacs.Config, listener nacs.ConfigListener) (context.CancelFunc, error) {
	normalize(&cfg)
	err := r.configClient.ListenConfig(vo.ConfigParam{
		DataId: cfg.Fname,
		Group:  cfg.Group,
		OnChange: func(namespace, group, dataId, data string) {
			listener(nacs.Config{
				Fname: dataId,
				Group: group,
			}, []byte(data), nil)
		},
	})
	if err != nil {
		return nil, err
	}

	return func() {
		r.configClient.CancelListenConfig(vo.ConfigParam{
			DataId: cfg.Fname,
			Group:  cfg.Group,
		})
	}, nil
}
