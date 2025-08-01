package trie

import (
	"testing"
)

func TestPathInsert(t *testing.T) {
	trie, err := New(map[string]any{
		"/":   "1",
		"/r":  "2",
		"/r/": "3",
	})
	if err != nil {
		t.Error()
	}

	if trie.root == nil {
		t.Error()
	}

	if trie.root.Children["/"] == nil {
		t.Error()
	}

	if trie.root.Children["/"].Children["r"] == nil {
		t.Error()
	}

	if trie.root.Children["/"].Children["r"].Children["/"] == nil {
		t.Error()
	}
}

func TestTrieCompression(t *testing.T) {
	trie, err := New(map[string]any{
		"/abc": "3",
		"/adc": "3",
	})
	if err != nil {
		t.Error()
	}

	// after compression
	if trie.root.Children["/abc"] == nil {
		t.Errorf("%+v", trie.root)
	}
	if trie.root.Children["/adc"] == nil {
		t.Errorf("%+v", trie.root)
	}
}

func TestParamInsert(t *testing.T) {
	trie, err := New(map[string]any{
		"/:id/":                  "",
		"/:id/:property.:format": "",
	})
	if err != nil {
		t.Error()
	}

	if trie.root.Children["/"].ParamChild.Children["/"] == nil {
		t.Error()
	}
	if trie.root.Children["/"].ParamName != "id" {
		t.Error()
	}

	if trie.root.Children["/"].ParamChild.Children["/"].ParamChild.Children["."].ParamChild == nil {
		t.Error()
	}
	if trie.root.Children["/"].ParamName != "id" {
		t.Error()
	}
	if trie.root.Children["/"].ParamChild.Children["/"].ParamName != "property" {
		t.Error()
	}
	if trie.root.Children["/"].ParamChild.Children["/"].ParamChild.Children["."].ParamName != "format" {
		t.Error()
	}
}

func TestSplatInsert(t *testing.T) {
	trie, err := New(map[string]any{"/*splat": ""})
	if err != nil {
		t.Error()
	}

	if trie.root.Children["/"].SplatChild == nil {
		t.Error()
	}
}

func isInMatches(test string, matches []Match) bool {
	for _, match := range matches {
		if match.Anchor.(string) == test {
			return true
		}
	}
	return false
}

func TestParam(t *testing.T) {
	trie, err := New(map[string]any{
		"/r/:id":           "resource",
		"/r/:id/*property": "property",
	})
	if err != nil {
		t.Error()
	}

	matches := trie.Find("/r/1")
	if len(matches) != 1 {
		t.Errorf("expected one anchor, got %d", len(matches))
	}
	if !isInMatches("resource", matches) {
		t.Errorf("expected 'resource', got %+v", matches)
	}
	if matches[0].Params.ByName("id") != "1" {
		t.Error()
	}

	matches = trie.Find("/r/1/property")
	if len(matches) != 1 {
		t.Errorf("expected one anchor, got %d", len(matches))
	}
	if !isInMatches("property", matches) {
		t.Error("expected 'property'")
	}
	if matches[0].Params.ByName("id") != "1" {
		t.Error()
	}
	if matches[0].Params.ByName("property") != "property" {
		t.Error()
	}

	matches = trie.Find("/r/1/property.json")
	if len(matches) != 1 {
		t.Errorf("expected one anchor, got %d", len(matches))
	}
	if !isInMatches("property", matches) {
		t.Error("expected 'property'")
	}
	if matches[0].Params.ByName("id") != "1" {
		t.Error()
	}
	if matches[0].Params.ByName("property") != "property.json" {
		t.Error()
	}
}

func TestFind(t *testing.T) {
	trie, err := New(map[string]any{
		"/":                       "root",
		"/r/:id":                  "resource",
		"/r/:id/property":         "property",
		"/r/:id/property.*format": "property_format",
	})
	if err != nil {
		t.Error()
	}

	matches := trie.Find("/")
	if len(matches) != 1 {
		t.Errorf("expected one anchor, got %d", len(matches))
	}
	if !isInMatches("root", matches) {
		t.Error("expected 'root'")
	}

	matches = trie.Find("/notfound")
	if len(matches) != 0 {
		t.Errorf("expected zero anchor, got %d", len(matches))
	}

	matches = trie.Find("/r/1")
	if len(matches) != 1 {
		t.Errorf("expected one anchor, got %d", len(matches))
	}
	if !isInMatches("resource", matches) {
		t.Errorf("expected 'resource', got %+v", matches)
	}
	if matches[0].Params.ByName("id") != "1" {
		t.Error()
	}

	matches = trie.Find("/r/1/property")
	if len(matches) != 1 {
		t.Errorf("expected one anchor, got %d", len(matches))
	}
	if !isInMatches("property", matches) {
		t.Error("expected 'property'")
	}
	if matches[0].Params.ByName("id") != "1" {
		t.Error()
	}

	matches = trie.Find("/r/1/property.json")
	if len(matches) != 1 {
		t.Errorf("expected one anchor, got %d", len(matches))
	}
	if !isInMatches("property_format", matches) {
		t.Error("expected 'property_format'")
	}
	if matches[0].Params.ByName("id") != "1" {
		t.Error()
	}
	if matches[0].Params.ByName("format") != "json" {
		t.Error()
	}
}

func TestFindMultipleMatches(t *testing.T) {
	trie, err := New(map[string]any{
		"/r/1":      "resource1",
		"/r/2":      "resource2",
		"/r/:id":    "resource_generic",
		"/s/*rest":  "special_all",
		"/s/:param": "special_generic",
		"/":         "root",
	})
	if err != nil {
		t.Error()
	}

	matches := trie.Find("/r/1")
	if len(matches) != 2 {
		t.Errorf("expected two matches, got %d", len(matches))
	}
	if !isInMatches("resource_generic", matches) {
		t.Error()
	}
	if !isInMatches("resource1", matches) {
		t.Error()
	}

	matches = trie.Find("/s/1")
	if len(matches) != 2 {
		t.Errorf("expected two matches, got %d", len(matches))
	}
	if !isInMatches("special_all", matches) {
		t.Error()
	}
	if !isInMatches("special_generic", matches) {
		t.Error()
	}
}
