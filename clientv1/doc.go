// Package clientv1 provides a stable, version-isolated client for the GitHub API.
//
// # Why Use This Package
//
// The google/go-github library increments major versions frequently (v88, v89, v90...),
// requiring import path changes in all consuming code. This package provides:
//
//   - A Client interface that isolates consumers from version churn
//   - Single upgrade point: update gogithub once, all consumers benefit
//
// Types are defined in the root gogithub package for reusability:
//
//   - [gogithub.User] - GitHub user
//   - [gogithub.Repository] - GitHub repository
//   - [gogithub.Reference] - Git reference (branch, tag)
//   - [gogithub.Commit] - Git commit
//   - [gogithub.PullRequest] - Pull request
//   - [gogithub.CheckRun] - CI check run
//   - [gogithub.Release] - GitHub release
//
// # Usage
//
// Instead of importing go-github directly:
//
//	// OLD: Coupled to go-github version
//	import "github.com/google/go-github/v89/github"
//	gh, _ := github.NewClient(nil)
//	user, _, _ := gh.Users.Get(ctx, "")
//
// Use this package:
//
//	// NEW: Version-isolated
//	import (
//	    "github.com/grokify/gogithub"
//	    "github.com/grokify/gogithub/clientv1"
//	)
//
//	client, _ := clientv1.NewClient(ctx, "token")
//	user, _ := client.GetAuthenticatedUser(ctx)  // returns *gogithub.User
//	repos, _ := client.ListUserRepos(ctx, "user") // returns []*gogithub.Repository
//
// # Escape Hatch
//
// For advanced use cases not yet wrapped, use [Client.Raw] to access the
// underlying go-github client. Note: this couples your code to a specific
// go-github version.
//
//	raw := client.Raw().(*github.Client)
//	// Use raw go-github client directly
package clientv1
