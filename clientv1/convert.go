package clientv1

import (
	"github.com/google/go-github/v88/github"
	"github.com/grokify/gogithub"
)

// userFromGitHub converts a go-github User to our stable User type.
func userFromGitHub(u *github.User) *gogithub.User {
	if u == nil {
		return nil
	}
	return &gogithub.User{
		ID:        u.GetID(),
		Login:     u.GetLogin(),
		Name:      u.GetName(),
		Email:     u.GetEmail(),
		AvatarURL: u.GetAvatarURL(),
		HTMLURL:   u.GetHTMLURL(),
		Type:      u.GetType(),
		Bio:       u.GetBio(),
		Company:   u.GetCompany(),
		Location:  u.GetLocation(),
		Blog:      u.GetBlog(),
		Followers: u.GetFollowers(),
		Following: u.GetFollowing(),
		CreatedAt: u.GetCreatedAt().Time,
		UpdatedAt: u.GetUpdatedAt().Time,
	}
}

// repositoryFromGitHub converts a go-github Repository to our stable Repository type.
func repositoryFromGitHub(r *github.Repository) *gogithub.Repository {
	if r == nil {
		return nil
	}
	return &gogithub.Repository{
		ID:              r.GetID(),
		Owner:           userFromGitHub(r.GetOwner()),
		Name:            r.GetName(),
		FullName:        r.GetFullName(),
		Description:     r.GetDescription(),
		HTMLURL:         r.GetHTMLURL(),
		CloneURL:        r.GetCloneURL(),
		SSHURL:          r.GetSSHURL(),
		DefaultBranch:   r.GetDefaultBranch(),
		Private:         r.GetPrivate(),
		Fork:            r.GetFork(),
		Archived:        r.GetArchived(),
		Disabled:        r.GetDisabled(),
		Language:        r.GetLanguage(),
		ForksCount:      r.GetForksCount(),
		StargazersCount: r.GetStargazersCount(),
		WatchersCount:   r.GetWatchersCount(),
		OpenIssuesCount: r.GetOpenIssuesCount(),
		Size:            r.GetSize(),
		CreatedAt:       r.GetCreatedAt().Time,
		UpdatedAt:       r.GetUpdatedAt().Time,
		PushedAt:        r.GetPushedAt().Time,
	}
}

// repositoriesFromGitHub converts a slice of go-github Repositories.
func repositoriesFromGitHub(repos []*github.Repository) []*gogithub.Repository {
	if repos == nil {
		return nil
	}
	result := make([]*gogithub.Repository, len(repos))
	for i, r := range repos {
		result[i] = repositoryFromGitHub(r)
	}
	return result
}

// referenceFromGitHub converts a go-github Reference to our stable Reference type.
func referenceFromGitHub(r *github.Reference) *gogithub.Reference {
	if r == nil {
		return nil
	}
	ref := &gogithub.Reference{
		Ref: r.GetRef(),
		URL: r.GetURL(),
	}
	if obj := r.GetObject(); obj != nil {
		ref.SHA = obj.GetSHA()
		ref.Object = &gogithub.GitObject{
			Type: obj.GetType(),
			SHA:  obj.GetSHA(),
			URL:  obj.GetURL(),
		}
	}
	return ref
}

// commitFromGitHub converts a go-github RepositoryCommit to our stable Commit type.
func commitFromGitHub(c *github.RepositoryCommit) *gogithub.Commit {
	if c == nil {
		return nil
	}
	commit := &gogithub.Commit{
		SHA:     c.GetSHA(),
		HTMLURL: c.GetHTMLURL(),
	}
	if gc := c.GetCommit(); gc != nil {
		commit.Message = gc.GetMessage()
		if author := gc.GetAuthor(); author != nil {
			commit.Author = &gogithub.CommitAuthor{
				Name:  author.GetName(),
				Email: author.GetEmail(),
				Date:  author.GetDate().Time,
			}
		}
		if committer := gc.GetCommitter(); committer != nil {
			commit.Committer = &gogithub.CommitAuthor{
				Name:  committer.GetName(),
				Email: committer.GetEmail(),
				Date:  committer.GetDate().Time,
			}
		}
	}
	for _, p := range c.Parents {
		commit.Parents = append(commit.Parents, gogithub.CommitParent{
			SHA: p.GetSHA(),
			URL: p.GetURL(),
		})
	}
	return commit
}

// commitsFromGitHub converts a slice of go-github RepositoryCommits.
func commitsFromGitHub(commits []*github.RepositoryCommit) []*gogithub.Commit {
	if commits == nil {
		return nil
	}
	result := make([]*gogithub.Commit, len(commits))
	for i, c := range commits {
		result[i] = commitFromGitHub(c)
	}
	return result
}

// branchFromGitHub converts a go-github Branch to our stable Branch type.
func branchFromGitHub(b *github.Branch) *gogithub.Branch {
	if b == nil {
		return nil
	}
	branch := &gogithub.Branch{
		Name:      b.GetName(),
		Protected: b.GetProtected(),
	}
	if c := b.GetCommit(); c != nil {
		branch.Commit = commitFromGitHub(c)
	}
	return branch
}

// branchesFromGitHub converts a slice of go-github Branches.
func branchesFromGitHub(branches []*github.Branch) []*gogithub.Branch {
	if branches == nil {
		return nil
	}
	result := make([]*gogithub.Branch, len(branches))
	for i, b := range branches {
		result[i] = branchFromGitHub(b)
	}
	return result
}

// pullRequestFromGitHub converts a go-github PullRequest to our stable PullRequest type.
func pullRequestFromGitHub(pr *github.PullRequest) *gogithub.PullRequest {
	if pr == nil {
		return nil
	}
	result := &gogithub.PullRequest{
		ID:        pr.GetID(),
		Number:    pr.GetNumber(),
		State:     pr.GetState(),
		Title:     pr.GetTitle(),
		Body:      pr.GetBody(),
		HTMLURL:   pr.GetHTMLURL(),
		User:      userFromGitHub(pr.GetUser()),
		Merged:    pr.GetMerged(),
		Mergeable: pr.Mergeable,
		Draft:     pr.GetDraft(),
		CreatedAt: pr.GetCreatedAt().Time,
		UpdatedAt: pr.GetUpdatedAt().Time,
	}
	if pr.ClosedAt != nil {
		t := pr.GetClosedAt().Time
		result.ClosedAt = &t
	}
	if pr.MergedAt != nil {
		t := pr.GetMergedAt().Time
		result.MergedAt = &t
	}
	if head := pr.GetHead(); head != nil {
		result.Head = &gogithub.PullRequestBranch{
			Label: head.GetLabel(),
			Ref:   head.GetRef(),
			SHA:   head.GetSHA(),
			User:  userFromGitHub(head.GetUser()),
			Repo:  repositoryFromGitHub(head.GetRepo()),
		}
	}
	if base := pr.GetBase(); base != nil {
		result.Base = &gogithub.PullRequestBranch{
			Label: base.GetLabel(),
			Ref:   base.GetRef(),
			SHA:   base.GetSHA(),
			User:  userFromGitHub(base.GetUser()),
			Repo:  repositoryFromGitHub(base.GetRepo()),
		}
	}
	return result
}

// pullRequestsFromGitHub converts a slice of go-github PullRequests.
func pullRequestsFromGitHub(prs []*github.PullRequest) []*gogithub.PullRequest {
	if prs == nil {
		return nil
	}
	result := make([]*gogithub.PullRequest, len(prs))
	for i, pr := range prs {
		result[i] = pullRequestFromGitHub(pr)
	}
	return result
}

// checkRunFromGitHub converts a go-github CheckRun to our stable CheckRun type.
func checkRunFromGitHub(cr *github.CheckRun) *gogithub.CheckRun {
	if cr == nil {
		return nil
	}
	result := &gogithub.CheckRun{
		ID:         cr.GetID(),
		HeadSHA:    cr.GetHeadSHA(),
		Status:     cr.GetStatus(),
		Conclusion: cr.GetConclusion(),
		Name:       cr.GetName(),
		HTMLURL:    cr.GetHTMLURL(),
	}
	if cr.StartedAt != nil {
		t := cr.GetStartedAt().Time
		result.StartedAt = &t
	}
	if cr.CompletedAt != nil {
		t := cr.GetCompletedAt().Time
		result.CompletedAt = &t
	}
	return result
}

// checkRunsFromGitHub converts a slice of go-github CheckRuns.
func checkRunsFromGitHub(runs []*github.CheckRun) []*gogithub.CheckRun {
	if runs == nil {
		return nil
	}
	result := make([]*gogithub.CheckRun, len(runs))
	for i, cr := range runs {
		result[i] = checkRunFromGitHub(cr)
	}
	return result
}

// releaseFromGitHub converts a go-github RepositoryRelease to our stable Release type.
func releaseFromGitHub(r *github.RepositoryRelease) *gogithub.Release {
	if r == nil {
		return nil
	}
	result := &gogithub.Release{
		ID:              r.GetID(),
		TagName:         r.GetTagName(),
		TargetCommitish: r.GetTargetCommitish(),
		Name:            r.GetName(),
		Body:            r.GetBody(),
		Draft:           r.GetDraft(),
		Prerelease:      r.GetPrerelease(),
		HTMLURL:         r.GetHTMLURL(),
		TarballURL:      r.GetTarballURL(),
		ZipballURL:      r.GetZipballURL(),
		CreatedAt:       r.GetCreatedAt().Time,
		Author:          userFromGitHub(r.GetAuthor()),
	}
	if r.PublishedAt != nil {
		t := r.GetPublishedAt().Time
		result.PublishedAt = &t
	}
	for _, a := range r.Assets {
		result.Assets = append(result.Assets, gogithub.ReleaseAsset{
			ID:                 a.GetID(),
			Name:               a.GetName(),
			Label:              a.GetLabel(),
			State:              a.GetState(),
			ContentType:        a.GetContentType(),
			Size:               a.GetSize(),
			DownloadCount:      a.GetDownloadCount(),
			BrowserDownloadURL: a.GetBrowserDownloadURL(),
			CreatedAt:          a.GetCreatedAt().Time,
			UpdatedAt:          a.GetUpdatedAt().Time,
		})
	}
	return result
}

// releasesFromGitHub converts a slice of go-github RepositoryReleases.
func releasesFromGitHub(releases []*github.RepositoryRelease) []*gogithub.Release {
	if releases == nil {
		return nil
	}
	result := make([]*gogithub.Release, len(releases))
	for i, r := range releases {
		result[i] = releaseFromGitHub(r)
	}
	return result
}

// fileContentFromGitHub converts a go-github RepositoryContent to our stable FileContent type.
func fileContentFromGitHub(c *github.RepositoryContent) *gogithub.FileContent {
	if c == nil {
		return nil
	}
	fc := &gogithub.FileContent{
		Path:        c.GetPath(),
		Name:        c.GetName(),
		SHA:         c.GetSHA(),
		Size:        c.GetSize(),
		Type:        c.GetType(),
		DownloadURL: c.GetDownloadURL(),
	}
	// Content is only present for files, not directories
	if c.Content != nil {
		if decoded, err := c.GetContent(); err == nil {
			fc.Content = []byte(decoded)
		}
	}
	return fc
}

// fileContentsFromGitHub converts a slice of go-github RepositoryContents.
func fileContentsFromGitHub(contents []*github.RepositoryContent) []*gogithub.FileContent {
	if contents == nil {
		return nil
	}
	result := make([]*gogithub.FileContent, len(contents))
	for i, c := range contents {
		result[i] = fileContentFromGitHub(c)
	}
	return result
}
