package tree_sitter_t2_test

import (
	"testing"

	tree_sitter "github.com/smacker/go-tree-sitter"
	tree_sitter_t2 "github.com/tree-sitter/tree-sitter-t2"
)

func TestCanLoadGrammar(t *testing.T) {
	language := tree_sitter.NewLanguage(tree_sitter_t2.Language())
	if language == nil {
		t.Errorf("Error loading GSX grammar")
	}
}
