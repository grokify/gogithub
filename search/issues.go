package search

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/v84/github"
	"github.com/grokify/gocharts/v2/data/histogram"
	"github.com/grokify/gocharts/v2/data/table"
	"github.com/grokify/gogithub"
	"github.com/grokify/mogo/pointer"
)

// Issue-related errors.
var (
	ErrIssueIsNotSet                 = errors.New("issue is not set")
	ErrUserIsNotSet                  = errors.New("user is not set")
	ErrUserLoginIsNotSet             = errors.New("user login is not set")
	ErrIssueCreatedAtIsNotSet        = errors.New("issue created at is not set")
	ErrIssueCreatedAtGetTimeIsNotSet = errors.New("issue created at gettime is not set")
)

// Issues is a slice of GitHub issues.
type Issues []*github.Issue

// RepositoryIssueCounts returns a map of repository URLs to issue counts.
func (iss Issues) RepositoryIssueCounts(htmlURLs bool) map[string]int {
	out := map[string]int{}
	for _, is := range iss {
		out[strings.TrimSpace(pointer.Dereference(is.RepositoryURL))]++
	}
	if !htmlURLs {
		return out
	}
	hURLs := map[string]int{}
	for k, v := range out {
		k2 := strings.Replace(k, gogithub.BaseURLRepoAPI, gogithub.BaseURLRepoHTML, 1)
		hURLs[k2] = v
	}
	return hURLs
}

// TableSet returns a table set with issues and repositories.
func (iss Issues) TableSet() (*table.TableSet, error) {
	ts := table.NewTableSet("Github Issues")
	tblIss, err := iss.Table("Issues")
	if err != nil {
		return nil, err
	}
	ts.TableMap[tblIss.Name] = tblIss

	tblRepo := iss.TableRepos("Repositories", true)
	ts.TableMap[tblRepo.Name] = tblRepo
	ts.Order = []string{tblIss.Name, tblRepo.Name}
	return ts, nil
}

// TableRepos creates a table of repositories with issue counts.
func (iss Issues) TableRepos(name string, htmlURLs bool) *table.Table {
	h := histogram.NewHistogram(name)
	h.MapAdd(iss.RepositoryIssueCounts(htmlURLs))
	tblRepo := h.Table("Repo", "Issue Count")
	tblRepo.Name = h.Name
	tblRepo.FormatMap[0] = table.FormatURL
	for i, row := range tblRepo.Rows {
		if len(row) == 0 {
			continue
		}
		if repoURL := strings.TrimSpace(row[0]); len(repoURL) == 0 {
			continue
		} else {
			row[0] = "[" + repoURL + "](" + repoURL + ")"
			tblRepo.Rows[i] = row
		}
	}
	return tblRepo
}

// Table creates a table of issues.
func (iss Issues) Table(name string) (*table.Table, error) {
	tbl := table.NewTable(name)
	tbl.Columns = []string{
		"Author",
		"Author User ID",
		"Title",
		"HTML URL",
		"State",
		"Created",
		"Age (Days)",
	}
	tbl.FormatMap = map[int]string{
		1: table.FormatInt,
		3: table.FormatURL,
		5: table.FormatTime,
		6: table.FormatInt,
	}
	for _, is := range iss {
		im := Issue{Issue: is}
		created, err := im.CreatedTime()
		if err != nil {
			return nil, err
		}
		createdDur, err := im.CreatedAge()
		if err != nil {
			return nil, err
		}
		row := []string{
			im.MustAuthorUsername(),
			strconv.Itoa(int(im.MustAuthorUserID())),
			pointer.Dereference(is.Title),
			pointer.Dereference(is.HTMLURL),
			pointer.Dereference(is.State),
			created.Format(time.RFC3339),
			strconv.Itoa(int(createdDur.Hours() / 24)),
		}
		tbl.Rows = append(tbl.Rows, row)
	}
	return &tbl, nil
}

// Issue wraps a GitHub issue with helper methods.
type Issue struct {
	*github.Issue
}

// AuthorUserID returns the author's user ID.
func (is *Issue) AuthorUserID() (int64, error) {
	if is.Issue == nil {
		return -1, ErrIssueIsNotSet
	}
	if is.Issue.User == nil {
		return -1, ErrUserIsNotSet
	}
	if is.Issue.User.ID == nil {
		return -1, ErrUserLoginIsNotSet
	}
	return pointer.Dereference(is.Issue.User.ID), nil
}

// AuthorUsername returns the author's username.
func (is *Issue) AuthorUsername() (string, error) {
	if is.Issue == nil {
		return "", ErrIssueIsNotSet
	}
	if is.Issue.User == nil {
		return "", ErrUserIsNotSet
	}
	if is.Issue.User.Login == nil {
		return "", ErrUserLoginIsNotSet
	}
	return pointer.Dereference(is.Issue.User.Login), nil
}

// CreatedTime returns the issue creation time.
func (is *Issue) CreatedTime() (time.Time, error) {
	if is.Issue == nil {
		return time.Time{}, ErrIssueIsNotSet
	}
	if is.Issue.CreatedAt == nil {
		return time.Time{}, ErrIssueCreatedAtIsNotSet
	}
	if dt := is.Issue.CreatedAt.GetTime(); dt == nil || dt.IsZero() {
		return time.Time{}, ErrIssueCreatedAtGetTimeIsNotSet
	} else {
		return *dt, nil
	}
}

// CreatedAge returns the duration since the issue was created.
func (is *Issue) CreatedAge() (time.Duration, error) {
	dt, err := is.CreatedTime()
	if err != nil {
		return -1, err
	}
	return time.Since(dt), nil
}

// MustAuthorUsername returns the author username or empty string on error.
func (is *Issue) MustAuthorUsername() string {
	username, err := is.AuthorUsername()
	if err != nil {
		return ""
	}
	return username
}

// MustAuthorUserID returns the author user ID or -1 on error.
func (is *Issue) MustAuthorUserID() int64 {
	userID, err := is.AuthorUserID()
	if err != nil {
		return -1
	}
	return userID
}
