package tries

import (
	"fmt"
	"reflect"
	"sync"
	"testing"
)

func TestRouterBasicRoute(t *testing.T) {
	router := Router{}
	router.Add("/hello", "hello-handler")
	router.Add("/world", "world-handler")

	handle, _ := router.Route("/hello")
	if handle != "hello-handler" {
		t.Errorf("Expected hello-handler, got %v", handle)
	}

	handle, _ = router.Route("/world")
	if handle != "world-handler" {
		t.Errorf("Expected world-handler, got %v", handle)
	}

	handle, _ = router.Route("/not-exist")
	if handle != nil {
		t.Errorf("Expected not found for non-existent route")
	}
}

func TestRouterPrefixMatching(t *testing.T) {
	router := Router{}
	router.Add("/api", "api-root")
	router.Add("/api/users", "api-users")
	router.Add("/api/users/profile", "api-users-profile")
	router.Add("/api/products", "api-products")

	tests := []struct {
		path     string
		expected any
		exists   bool
	}{{
		path:     "/api",
		expected: "api-root",
		exists:   true,
	}, {
		path:     "/api/users",
		expected: "api-users",
		exists:   true,
	}, {
		path:     "/api/users/profile",
		expected: "api-users-profile",
		exists:   true,
	}, {
		path:     "/api/products",
		expected: "api-products",
		exists:   true,
	}, {
		path:     "/api/nope",
		expected: nil,
		exists:   false,
	}, {
		path:     "/ap",
		expected: nil,
		exists:   false,
	}}

	for _, tt := range tests {
		handle, _ := router.Route(tt.path)
		if tt.exists {
			if handle != tt.expected {
				t.Errorf("Path %s: expected handle %v, got %v", tt.path, tt.expected, handle)
			}
		} else {
			if handle != nil {
				t.Errorf("Path %s: expected not found, got %v", tt.path, handle)
			}
		}
	}
}

func TestRouterParamRoute(t *testing.T) {
	router := Router{}
	router.Add("/users/:id", "user-handler")
	router.Add("/users/:id/profile", "user-profile-handler")
	router.Add("/files/*filepath", "file-handler")

	tests := []struct {
		path           string
		expected       any
		expectedParams Params
	}{{
		path:           "/users/123",
		expected:       "user-handler",
		expectedParams: Params{{Key: "id", Value: "123"}},
	}, {
		path:           "/users/456/profile",
		expected:       "user-profile-handler",
		expectedParams: Params{{Key: "id", Value: "456"}},
	}, {
		path:           "/files/documents/report.pdf",
		expected:       "file-handler",
		expectedParams: Params{{Key: "filepath", Value: "/documents/report.pdf"}},
	}}

	for _, tt := range tests {
		handle, params := router.Route(tt.path)
		if handle != tt.expected {
			t.Errorf("Path %s: expected handle %v, got %v", tt.path, tt.expected, handle)
		}
		if !reflect.DeepEqual(params, tt.expectedParams) {
			t.Errorf("Path %s: expected params %v, got %v", tt.path, tt.expectedParams, params)
		}
	}
}

func TestRouterRouteConflict_ParamVsStatic(t *testing.T) {
	t.Run("ParamFirstThenStatic", func(t *testing.T) {
		router := Router{}

		if err := router.Add("/users/:id", "user-id"); err != nil {
			t.Errorf("Expected no error for /users/:id, got %v", err)
		}
		if err := router.Add("/users/new", "new-user"); err == nil {
			t.Errorf("Expected error for /users/new, but got no error")
		}

		handle, params := router.Route("/users/123")
		if handle != "user-id" || params[0].Value != "123" {
			t.Errorf("Expected user-id with param 123 for /users/123, got %v with params %v", handle, params)
		}
	})

	t.Run("StaticFirstThenParam", func(t *testing.T) {
		router := Router{}

		if err := router.Add("/books/new", "new-book"); err != nil {
			t.Errorf("Expected no error for /books/new, got %v", err)
		}
		if err := router.Add("/books/:id", "book-id"); err == nil {
			t.Errorf("Expected error for /books/:id, but got no error")
		}

		handle, _ := router.Route("/books/new")
		if handle != "new-book" {
			t.Errorf("Expected new-book for /books/new, got %v", handle)
		}
	})

	t.Run("CatchAllFirstThenStatic", func(t *testing.T) {
		router := Router{}

		if err := router.Add("/files/*filepath", "files-handler"); err != nil {
			t.Errorf("Expected no error for /files/*filepath, got %v", err)
		}
		if err := router.Add("/files/special", "special-files"); err == nil {
			t.Errorf("Expected error for /files/special, but got no error")
		}

		handle, params := router.Route("/files/special")
		if handle != "files-handler" || params[0].Value != "/special" {
			t.Errorf("Expected files-handler with param special for /files/special, got %v with params %v", handle, params)
		}
	})

	t.Run("StaticFirstThenCatchAll", func(t *testing.T) {
		router := Router{}

		if err := router.Add("/musics/special", "special-musics"); err != nil {
			t.Errorf("Expected no error for /musics/special, got %v", err)
		}
		if err := router.Add("/musics/*filepath", "musics-handler"); err == nil {
			t.Errorf("Expected error for /musics/*filepath, but got no error")
		}

		handle, _ := router.Route("/musics/special")
		if handle != "special-musics" {
			t.Errorf("Expected special-musics for /musics/special, got %v", handle)
		}
	})
}

func TestRouterTrailingSlash(t *testing.T) {
	router := Router{}
	if err := router.Add("/api", "api-without-slash"); err != nil {
		t.Errorf("Expected no error for /api, got %v", err)
	}
	if err := router.Add("/api/", "api-with-slash"); err == nil {
		t.Errorf("Expected error for /api/, but got no error")
	}

	handle, params := router.Route("/api")
	if handle != "api-without-slash" {
		t.Errorf("Expected api-without-slash for /api, got %v", handle)
	}
	if len(params) != 0 {
		t.Errorf("Expected empty params for /api, got %v", params)
	}
}

func TestRouterCaseInsensitivePath(t *testing.T) {
	router := Router{}
	node := router.tree

	// Add some routes
	node.addRoute("/Users/Profile", "user-profile-handler")

	// Test case-insensitive lookup
	ciPath, found := node.findCaseInsensitivePath("/users/profile", false)
	if !found || string(ciPath) != "/Users/Profile" {
		t.Errorf("Expected case-insensitive match for /users/profile, got %s, found: %v", ciPath, found)
	}

	ciPath, found = node.findCaseInsensitivePath("/USERS/PROFILE", false)
	if !found || string(ciPath) != "/Users/Profile" {
		t.Errorf("Expected case-insensitive match for /USERS/PROFILE, got %s, found: %v", ciPath, found)
	}

	ciPath, found = node.findCaseInsensitivePath("/nonexistent", false)
	if found {
		t.Errorf("Expected no match for /nonexistent, but found %s", ciPath)
	}
}

func TestRouterConcurrency(t *testing.T) {
	router := Router{}
	var wg sync.WaitGroup
	routeCount := 100

	// Add routes concurrently
	for i := range routeCount {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			path := fmt.Sprintf("/concurrent/route/%d", id)
			handle := fmt.Sprintf("handle-%d", id)
			router.Add(path, handle)
		}(i)
	}
	wg.Wait()

	// Verify all routes were added correctly
	for i := range routeCount {
		path := fmt.Sprintf("/concurrent/route/%d", i)
		expectedHandle := fmt.Sprintf("handle-%d", i)
		handle, params := router.Route(path)
		if handle != expectedHandle {
			t.Errorf("Expected handle %s for path %s, got %v", expectedHandle, path, handle)
		}
		if len(params) != 0 {
			t.Errorf("Expected empty params for path %s, got %v", path, params)
		}
	}
}

func TestRouterComplexPrefixMatching(t *testing.T) {
	router := Router{}
	router.Add("/a", "a")
	router.Add("/ab", "ab")
	router.Add("/abc", "abc")
	router.Add("/abcd", "abcd")
	router.Add("/abcde", "abcde")

	tests := []struct {
		path     string
		expected any
	}{{
		path:     "/a",
		expected: "a",
	}, {
		path:     "/ab",
		expected: "ab",
	}, {
		path:     "/abc",
		expected: "abc",
	}, {
		path:     "/abcd",
		expected: "abcd",
	}, {
		path:     "/abcde",
		expected: "abcde",
	}, {
		path:     "/abcdef",
		expected: nil,
	}}

	for _, tt := range tests {
		handle, _ := router.Route(tt.path)
		if handle != tt.expected {
			t.Errorf("Path %s: expected handle %v, got %v", tt.path, tt.expected, handle)
		}
	}
}

func TestCountParams(t *testing.T) {
	tests := []struct {
		path     string
		expected uint8
	}{{
		path:     "/simple/path",
		expected: 0,
	}, {
		path:     "/users/:id",
		expected: 1,
	}, {
		path:     "/users/:id/posts/:postId",
		expected: 2,
	}, {
		path:     "/files/*filepath",
		expected: 1,
	}, {
		path:     "/api/:version/users/:id/*resource",
		expected: 3,
	}}

	for _, tt := range tests {
		count := countParams(tt.path)
		if count != tt.expected {
			t.Errorf("Path %s: expected %d params, got %d", tt.path, tt.expected, count)
		}
	}
}

func TestParamsByName(t *testing.T) {
	params := Params{
		{Key: "id", Value: "123"},
		{Key: "name", Value: "test"},
		{Key: "type", Value: "user"},
	}

	if val := params.ByName("id"); val != "123" {
		t.Errorf("Expected '123' for key 'id', got '%s'", val)
	}

	if val := params.ByName("name"); val != "test" {
		t.Errorf("Expected 'test' for key 'name', got '%s'", val)
	}

	if val := params.ByName("nonexistent"); val != "" {
		t.Errorf("Expected empty string for nonexistent key, got '%s'", val)
	}
}

func BenchmarkRouterAdd(b *testing.B) {
	router := Router{}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		path := fmt.Sprintf("/benchmark/route/%d", i%100)
		router.Add(path, fmt.Sprintf("handle-%d", i%100))
	}
}

func BenchmarkRouterLookup(b *testing.B) {
	router := Router{}

	// Add 100 routes
	for i := 0; i < 100; i++ {
		path := fmt.Sprintf("/benchmark/route/%d", i)
		router.Add(path, fmt.Sprintf("handle-%d", i))
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		path := fmt.Sprintf("/benchmark/route/%d", i%100)
		router.Route(path)
	}
}
