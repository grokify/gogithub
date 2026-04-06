package tag

import (
	"testing"
)

func TestTagConstants(t *testing.T) {
	// Test reference prefixes
	if refTagsPrefix != "refs/tags/" {
		t.Errorf("refTagsPrefix = %q, want %q", refTagsPrefix, "refs/tags/")
	}

	if tagsPrefix != "tags/" {
		t.Errorf("tagsPrefix = %q, want %q", tagsPrefix, "tags/")
	}

	// Test error fragment
	if errAlreadyExists != "already exists" {
		t.Errorf("errAlreadyExists = %q, want %q", errAlreadyExists, "already exists")
	}
}

func TestTagsPrefixUsage(t *testing.T) {
	// Verify the prefix can be used to construct valid refs
	tagName := "v1.0.0"

	fullRef := refTagsPrefix + tagName
	expected := "refs/tags/v1.0.0"
	if fullRef != expected {
		t.Errorf("refTagsPrefix + tagName = %q, want %q", fullRef, expected)
	}

	shortRef := tagsPrefix + tagName
	expectedShort := "tags/v1.0.0"
	if shortRef != expectedShort {
		t.Errorf("tagsPrefix + tagName = %q, want %q", shortRef, expectedShort)
	}
}
