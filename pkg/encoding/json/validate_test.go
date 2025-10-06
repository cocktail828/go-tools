package json

import (
	"bytes"
	"testing"

	"github.com/go-playground/validator/v10"
)

var (
	validateFn ValidateFn = validator.New().Var
)

func TestValidate(t *testing.T) {
	jsonObj := map[string]any{
		"struct1": struct {
			A int
		}{A: 1},
		"structArr": []struct {
			A int
		}{{A: 1}, {A: 2}},
		"app_id": "abc", // 符合规则
		"usr_id": "xxx", // 违反 required
		"user": map[string]any{
			"name":  "A",             // 违反 min=2
			"age":   17,              // 违反 min=18
			"email": "invalid-email", // 违反 email
			"address": map[string]any{
				"city": "Beijing",
				"zip":  "10000", // 违反 len=6
			},
			"sinm": struct{ X int }{X: 10},
		},
		"tags": []any{
			"",           // 违反 tags[0].required 和 tags[*].max=10
			"toolongtag", // 违反 tags[*].max=10
		},
	}
	jsonData, _ := Marshal(jsonObj)

	dec := NewDecoder(bytes.NewReader(jsonData))
	dec.WithValidateRules(validateFn, map[string]string{
		"app_id": "required,min=2",
		// "usr_id":            "required",
		// "user.name":         "required,min=2,max=20",
		// "user.age":          "required,min=18,max=120",
		// "user.email":        "required,email",
		// "user.address.city": "required",
		// "user.address.zip":  "required,len=6",
		// "tags[0]":           "required,min=1",
		// "tags[*]":           "max=10",
	})

	obj := map[string]any{}
	if err := dec.Decode(&obj); err != nil {
		t.Fatal(err)
	}
}

func BenchmarkValidate(b *testing.B) {
	jsonObj := map[string]any{
		"struct1": struct {
			A int
		}{A: 1},
		"structArr": []struct {
			A int
		}{{A: 1}, {A: 2}},
		"app_id": "abc", // 符合规则
		"usr_id": "xxx", // 违反 required
		"user": map[string]any{
			"name":  "A",             // 违反 min=2
			"age":   17,              // 违反 min=18
			"email": "invalid-email", // 违反 email
			"address": map[string]any{
				"city": "Beijing",
				"zip":  "10000", // 违反 len=6
			},
			"sinm": struct{ X int }{X: 10},
		},
		"tags": []any{
			"",           // 违反 tags[0].required 和 tags[*].max=10
			"toolongtag", // 违反 tags[*].max=10
		},
	}
	jsonData, _ := Marshal(jsonObj)

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			dec := NewDecoder(bytes.NewReader(jsonData))
			dec.WithValidateRules(validateFn, map[string]string{
				"app_id": "required,min=2",
				// "usr_id":            "required",
				// "user.name":         "required,min=2,max=20",
				// "user.age":          "required,min=18,max=120",
				// "user.email":        "required,email",
				// "user.address.city": "required",
				// "user.address.zip":  "required,len=6",
				// "tags[0]":           "required,min=1",
				// "tags[*]":           "max=10",
			})
			obj := map[string]any{}
			if err := dec.Decode(&obj); err != nil {
				b.Fatal(err)
			}
		}
	})
}
