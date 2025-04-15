package router

import (
	"reflect"
	"testing"
)

func TestParams(t *testing.T) {
	ps := Params{
		Param{"param1", "value1"},
		Param{"param2", "value2"},
		Param{"param3", "value3"},
	}
	for i := range ps {
		if val := ps.ByName(ps[i].Key); val != ps[i].Value {
			t.Errorf("Wrong value for %s: Got %s; Want %s", ps[i].Key, val, ps[i].Value)
		}
	}
	if val := ps.ByName("noKey"); val != "" {
		t.Errorf("Expected empty string for not found key; got: %s", val)
	}
}

func TestRouter(t *testing.T) {
	router := New()

	routed := false
	router.Handle("noop", "/user/:name", func(c Context) error {
		routed = true
		want := Params{Param{"name", "gopher"}}
		if !reflect.DeepEqual(c.Params, want) {
			t.Fatalf("wrong wildcard values: want %v, got %v", want, c.Params)
		}
		return nil
	})

	router.ServeURI("noop://user/gopher")
	if !routed {
		t.Fatal("routing failed")
	}
}

func TestRouterRoot(t *testing.T) {
	router := New()
	recv := catchPanic(func() {
		router.Handle("noop", "noSlashRoot", nil)
	})
	if recv == nil {
		t.Fatal("registering path not beginning with '/' did not panic")
	}
}

func TestRouterNotFound(t *testing.T) {
	router := New()

	testRoutes := []struct {
		route    string
		location string
	}{
		{"noop://path/", "/path"},   // TSR -/
		{"noop://dir", "/dir/"},     // TSR +/
		{"noop://", "/"},            // TSR +/
		{"noop://PATH", "/path"},    // Fixed Case
		{"noop://DIR/", "/dir/"},    // Fixed Case
		{"noop://PATH/", "/path"},   // Fixed Case -/
		{"noop://DIR", "/dir/"},     // Fixed Case +/
		{"noop://../path", "/path"}, // CleanPath
		{"noop://nope", ""},         // NotFound
	}
	for _, tr := range testRoutes {
		if router.ServeURI(tr.route) != ErrNotFound {
			t.Errorf("NotFound handling route %s failed", tr.route)
		}
	}
}

func TestRouterTrailingSlash(t *testing.T) {
	handlerFunc := func(c Context) error {
		if !c.RedirectTrailingSlash {
			t.Errorf("RedirectTrailingSlash should be true, route: %s", c.URI)
		}
		return nil
	}

	router := New()
	router.Handle("noop", "/path", handlerFunc)
	router.Handle("noop", "/dir/", handlerFunc)
	router.Handle("noop", "/", handlerFunc)

	testRoutes := []struct {
		route    string
		location string
	}{
		{"noop://path/", "/path"}, // TSR -/
		{"noop://dir", "/dir/"},   // TSR +/
	}
	for _, tr := range testRoutes {
		if router.ServeURI(tr.route) != nil {
			t.Errorf("NotFound handling route %s failed", tr.route)
		}
	}
}

func TestRouterFixedPath(t *testing.T) {
	handlerFunc := func(c Context) error {
		if !c.RedirectFixedPath {
			t.Errorf("RedirectFixedPath should be true, route: %s", c.URI)
		}
		return nil
	}

	router := New()
	router.Handle("noop", "/path", handlerFunc)
	router.Handle("noop", "/dir/", handlerFunc)
	router.Handle("noop", "/", handlerFunc)

	testRoutes := []struct {
		route    string
		location string
	}{
		{"noop://PATH", "/path"},    // Fixed Case
		{"noop://DIR/", "/dir/"},    // Fixed Case
		{"noop://PATH/", "/path"},   // Fixed Case -/
		{"noop://DIR", "/dir/"},     // Fixed Case +/
		{"noop://../path", "/path"}, // CleanPath
	}
	for _, tr := range testRoutes {
		if router.ServeURI(tr.route) != nil {
			t.Errorf("NotFound handling route %s failed", tr.route)
		}
	}
}

func BenchmarkRouter(b *testing.B) {
	handlerFunc := func(c Context) error {
		if !c.RedirectFixedPath {
			b.Errorf("RedirectFixedPath should be true, route: %s", c.URI)
		}
		return nil
	}

	router := New()
	router.Handle("noop", "/path", handlerFunc)
	router.Handle("noop", "/dir/", handlerFunc)
	router.Handle("noop", "/", handlerFunc)

	testRoutes := []struct {
		route    string
		location string
	}{
		{"noop://PATH", "/path"},    // Fixed Case
		{"noop://DIR/", "/dir/"},    // Fixed Case
		{"noop://PATH/", "/path"},   // Fixed Case -/
		{"noop://DIR", "/dir/"},     // Fixed Case +/
		{"noop://../path", "/path"}, // CleanPath
	}

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			for _, tr := range testRoutes {
				if router.ServeURI(tr.route) != nil {
					b.Errorf("NotFound handling route %s failed", tr.route)
				}
			}
		}
	})
}

func TestRouterLookup(t *testing.T) {
	router := New()

	// try empty router first
	handle, _, tsr := router.Lookup("noop://nope")
	if handle != nil {
		t.Fatalf("Got handle for unregistered pattern: %v", handle)
	}
	if tsr {
		t.Error("Got wrong TSR recommendation!")
	}

	// insert route and try again
	router.Handle("noop", "/user/:name", func(Context) error { return nil })

	handle, params, _ := router.Lookup("noop://user/gopher")
	if handle == nil {
		t.Fatal("Got no handle!")
	} else {
		if err := handle(Context{Params: params, URI: "noop://user/gopher"}); err != nil {
			t.Fatal("Routing failed!")
		}
	}

	wantParams := Params{Param{"name", "gopher"}}
	if !reflect.DeepEqual(params, wantParams) {
		t.Fatalf("Wrong parameter values: want %v, got %v", wantParams, params)
	}

	handle, _, tsr = router.Lookup("noop://user/gopher/")
	if handle != nil {
		t.Fatalf("Got handle for unregistered pattern: %v", handle)
	}
	if !tsr {
		t.Error("Got no TSR recommendation!")
	}

	handle, _, tsr = router.Lookup("noop://nope")
	if handle != nil {
		t.Fatalf("Got handle for unregistered pattern: %v", handle)
	}
	if tsr {
		t.Error("Got wrong TSR recommendation!")
	}
}
