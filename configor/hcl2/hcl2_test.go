package hcl2

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type NestedStruct struct {
	NestedField string `hcl:"nested_field"`
}

type ExampleStruct struct {
	Label   string            `hcl:"label,label"` // 使用 label 表示块的标签, 第一个 label
	ID      string            `hcl:"label,label"` // 使用 label 表示块的标签, 第二个 label
	Name    string            `hcl:"name"`
	Age     int               `hcl:"age"`
	Active  *bool             `hcl:"active"`
	Details *NestedStruct     `hcl:"details,block"` // 子结构体，作为子块
	Tags    map[string]string `hcl:"tags"`
}

type Config struct {
	Example []ExampleStruct `hcl:"example,block"` // 支持多实例的块
}

func boolPtr(b bool) *bool { return &b }

func TestHCL2(t *testing.T) {
	hclData := []byte(`
        // 这里注释支持 #, //
        # label1 = person_example
        // label2 = 123
        example "person_example" "123" {
            name   = "Alice"
            age    = 30
            active = true

            // block, 所以这里不使用 '='
            details {
                nested_field = "Some details about Alice"
            }

            tags = {
                "role"    = "admin"
                "team"    = "engineering"
                "project" = "terraform"
            }
        }

        example "person_example" "234" {
            name   = "Alice"
            age    = 30
            active = true

            details {
                nested_field = "Some details about Alice"
            }

            tags = {
                "role"    = "admin"
                "team"    = "engineering"
                "project" = "terraform"
            }
        }
    `)

	var c Config
	err := Unmarshal(hclData, &c)
	if err != nil {
		t.Errorf("Error unmarshalling HCL: %v", err)
	}
	assert.EqualValues(t, Config{
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
	}, c)
}
