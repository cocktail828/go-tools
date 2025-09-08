package hcl2

import "github.com/hashicorp/hcl/v2/hclsimple"

func Unmarshal(data []byte, v any) error {
	return hclsimple.Decode("example.hcl", data, nil, v)
}
