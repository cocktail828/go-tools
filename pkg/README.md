# Reference
```
golang.org/x/sync/errgroup
golang.org/x/sync/singleflight
golang.org/x/sync/semaphore

# sql builder & executer
github.com/Masterminds/squirrel
github.com/jmoiron/sqlx

#
github.com/tidwall/gjson
github.com/mitchellh/mapstructure

https://github.com/soheilhy/cmux
https://github.com/xtaci/smux
```


# JSON Schema Benchmark
```go
package main

import (
	"fmt"
	"testing"

	"github.com/xeipuuv/gojsonschema"
)

var schemaBody = []byte(`
{
    "$schema": "http://json-schema.org/draft-07/schema#",
    "type": "object",
    "required": [
        "name",
        "email",
        "age"
    ],
    "properties": {
        "name": {
            "type": "string",
            "minLength": 1,
            "maxLength": 50,
            "pattern": "^[a-zA-Z\\s]+$"
        },
        "email": {
            "type": "string",
            "format": "email",
            "maxLength": 100
        },
        "age": {
            "type": "integer",
            "minimum": 0,
            "maximum": 150
        },
        "status": {
            "type": "string",
            "enum": [
                "active",
                "inactive",
                "pending"
            ]
        },
        "tags": {
            "type": "array",
            "items": {
                "type": "string",
                "minLength": 1
            },
            "minItems": 0,
            "maxItems": 10
        },
        "profile": {
            "type": "object",
            "properties": {
                "bio": {
                    "type": "string",
                    "maxLength": 500
                },
                "website": {
                    "type": "string",
                    "format": "uri"
                }
            },
            "additionalProperties": false
        }
    },
    "additionalProperties": false
}
`)

type Profile struct {
	Bio     string `json:"bio"`
	Website string `json:"website"`
}

type UserData struct {
	Name    string   `json:"name"`
	Email   string   `json:"email"`
	Age     int      `json:"age"`
	Status  string   `json:"status"`
	Tags    []string `json:"tags"`
	Profile Profile  `json:"profile"`
}

func BenchmarkJsonSchema_go(b *testing.B) {
	// validator, err := gojsonschema.NewSchemaLoader().Compile(gojsonschema.NewReferenceLoader("file://./schema.json"))
	validator, err := gojsonschema.NewSchemaLoader().Compile(gojsonschema.NewBytesLoader(schemaBody))
	if err != nil {
		panic(err.Error())
	}

	// var userData = UserData{
	// 	Name:   "John Doe",
	// 	Email:  "john@example.com",
	// 	Age:    30,
	// 	Status: "active",
	// 	Tags:   []string{"developer", "golang"},
	// 	Profile: Profile{
	// 		Bio:     "Software developer",
	// 		Website: "https://example.com",
	// 	},
	// }

	f := func() {
		// result, err := validator.Validate(gojsonschema.NewGoLoader(userData))
		result, err := validator.Validate(gojsonschema.NewBytesLoader([]byte(`{
			"name": "John Doe",
			"email": "john@example.com",
			"age": 30,
			"status": "active",
			"tags": ["developer", "golang"],
			"profile": {
				"bio": "Software developer",
				"website": "https://example.com"
			}
		}`)))
		if err != nil {
			panic(err.Error())
		}
		if !result.Valid() {
			panic(fmt.Errorf("验证失败: %v", result.Errors()))
		}
	}

	f()
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			f()
		}
	})
}
```