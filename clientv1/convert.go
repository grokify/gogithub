package clientv1

import (
	"github.com/google/go-github/v89/github"
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
		if tree := gc.GetTree(); tree != nil {
			commit.Tree = &gogithub.GitObject{SHA: tree.GetSHA()}
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
		Additions: pr.GetAdditions(),
		Deletions: pr.GetDeletions(),
		Commits:   pr.GetCommits(),
		CreatedAt: pr.GetCreatedAt().Time,
		UpdatedAt: pr.GetUpdatedAt().Time,
	}
	for _, l := range pr.Labels {
		result.Labels = append(result.Labels, gogithub.Label{
			ID:          l.GetID(),
			Name:        l.GetName(),
			Description: l.GetDescription(),
			Color:       l.GetColor(),
		})
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

// tagFromGitHub converts a go-github RepositoryTag to our stable Tag type.
func tagFromGitHub(t *github.RepositoryTag) *gogithub.Tag {
	if t == nil {
		return nil
	}
	tag := &gogithub.Tag{
		Name: t.GetName(),
	}
	if c := t.GetCommit(); c != nil {
		tag.SHA = c.GetSHA()
	}
	return tag
}

// tagsFromGitHub converts a slice of go-github RepositoryTags.
func tagsFromGitHub(tags []*github.RepositoryTag) []*gogithub.Tag {
	if tags == nil {
		return nil
	}
	result := make([]*gogithub.Tag, len(tags))
	for i, t := range tags {
		result[i] = tagFromGitHub(t)
	}
	return result
}

// gitCommitToCommit converts a go-github git Commit to our stable Commit type.
func gitCommitToCommit(c *github.Commit) *gogithub.Commit {
	if c == nil {
		return nil
	}
	commit := &gogithub.Commit{
		SHA:     c.GetSHA(),
		Message: c.GetMessage(),
		HTMLURL: c.GetHTMLURL(),
	}
	if author := c.GetAuthor(); author != nil {
		commit.Author = &gogithub.CommitAuthor{
			Name:  author.GetName(),
			Email: author.GetEmail(),
			Date:  author.GetDate().Time,
		}
	}
	if committer := c.GetCommitter(); committer != nil {
		commit.Committer = &gogithub.CommitAuthor{
			Name:  committer.GetName(),
			Email: committer.GetEmail(),
			Date:  committer.GetDate().Time,
		}
	}
	if tree := c.GetTree(); tree != nil {
		commit.Tree = &gogithub.GitObject{SHA: tree.GetSHA()}
	}
	for _, p := range c.Parents {
		commit.Parents = append(commit.Parents, gogithub.CommitParent{
			SHA: p.GetSHA(),
			URL: p.GetURL(),
		})
	}
	return commit
}

// commitFileFromGitHub converts a go-github CommitFile to our stable CommitFile type.
func commitFileFromGitHub(f *github.CommitFile) *gogithub.CommitFile {
	if f == nil {
		return nil
	}
	return &gogithub.CommitFile{
		SHA:         f.GetSHA(),
		Filename:    f.GetFilename(),
		Status:      f.GetStatus(),
		Additions:   f.GetAdditions(),
		Deletions:   f.GetDeletions(),
		Changes:     f.GetChanges(),
		Patch:       f.GetPatch(),
		BlobURL:     f.GetBlobURL(),
		RawURL:      f.GetRawURL(),
		ContentsURL: f.GetContentsURL(),
		Previous:    f.GetPreviousFilename(),
	}
}

// commitFilesFromGitHub converts a slice of go-github CommitFiles.
func commitFilesFromGitHub(files []*github.CommitFile) []*gogithub.CommitFile {
	if files == nil {
		return nil
	}
	result := make([]*gogithub.CommitFile, len(files))
	for i, f := range files {
		result[i] = commitFileFromGitHub(f)
	}
	return result
}

// pullRequestReviewFromGitHub converts a go-github PullRequestReview to our stable type.
func pullRequestReviewFromGitHub(r *github.PullRequestReview) *gogithub.PullRequestReview {
	if r == nil {
		return nil
	}
	review := &gogithub.PullRequestReview{
		ID:       r.GetID(),
		User:     userFromGitHub(r.GetUser()),
		Body:     r.GetBody(),
		State:    r.GetState(),
		HTMLURL:  r.GetHTMLURL(),
		CommitID: r.GetCommitID(),
	}
	if r.SubmittedAt != nil {
		t := r.GetSubmittedAt().Time
		review.SubmittedAt = &t
	}
	return review
}

// pullRequestReviewsFromGitHub converts a slice of go-github PullRequestReviews.
func pullRequestReviewsFromGitHub(reviews []*github.PullRequestReview) []*gogithub.PullRequestReview {
	if reviews == nil {
		return nil
	}
	result := make([]*gogithub.PullRequestReview, len(reviews))
	for i, r := range reviews {
		result[i] = pullRequestReviewFromGitHub(r)
	}
	return result
}

// pullRequestCommentFromGitHub converts a go-github PullRequestComment to our stable type.
func pullRequestCommentFromGitHub(c *github.PullRequestComment) *gogithub.PullRequestComment {
	if c == nil {
		return nil
	}
	return &gogithub.PullRequestComment{
		ID:        c.GetID(),
		User:      userFromGitHub(c.GetUser()),
		Body:      c.GetBody(),
		Path:      c.GetPath(),
		Line:      c.GetLine(),
		Side:      c.GetSide(),
		CommitID:  c.GetCommitID(),
		HTMLURL:   c.GetHTMLURL(),
		CreatedAt: c.GetCreatedAt().Time,
		UpdatedAt: c.GetUpdatedAt().Time,
	}
}

// pullRequestCommentsFromGitHub converts a slice of go-github PullRequestComments.
func pullRequestCommentsFromGitHub(comments []*github.PullRequestComment) []*gogithub.PullRequestComment {
	if comments == nil {
		return nil
	}
	result := make([]*gogithub.PullRequestComment, len(comments))
	for i, c := range comments {
		result[i] = pullRequestCommentFromGitHub(c)
	}
	return result
}

// issueCommentFromGitHub converts a go-github IssueComment to our stable type.
func issueCommentFromGitHub(c *github.IssueComment) *gogithub.IssueComment {
	if c == nil {
		return nil
	}
	return &gogithub.IssueComment{
		ID:        c.GetID(),
		User:      userFromGitHub(c.GetUser()),
		Body:      c.GetBody(),
		HTMLURL:   c.GetHTMLURL(),
		CreatedAt: c.GetCreatedAt().Time,
		UpdatedAt: c.GetUpdatedAt().Time,
	}
}

// checkSuiteFromGitHub converts a go-github CheckSuite to our stable type.
func checkSuiteFromGitHub(cs *github.CheckSuite) *gogithub.CheckSuite {
	if cs == nil {
		return nil
	}
	suite := &gogithub.CheckSuite{
		ID:         cs.GetID(),
		HeadBranch: cs.GetHeadBranch(),
		HeadSHA:    cs.GetHeadSHA(),
		Status:     cs.GetStatus(),
		Conclusion: cs.GetConclusion(),
		URL:        cs.GetURL(),
		CreatedAt:  cs.GetCreatedAt().Time,
		UpdatedAt:  cs.GetUpdatedAt().Time,
	}
	if app := cs.GetApp(); app != nil {
		suite.App = &gogithub.App{
			ID:          app.GetID(),
			Slug:        app.GetSlug(),
			Name:        app.GetName(),
			Description: app.GetDescription(),
			HTMLURL:     app.GetHTMLURL(),
		}
	}
	return suite
}

// checkSuitesFromGitHub converts a slice of go-github CheckSuites.
func checkSuitesFromGitHub(suites []*github.CheckSuite) []*gogithub.CheckSuite {
	if suites == nil {
		return nil
	}
	result := make([]*gogithub.CheckSuite, len(suites))
	for i, cs := range suites {
		result[i] = checkSuiteFromGitHub(cs)
	}
	return result
}

// releaseAssetFromGitHub converts a go-github ReleaseAsset to our stable type.
func releaseAssetFromGitHub(a *github.ReleaseAsset) *gogithub.ReleaseAsset {
	if a == nil {
		return nil
	}
	return &gogithub.ReleaseAsset{
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
	}
}

// releaseAssetsFromGitHub converts a slice of go-github ReleaseAssets.
func releaseAssetsFromGitHub(assets []*github.ReleaseAsset) []*gogithub.ReleaseAsset {
	if assets == nil {
		return nil
	}
	result := make([]*gogithub.ReleaseAsset, len(assets))
	for i, a := range assets {
		result[i] = releaseAssetFromGitHub(a)
	}
	return result
}

// issueFromGitHub converts a go-github Issue to our stable Issue type.
func issueFromGitHub(i *github.Issue) *gogithub.Issue {
	if i == nil {
		return nil
	}
	issue := &gogithub.Issue{
		ID:            i.GetID(),
		Number:        i.GetNumber(),
		State:         i.GetState(),
		Title:         i.GetTitle(),
		Body:          i.GetBody(),
		HTMLURL:       i.GetHTMLURL(),
		RepositoryURL: i.GetRepositoryURL(),
		User:          userFromGitHub(i.GetUser()),
		Comments:      i.GetComments(),
		IsPullRequest: i.IsPullRequest(),
		CreatedAt:     i.GetCreatedAt().Time,
		UpdatedAt:     i.GetUpdatedAt().Time,
	}
	if i.ClosedAt != nil {
		t := i.GetClosedAt().Time
		issue.ClosedAt = &t
	}
	for _, l := range i.Labels {
		issue.Labels = append(issue.Labels, gogithub.Label{
			ID:          l.GetID(),
			Name:        l.GetName(),
			Description: l.GetDescription(),
			Color:       l.GetColor(),
		})
	}
	for _, a := range i.Assignees {
		issue.Assignees = append(issue.Assignees, userFromGitHub(a))
	}
	return issue
}

// issueSearchResultFromGitHub converts a go-github IssuesSearchResult to our stable type.
func issueSearchResultFromGitHub(r *github.IssuesSearchResult) *gogithub.IssueSearchResult {
	if r == nil {
		return nil
	}
	result := &gogithub.IssueSearchResult{
		Total:             r.GetTotal(),
		IncompleteResults: r.GetIncompleteResults(),
	}
	for _, i := range r.Issues {
		result.Items = append(result.Items, issueFromGitHub(i))
	}
	return result
}

// contributorFromGitHub converts a go-github Contributor to our stable User type.
func contributorFromGitHub(c *github.Contributor) *gogithub.User {
	if c == nil {
		return nil
	}
	return &gogithub.User{
		ID:        c.GetID(),
		Login:     c.GetLogin(),
		AvatarURL: c.GetAvatarURL(),
		HTMLURL:   c.GetHTMLURL(),
		Type:      c.GetType(),
	}
}

// contributorStatsFromGitHub converts go-github ContributorStats to our stable type.
func contributorStatsFromGitHub(stats []*github.ContributorStats) []*gogithub.ContributorStats {
	if stats == nil {
		return nil
	}
	result := make([]*gogithub.ContributorStats, len(stats))
	for i, s := range stats {
		cs := &gogithub.ContributorStats{
			Author: contributorFromGitHub(s.GetAuthor()),
			Total:  s.GetTotal(),
		}
		for _, w := range s.Weeks {
			cs.Weeks = append(cs.Weeks, gogithub.WeeklyStats{
				Week:      w.GetWeek().Time,
				Additions: w.GetAdditions(),
				Deletions: w.GetDeletions(),
				Commits:   w.GetCommits(),
			})
		}
		result[i] = cs
	}
	return result
}

// treeNodeFromGitHub converts a go-github TreeEntry to our stable TreeNode type.
func treeNodeFromGitHub(e *github.TreeEntry) *gogithub.TreeNode {
	if e == nil {
		return nil
	}
	return &gogithub.TreeNode{
		Path: e.GetPath(),
		Mode: e.GetMode(),
		Type: e.GetType(),
		SHA:  e.GetSHA(),
		Size: e.GetSize(),
		URL:  e.GetURL(),
	}
}

// treeNodesFromGitHub converts a slice of go-github TreeEntries.
func treeNodesFromGitHub(entries []*github.TreeEntry) []*gogithub.TreeNode {
	if entries == nil {
		return nil
	}
	result := make([]*gogithub.TreeNode, len(entries))
	for i, e := range entries {
		result[i] = treeNodeFromGitHub(e)
	}
	return result
}

// createFileResultFromGitHub converts a go-github RepositoryContentResponse to our stable type.
func createFileResultFromGitHub(r *github.RepositoryContentResponse) *gogithub.CreateFileResult {
	if r == nil {
		return nil
	}
	return &gogithub.CreateFileResult{
		Content: fileContentFromGitHub(r.Content),
		Commit:  repositoryCommitToCommit(&r.Commit),
	}
}

// deleteFileResultFromGitHub converts a go-github RepositoryContentResponse to our stable type.
func deleteFileResultFromGitHub(r *github.RepositoryContentResponse) *gogithub.DeleteFileResult {
	if r == nil {
		return nil
	}
	return &gogithub.DeleteFileResult{
		Commit: repositoryCommitToCommit(&r.Commit),
	}
}

// repositoryCommitToCommit converts a go-github Commit (from content response) to our stable Commit type.
func repositoryCommitToCommit(c *github.Commit) *gogithub.Commit {
	if c == nil {
		return nil
	}
	commit := &gogithub.Commit{
		SHA:     c.GetSHA(),
		Message: c.GetMessage(),
		HTMLURL: c.GetHTMLURL(),
	}
	if author := c.GetAuthor(); author != nil {
		commit.Author = &gogithub.CommitAuthor{
			Name:  author.GetName(),
			Email: author.GetEmail(),
			Date:  author.GetDate().Time,
		}
	}
	if committer := c.GetCommitter(); committer != nil {
		commit.Committer = &gogithub.CommitAuthor{
			Name:  committer.GetName(),
			Email: committer.GetEmail(),
			Date:  committer.GetDate().Time,
		}
	}
	return commit
}

// issuesFromGitHub converts a slice of go-github Issues.
func issuesFromGitHub(issues []*github.Issue) []*gogithub.Issue {
	if issues == nil {
		return nil
	}
	result := make([]*gogithub.Issue, len(issues))
	for i, issue := range issues {
		result[i] = issueFromGitHub(issue)
	}
	return result
}

// codeSearchResultFromGitHub converts a go-github CodeSearchResult to our stable type.
func codeSearchResultFromGitHub(r *github.CodeSearchResult) *gogithub.CodeSearchResult {
	if r == nil {
		return nil
	}
	result := &gogithub.CodeSearchResult{
		Total:             r.GetTotal(),
		IncompleteResults: r.GetIncompleteResults(),
	}
	for _, item := range r.CodeResults {
		result.Items = append(result.Items, codeResultFromGitHub(item))
	}
	return result
}

// codeResultFromGitHub converts a go-github CodeResult to our stable type.
func codeResultFromGitHub(c *github.CodeResult) *gogithub.CodeResult {
	if c == nil {
		return nil
	}
	return &gogithub.CodeResult{
		Name:       c.GetName(),
		Path:       c.GetPath(),
		SHA:        c.GetSHA(),
		HTMLURL:    c.GetHTMLURL(),
		Repository: repositoryFromGitHub(c.GetRepository()),
	}
}

// eventFromGitHub converts a go-github Event to our stable Event type.
func eventFromGitHub(e *github.Event) *gogithub.Event {
	if e == nil {
		return nil
	}
	event := &gogithub.Event{
		ID:        e.GetID(),
		Type:      e.GetType(),
		Public:    e.GetPublic(),
		Actor:     userFromGitHub(e.GetActor()),
		CreatedAt: e.GetCreatedAt().Time,
	}
	if repo := e.GetRepo(); repo != nil {
		event.Repo = &gogithub.EventRepo{
			ID:   repo.GetID(),
			Name: repo.GetName(),
			URL:  repo.GetURL(),
		}
	}
	return event
}

// eventsFromGitHub converts a slice of go-github Events.
func eventsFromGitHub(events []*github.Event) []*gogithub.Event {
	if events == nil {
		return nil
	}
	result := make([]*gogithub.Event, len(events))
	for i, e := range events {
		result[i] = eventFromGitHub(e)
	}
	return result
}
