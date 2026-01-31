// Package tag provides GitHub Git tag operations.
package tag

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v82/github"
)

// ListTags lists all tags for a repository.
func ListTags(ctx context.Context, gh *github.Client, owner, repo string) ([]*github.RepositoryTag, error) {
	var allTags []*github.RepositoryTag

	opts := &github.ListOptions{
		PerPage: 100,
	}

	for {
		tags, resp, err := gh.Repositories.ListTags(ctx, owner, repo, opts)
		if err != nil {
			return nil, fmt.Errorf("list tags: %w", err)
		}

		allTags = append(allTags, tags...)

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allTags, nil
}

// GetTagSHA returns the commit SHA for a tag.
func GetTagSHA(ctx context.Context, gh *github.Client, owner, repo, tagName string) (string, error) {
	ref, _, err := gh.Git.GetRef(ctx, owner, repo, "tags/"+tagName)
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
		Ref: "refs/tags/" + tagName,
		SHA: createdTag.GetSHA(),
	}

	_, _, err = gh.Git.CreateRef(ctx, owner, repo, ref)
	if err != nil {
		// Tag ref might already exist if tag was created differently
		if strings.Contains(err.Error(), "already exists") {
			return nil
		}
		return fmt.Errorf("create tag reference: %w", err)
	}

	return nil
}

// CreateLightweightTag creates a lightweight tag (just a reference).
func CreateLightweightTag(ctx context.Context, gh *github.Client, owner, repo, tagName, sha string) error {
	ref := github.CreateRef{
		Ref: "refs/tags/" + tagName,
		SHA: sha,
	}

	_, _, err := gh.Git.CreateRef(ctx, owner, repo, ref)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			return nil
		}
		return fmt.Errorf("create tag reference: %w", err)
	}

	return nil
}

// DeleteTag deletes a tag.
func DeleteTag(ctx context.Context, gh *github.Client, owner, repo, tagName string) error {
	_, err := gh.Git.DeleteRef(ctx, owner, repo, "tags/"+tagName)
	return err
}

// TagExists checks if a tag exists.
func TagExists(ctx context.Context, gh *github.Client, owner, repo, tagName string) (bool, error) {
	_, resp, err := gh.Git.GetRef(ctx, owner, repo, "tags/"+tagName)
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
