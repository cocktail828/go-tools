package netx

import (
	"net/url"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func checkEqual(t *testing.T, u0, u1 *URI) {
	// Check basic fields
	if u0.Scheme != u1.Scheme {
		t.Errorf("Scheme = %v, want %v", u0.Scheme, u1.Scheme)
	}
	if u0.User.String() != u1.User.String() {
		t.Errorf("User = %v, want %v", u0.User, u1.User)
	}
	// Check addresses
	if !reflect.DeepEqual(u0.Addresses, u1.Addresses) {
		t.Errorf("Addresses = %v, want %v", u0.Addresses, u1.Addresses)
	}
	if u0.Path != u1.Path {
		t.Errorf("Path = %v, want %v", u0.Path, u1.Path)
	}
	if u0.RawQuery != u1.RawQuery {
		t.Errorf("RawQuery = %v, want %v", u0.RawQuery, u1.RawQuery)
	}
}

func TestParseNonStandardURI(t *testing.T) {
	tests := []struct {
		name        string
		rawURL      string
		expectedURI *URI
		expectErr   bool
	}{{
		name:   "Single address with path and query",
		rawURL: "http://127.0.0.1:8080/asd/asdf?a=b&b=c",
		expectedURI: &URI{
			Scheme:    "http",
			Addresses: []string{"127.0.0.1:8080"},
			Path:      "/asd/asdf",
			RawQuery:  "a=b&b=c",
		},
		expectErr: false,
	}, {
		name:   "Multiple addresses with path and query",
		rawURL: "http://127.0.0.1:8080,127.0.0.1:8081/asd/asdf?a=b&b=c",
		expectedURI: &URI{
			Scheme:    "http",
			Addresses: []string{"127.0.0.1:8080", "127.0.0.1:8081"},
			Path:      "/asd/asdf",
			RawQuery:  "a=b&b=c",
		},
		expectErr: false,
	}, {
		name:   "With authentication",
		rawURL: "http://user:password@127.0.0.1:8080/asd/asdf?a=b&b=c",
		expectedURI: &URI{
			Scheme:    "http",
			User:      url.UserPassword("user", "password"),
			Addresses: []string{"127.0.0.1:8080"},
			Path:      "/asd/asdf",
			RawQuery:  "a=b&b=c",
		},
		expectErr: false,
	}, {
		name:   "With authentication and multiple addresses",
		rawURL: "http://user:password@127.0.0.1:8080,127.0.0.1:8081/asd/asdf?a=b&b=c",
		expectedURI: &URI{
			Scheme:    "http",
			User:      url.UserPassword("user", "password"),
			Addresses: []string{"127.0.0.1:8080", "127.0.0.1:8081"},
			Path:      "/asd/asdf",
			RawQuery:  "a=b&b=c",
		},
		expectErr: false,
	}, {
		name:   "Username only authentication",
		rawURL: "ftp://anonymous@192.168.1.1/pub",
		expectedURI: &URI{
			Scheme:    "ftp",
			User:      url.User("anonymous"),
			Addresses: []string{"192.168.1.1"},
			Path:      "/pub",
		},
		expectErr: false,
	}, {
		name:   "No path",
		rawURL: "grpc://localhost:9000",
		expectedURI: &URI{
			Scheme:    "grpc",
			Addresses: []string{"localhost:9000"},
			Path:      "/",
		},
		expectErr: false,
	}, {
		name:      "Empty URL",
		rawURL:    "",
		expectErr: true,
	}, {
		name:      "Missing scheme",
		rawURL:    "127.0.0.1:8080/path",
		expectErr: true,
	}, {
		name:      "Empty address in list",
		rawURL:    "http://127.0.0.1:8080,,127.0.0.1:8081/path",
		expectErr: true,
	}, {
		name:   "HTTPS scheme",
		rawURL: "https://secure.example.com:443/api/v1/users",
		expectedURI: &URI{
			Scheme:    "https",
			Addresses: []string{"secure.example.com:443"},
			Path:      "/api/v1/users",
		},
		expectErr: false,
	}, {
		name:   "URL with HTML entities in query",
		rawURL: "http://example.com/path?param1=value1&amp;param2=value2",
		expectedURI: &URI{
			Scheme:    "http",
			Addresses: []string{"example.com"},
			Path:      "/path",
			RawQuery:  "param1=value1&amp;param2=value2",
		},
		expectErr: false,
	},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uri, err := ParseURI(tt.rawURL)
			if err != nil && !tt.expectErr {
				t.Errorf("ParseURI(%q) error = %v, expectErr %v", tt.rawURL, err, tt.expectErr)
				return
			}

			if !tt.expectErr && uri != nil {
				checkEqual(t, tt.expectedURI, uri)
			}
		})
	}
}

func TestParseURIWithEncoding(t *testing.T) {
	tests := []struct {
		name        string
		rawURL      string
		expectedURI *URI
		expectErr   bool
	}{{
		name:   "Simple URL without encoding",
		rawURL: "http://127.0.0.1:8080/path/to/resource",
		expectedURI: &URI{
			Scheme:    "http",
			Addresses: []string{"127.0.0.1:8080"},
			Path:      "/path/to/resource",
			RawPath:   "/path/to/resource",
		},
		expectErr: false,
	}, {
		name:   "URL with encoded username and password",
		rawURL: "http://user%40domain:pass%23word@127.0.0.1:8080/path",
		expectedURI: &URI{
			Scheme:    "http",
			User:      url.UserPassword("user@domain", "pass#word"),
			Addresses: []string{"127.0.0.1:8080"},
			Path:      "/path",
			RawPath:   "/path",
		},
		expectErr: false,
	}, {
		name:   "URL with encoded path",
		rawURL: "http://example.com/path%20with%20spaces/%E4%B8%AD%E6%96%87",
		expectedURI: &URI{
			Scheme:    "http",
			Addresses: []string{"example.com"},
			Path:      "/path with spaces/中文",
			RawPath:   "/path%20with%20spaces/%E4%B8%AD%E6%96%87",
		},
		expectErr: false,
	}, {
		name:   "Full URL with all encoded components",
		rawURL: "https://user%2Btag:secret%21pass@server1:9000,server2:9001/api%2Fv1/users%3Fid%3D123",
		expectedURI: &URI{
			Scheme:    "https",
			User:      url.UserPassword("user+tag", "secret!pass"),
			Addresses: []string{"server1:9000", "server2:9001"},
			Path:      "/api/v1/users?id=123",
			RawPath:   "/api%2Fv1/users%3Fid%3D123",
		},
		expectErr: false,
	}, {
		name:   "Multiple addresses with query parameters",
		rawURL: "http://127.0.0.1:8080,127.0.0.1:8081/path?param1=value1&param2=value+with+space",
		expectedURI: &URI{
			Scheme:    "http",
			Addresses: []string{"127.0.0.1:8080", "127.0.0.1:8081"},
			Path:      "/path",
			RawPath:   "/path",
			RawQuery:  "param1=value1&param2=value+with+space",
		},
		expectErr: false,
	}, {
		name:      "Invalid URL encoding",
		rawURL:    "http://user%zz:password@example.com",
		expectErr: true,
	},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uri, err := ParseURI(tt.rawURL)

			if err != nil && !tt.expectErr {
				t.Errorf("ParseURI() error = %v, expectErr %v", err, tt.expectErr)
				return
			}

			if !tt.expectErr && uri != nil {
				checkEqual(t, tt.expectedURI, uri)
			}
		})
	}
}

func TestURIStringMethod(t *testing.T) {
	uri := &URI{
		Scheme:    "https",
		User:      url.UserPassword("admin", "secret"),
		Addresses: []string{"server1:443", "server2:443"},
		RawPath:   "/api/v1/users",
		RawQuery:  "sort=name&limit=10",
	}
	expected := "https://admin:secret@server1:443,server2:443/api/v1/users?sort=name&limit=10"

	assert.Equal(t, expected, uri.String())
}
