package dsn

import "testing"

func TestParseRI(t *testing.T) {
	ri := []struct {
		uri string
		ri  RI
	}{
		{
			uri: "mq://a/b?c=e&d=f",
			ri: RI{
				Scheme: "mq",
				Path:   "/a/b",
				Query:  "c=e&d=f",
			},
		},
		{
			uri: "mq://a/b",
			ri: RI{
				Scheme: "mq",
				Path:   "/a/b",
				Query:  "",
			},
		},
		{
			uri: "/a/b?c=e",
			ri: RI{
				Scheme: "",
				Path:   "/a/b",
				Query:  "c=e",
			},
		},
	}

	for _, v := range ri {
		if ri := Parse(v.uri); ri != v.ri {
			t.Fatalf("expect ri %v, got %v", v.ri, ri)
		}
	}
}
