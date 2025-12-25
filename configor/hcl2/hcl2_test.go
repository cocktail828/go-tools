package hcl2

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type NestedStruct struct {
	NestedField string `hcl:"nested_field"`
}

type ExampleStruct struct {
	Label   string            `hcl:"label,label"`
	ID      string            `hcl:"ID,label"`
	Name    string            `hcl:"name"`
	Age     int               `hcl:"age"`
	Active  *bool             `hcl:"active"`
	Details *NestedStruct     `hcl:"details,block"`
	Tags    map[string]string `hcl:"tags"`
}

type Config struct {
	Example []ExampleStruct `hcl:"example,block"`
}

func boolPtr(b bool) *bool { return &b }

type DatabaseConfig struct {
	Engine     string            `hcl:"engine,label"`
	Port       int               `hcl:"port"`
	Version    string            `hcl:"version,label"`
	AllowedIPs []string          `hcl:"allowed_ips"`
	Parameters map[string]string `hcl:"parameters"`
}

type S3Config struct {
	Engine     string            `hcl:"s3_engine"`
	AllowedIPs []string          `hcl:"s3_allowed_ips"`
	Parameters map[string]string `hcl:"s3_parameters"`
}

type AWSInstance struct {
	Timeout          time.Duration     `hcl:"timeout"`
	InstanceType     string            `hcl:"instance_type"`
	AMI              string            `hcl:"ami"`
	S3               S3Config          `hcl:"s3_config,block"`
	Tags             map[string]string `hcl:"tags"`
	Database         []DatabaseConfig  `hcl:"database_config,block"`
	SecurityGroupIDs []string          `hcl:"vpc_security_group_ids"`
}

func TestHCL2(t *testing.T) {
	{
		obj := Config{
			Example: []ExampleStruct{
				{
					Label:   "person_example",
					ID:      "123",
					Name:    "Alice",
					Age:     30,
					Active:  boolPtr(true),
					Details: &NestedStruct{NestedField: "Some details about Alice"},
					Tags: map[string]string{
						"role":    "admin",
						"team":    "engineering",
						"project": "terraform",
					},
				}, {
					Label:   "person_example",
					ID:      "234",
					Name:    "Alice",
					Age:     30,
					Active:  boolPtr(true),
					Details: &NestedStruct{NestedField: "Some details about Alice"},
					Tags: map[string]string{
						"role":    "admin",
						"team":    "engineering",
						"project": "terraform",
					},
				},
			},
		}

		t.Run("Config", func(t *testing.T) {
			data, _ := Marshal(obj)
			t.Log(string(data))

			target := Config{}
			if err := Unmarshal(data, &target); err != nil {
				t.Errorf("Error unmarshalling HCL: %v", err)
			}
			assert.EqualValues(t, obj, target)
		})
	}

	{
		obj := AWSInstance{
			Timeout:      time.Second,
			InstanceType: "t3.medium",
			AMI:          "ami-0c55b159cbfafe1f0",
			S3: S3Config{
				Engine:     "s3",
				AllowedIPs: []string{"192.168.1.0/24", "10.0.0.0/16"},
				Parameters: map[string]string{
					"bucket_name": "my-bucket",
				},
			},
			Tags: map[string]string{
				"Name":  "web-server",
				"Env":   "production",
				"Owner": "dev-team",
			},
			SecurityGroupIDs: []string{"sg-123456", "sg-7890ab"},
			Database: []DatabaseConfig{
				{
					Engine:     "mysql",
					Version:    "8.0",
					Port:       3306,
					AllowedIPs: []string{"192.168.1.0/24", "10.0.0.0/16"},
					Parameters: map[string]string{
						"max_connections": "1000",
						"character_set":   "utf8mb4",
					},
				},
				{
					Engine:     "postgreSQL",
					Version:    "13.0",
					Port:       5432,
					AllowedIPs: []string{"192.168.1.0/24", "10.0.0.0/16"},
					Parameters: map[string]string{
						"max_connections": "1000",
						"character_set":   "utf8mb4",
					},
				},
			},
		}

		t.Run("AWSInstance", func(t *testing.T) {
			data, _ := Marshal(obj)
			t.Log(string(data))

			target := AWSInstance{}
			if err := Unmarshal(data, &target); err != nil {
				t.Errorf("Error unmarshalling HCL: %v", err)
			}
			assert.EqualValues(t, obj, target)
		})
	}
}

func TestUnmarshal(t *testing.T) {
	type Config struct {
		Foo time.Duration `hcl:"foo"`
		// Baz float64       `hcl:"baz"`
	}

	const exampleConfig = `
	foo = "10s"
	baz = "10.5"
	`

	var config Config
	if err := Unmarshal([]byte(exampleConfig), &config); err != nil {
		t.Fatalf("Failed to load configuration: %s, config: %#v", err, config)
	}
	t.Logf("Configuration is %v\n", config)
}
