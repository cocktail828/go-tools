package nacs_test

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/cocktail828/go-tools/pkg/nacs"
	"github.com/cocktail828/go-tools/pkg/nacs/etcd"
	"github.com/cocktail828/go-tools/pkg/nacs/nacos"
	"github.com/cocktail828/go-tools/pkg/nacs/native"
	"github.com/cocktail828/go-tools/pkg/nacs/static"
)

// Example demonstrates how to use the Registry interface with etcd
func ExampleRegistry_etcd() {
	// Parse the etcd URI
	u, err := url.Parse("etcd://localhost:2379?namespace=myapp&service=userservice&version=v1.0.0")
	if err != nil {
		log.Fatal(err)
	}

	// Create etcd registry client
	registry, err := etcd.NewEtcdClient(u)
	if err != nil {
		log.Fatal(err)
	}
	defer registry.Close()

	// Register a service instance
	instance := nacs.Instance{
		Service: "userservice",
		Version: "v1.0.0",
		Host:    "192.168.1.100",
		Port:    8080,
		Meta: map[string]string{
			"region": "us-west",
			"zone":   "zone-a",
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Register returns a cancel function for automatic deregistration
	deregister, err := registry.Register(ctx, instance)
	if err != nil {
		log.Fatal(err)
	}
	defer deregister() // Automatically deregister on exit

	// Discover service instances
	instances, err := registry.Discover(ctx)
	if err != nil {
		log.Fatal(err)
	}

	for _, inst := range instances {
		fmt.Printf("Found instance: %s:%d (service=%s, version=%s)\n",
			inst.Host, inst.Port, inst.Service, inst.Version)
	}

	// Watch for service changes
	cancelWatch, err := registry.Watch(func(instances []nacs.Instance, err error) {
		if err != nil {
			log.Printf("Watch error: %v", err)
			return
		}
		fmt.Printf("Service instances updated: %d instances\n", len(instances))
	})
	if err != nil {
		log.Fatal(err)
	}
	defer cancelWatch()

	// Keep running...
	time.Sleep(time.Minute)
}

// Example demonstrates how to use the Registry interface with nacos
func ExampleRegistry_nacos() {
	// Parse the nacos URI
	u, err := url.Parse("nacos://admin:admin@localhost:8848?namespace=myapp&group=DEFAULT_GROUP&service=userservice&version=v1.0.0")
	if err != nil {
		log.Fatal(err)
	}

	// Create nacos registry client
	registry, err := nacos.NewNacosClient(u)
	if err != nil {
		log.Fatal(err)
	}
	defer registry.Close()

	// Register a service instance
	instance := nacs.Instance{
		Service: "userservice",
		Version: "v1.0.0",
		Host:    "192.168.1.100",
		Port:    8080,
		Meta: map[string]string{
			"weight": "100",
		},
	}

	ctx := context.Background()
	deregister, err := registry.Register(ctx, instance)
	if err != nil {
		log.Fatal(err)
	}
	defer deregister()

	// Use the registry...
}

// Example demonstrates how to use the Configor interface with etcd
func ExampleConfigor_etcd() {
	u, err := url.Parse("etcd://localhost:2379?namespace=myapp&service=userservice&version=v1.0.0")
	if err != nil {
		log.Fatal(err)
	}

	configor, err := etcd.NewEtcdClient(u)
	if err != nil {
		log.Fatal(err)
	}
	defer configor.Close()

	// Load configuration
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	info, err := configor.Load(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Config loaded from %s: %s\n", info.DataID, string(info.Payload))

	// Monitor configuration changes
	cancelMonitor, err := configor.Monitor(func(info nacs.ConfigInfo, err error) {
		if err != nil {
			log.Printf("Monitor error: %v", err)
			return
		}
		fmt.Printf("Config updated (%s): %s\n", info.DataID, string(info.Payload))
	})
	if err != nil {
		log.Fatal(err)
	}
	defer cancelMonitor()

	// Keep running...
	time.Sleep(time.Minute)
}

// Example demonstrates how to use the Configor interface with file
func ExampleConfigor_file() {
	u, err := url.Parse("file:///etc/myapp/config.json")
	if err != nil {
		log.Fatal(err)
	}

	configor, err := native.NewFileConfigor(u)
	if err != nil {
		log.Fatal(err)
	}
	defer configor.Close()

	// Load configuration
	ctx := context.Background()
	info, err := configor.Load(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Config from %s: %s\n", info.DataID, string(info.Payload))

	// Monitor file changes
	cancelMonitor, err := configor.Monitor(func(info nacs.ConfigInfo, err error) {
		if err != nil {
			log.Printf("File change error: %v", err)
			return
		}
		fmt.Printf("Config file updated (%s): %s\n", info.DataID, string(info.Payload))
	})
	if err != nil {
		log.Fatal(err)
	}
	defer cancelMonitor()
}

// Example demonstrates how to use static configuration
func ExampleConfigor_static() {
	// Static configuration for testing or default values
	configData := []byte(`{"database": "localhost:5432", "cache": "redis:6379"}`)
	configor := static.NewStaticConfigor(configData)
	defer configor.Close()

	ctx := context.Background()
	info, err := configor.Load(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Static config (%s): %s\n", info.DataID, string(info.Payload))

	// Monitor does nothing for static config (never changes)
	_, _ = configor.Monitor(func(info nacs.ConfigInfo, err error) {
		// This callback is never called for static config
	})
}

// Example of sharing the same base client for multiple services
func Example_sharing() {
	// Create base nacos client
	u, err := url.Parse("nacos://admin:admin@localhost:8848?namespace=myapp&group=DEFAULT_GROUP&service=userservice&version=v1.0.0")
	if err != nil {
		log.Fatal(err)
	}

	client1, err := nacos.NewNacosClient(u)
	if err != nil {
		log.Fatal(err)
	}
	defer client1.Close()

	// Share the base client for different service
	client2, err := client1.Share("DEFAULT_GROUP", "orderservice", "v2.0.0")
	if err != nil {
		log.Fatal(err)
	}
	defer client2.Close()

	// Now you can use client1 for userservice and client2 for orderservice
	// They share the same underlying connection to nacos
}
