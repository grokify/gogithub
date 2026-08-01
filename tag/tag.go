// Package tag provides GitHub Git tag operations.
package tag

import (
	"context"

	"github.com/grokify/gogithub"
	"github.com/grokify/gogithub/clientv1"
)

// ListTags lists all tags for a repository.
func ListTags(ctx context.Context, client clientv1.Client, owner, repo string) ([]*gogithub.Tag, error) {
	return client.ListTags(ctx, owner, repo)
}

// GetTagSHA returns the commit SHA for a tag.
func GetTagSHA(ctx context.Context, client clientv1.Client, owner, repo, tagName string) (string, error) {
	return client.GetTagSHA(ctx, owner, repo, tagName)
}

// CreateTag creates an annotated tag.
func CreateTag(ctx context.Context, client clientv1.Client, owner, repo, tagName, sha, message string) error {
	return client.CreateTag(ctx, owner, repo, tagName, sha, message)
}

// CreateLightweightTag creates a lightweight tag (just a reference).
func CreateLightweightTag(ctx context.Context, client clientv1.Client, owner, repo, tagName, sha string) error {
	_, err := client.CreateRef(ctx, owner, repo, refTagsPrefix+tagName, sha)
	return err
}

// DeleteTag deletes a tag.
func DeleteTag(ctx context.Context, client clientv1.Client, owner, repo, tagName string) error {
	return client.DeleteRef(ctx, owner, repo, tagsPrefix+tagName)
}

// TagExists checks if a tag exists.
func TagExists(ctx context.Context, client clientv1.Client, owner, repo, tagName string) (bool, error) {
	_, err := client.GetRef(ctx, owner, repo, tagsPrefix+tagName)
	if err != nil {
		// Check if it's a 404 error
		if isNotFoundError(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// GetTagNames returns just the tag names from a repository.
func GetTagNames(ctx context.Context, client clientv1.Client, owner, repo string) ([]string, error) {
	tags, err := client.ListTags(ctx, owner, repo)
	if err != nil {
		return nil, err
	}

	names := make([]string, len(tags))
	for i, t := range tags {
		names[i] = t.Name
	}
	return names, nil
}

// isNotFoundError checks if an error is a 404 not found error.
func isNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	// Check for common 404 error patterns
	errStr := err.Error()
	return contains(errStr, "404") || contains(errStr, "not found")
}

// contains is a simple substring check.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsAt(s, substr))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
