package hcl2

import (
	"fmt"
	"testing"
)

// 子结构体
type NestedStruct struct {
	NestedField string `hcl:"nested_field"`
}

// 主结构体
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

func TestHCL2(t *testing.T) {
	// 示例 HCL 数据
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

	// 初始化 Config 结构体
	var config Config

	// 使用 Unmarshal 解析 HCL 数据
	err := Unmarshal(hclData, &config)
	if err != nil {
		fmt.Println("Error unmarshalling HCL:", err)
		return
	}

	// 输出解析结果
	fmt.Printf("Parsed config: %+v\n", config)
}
