package netx

import "testing"

func TestParseRI(t *testing.T) {
	ri := []struct {
		uri string
		ri  RI
	}{
		{
			uri: "mq://a/b?c=e",
			ri: RI{
				Schema: "mq",
				Path:   "/a/b",
				Query:  "c=e",
			},
		},
		{
			uri: "/a/b?c=e",
			ri: RI{
				Schema: "",
				Path:   "/a/b",
				Query:  "c=e",
			},
		},
	}

	for _, v := range ri {
		ri, err := ParseRI(v.uri)
		if err != nil {
			t.Fatal(err)
		}

		if *ri != v.ri {
			t.Fatalf("expect ri %v, got %v", v.ri, *ri)
		}
	}
}
