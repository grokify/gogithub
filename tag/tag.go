// Package tag provides GitHub Git tag operations.
package tag

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v84/github"
)

// ListTags lists all tags for a repository.
// Uses go-github's built-in iterator for automatic pagination handling.
func ListTags(ctx context.Context, gh *github.Client, owner, repo string) ([]*github.RepositoryTag, error) {
	var allTags []*github.RepositoryTag

	for tag, err := range gh.Repositories.ListTagsIter(ctx, owner, repo, nil) {
		if err != nil {
			return nil, fmt.Errorf("list tags: %w", err)
		}
		allTags = append(allTags, tag)
	}

	return allTags, nil
}

// GetTagSHA returns the commit SHA for a tag.
func GetTagSHA(ctx context.Context, gh *github.Client, owner, repo, tagName string) (string, error) {
	ref, _, err := gh.Git.GetRef(ctx, owner, repo, tagsPrefix+tagName)
	if err != nil {
		return "", fmt.Errorf("get tag ref: %w", err)
	}
	return ref.GetObject().GetSHA(), nil
}

// CreateTag creates an annotated tag.
func CreateTag(ctx context.Context, gh *github.Client, owner, repo, tagName, sha, message string) error {
	// Create annotated tag object
	tag := github.CreateTag{
		Tag:     tagName,
		Message: message,
		Object:  sha,
		Type:    "commit",
	}

	createdTag, _, err := gh.Git.CreateTag(ctx, owner, repo, tag)
	if err != nil {
		return fmt.Errorf("create tag object: %w", err)
	}

	// Create reference to tag
	ref := github.CreateRef{
		Ref: refTagsPrefix + tagName,
		SHA: createdTag.GetSHA(),
	}

	_, _, err = gh.Git.CreateRef(ctx, owner, repo, ref)
	if err != nil {
		// Tag ref might already exist if tag was created differently
		if strings.Contains(err.Error(), errAlreadyExists) {
			return nil
		}
		return fmt.Errorf("create tag reference: %w", err)
	}

	return nil
}

// CreateLightweightTag creates a lightweight tag (just a reference).
func CreateLightweightTag(ctx context.Context, gh *github.Client, owner, repo, tagName, sha string) error {
	ref := github.CreateRef{
		Ref: refTagsPrefix + tagName,
		SHA: sha,
	}

	_, _, err := gh.Git.CreateRef(ctx, owner, repo, ref)
	if err != nil {
		if strings.Contains(err.Error(), errAlreadyExists) {
			return nil
		}
		return fmt.Errorf("create tag reference: %w", err)
	}

	return nil
}

// DeleteTag deletes a tag.
func DeleteTag(ctx context.Context, gh *github.Client, owner, repo, tagName string) error {
	_, err := gh.Git.DeleteRef(ctx, owner, repo, tagsPrefix+tagName)
	return err
}

// TagExists checks if a tag exists.
func TagExists(ctx context.Context, gh *github.Client, owner, repo, tagName string) (bool, error) {
	_, resp, err := gh.Git.GetRef(ctx, owner, repo, tagsPrefix+tagName)
	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// GetTagNames returns just the tag names from a repository.
func GetTagNames(ctx context.Context, gh *github.Client, owner, repo string) ([]string, error) {
	tags, err := ListTags(ctx, gh, owner, repo)
	if err != nil {
		return nil, err
	}

	names := make([]string, len(tags))
	for i, t := range tags {
		names[i] = t.GetName()
	}
	return names, nil
}
