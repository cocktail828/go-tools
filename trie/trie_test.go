package trie

import (
	"testing"
)

func TestPathInsert(t *testing.T) {
	trie := New()
	if trie.root == nil {
		t.Error()
	}

	trie.Add("/", "1")
	if trie.root.Children["/"] == nil {
		t.Error()
	}

	trie.Add("/r", "2")
	if trie.root.Children["/"].Children["r"] == nil {
		t.Error()
	}

	trie.Add("/r/", "3")
	if trie.root.Children["/"].Children["r"].Children["/"] == nil {
		t.Error()
	}
}

func TestTrieCompression(t *testing.T) {
	trie := New()
	trie.Add("/abc", "3")
	trie.Add("/adc", "3")

	// before compression
	if trie.root.Children["/"].Children["a"].Children["b"].Children["c"] == nil {
		t.Error()
	}
	if trie.root.Children["/"].Children["a"].Children["d"].Children["c"] == nil {
		t.Error()
	}

	trie.Compress()

	// after compression
	if trie.root.Children["/abc"] == nil {
		t.Errorf("%+v", trie.root)
	}
	if trie.root.Children["/adc"] == nil {
		t.Errorf("%+v", trie.root)
	}

}

func TestParamInsert(t *testing.T) {
	trie := New()

	trie.Add("/:id/", "")
	if trie.root.Children["/"].ParamChild.Children["/"] == nil {
		t.Error()
	}
	if trie.root.Children["/"].ParamName != "id" {
		t.Error()
	}

	trie.Add("/:id/:property.:format", "")
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
	trie := New()
	trie.Add("/*splat", "")
	if trie.root.Children["/"].SplatChild == nil {
		t.Error()
	}
}

func TestDupeInsert(t *testing.T) {
	trie := New()
	trie.Add("/", "1")
	err := trie.Add("/", "2")
	if err == nil {
		t.Error()
	}
	if trie.root.Children["/"].Anchor != "1" {
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

func TestFind(t *testing.T) {
	trie := New()

	trie.Add("/", "root")
	trie.Add("/r/:id", "resource")
	trie.Add("/r/:id/property", "property")
	trie.Add("/r/:id/property.*format", "property_format")

	trie.Compress()

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
	trie := New()

	trie.Add("/r/1", "resource1")
	trie.Add("/r/2", "resource2")
	trie.Add("/r/:id", "resource_generic")
	trie.Add("/s/*rest", "special_all")
	trie.Add("/s/:param", "special_generic")
	trie.Add("/", "root")

	trie.Compress()

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
