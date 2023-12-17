package gogithub

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/v56/github"
	"github.com/grokify/gocharts/v2/data/histogram"
	"github.com/grokify/gocharts/v2/data/table"
	"github.com/grokify/mogo/pointer"
)

var (
	ErrIssueIsNotSet                 = errors.New("issue is not set")
	ErrUserIsNotSet                  = errors.New("user is not set")
	ErrUserLoginIsNotSet             = errors.New("user login is not set")
	ErrIssueCreatedAtIsNotSet        = errors.New("issue created at is not set")
	ErrIssueCreatedAtGetTimeIsNotSet = errors.New("issue created at gettime is not set")
)

type Issues []*github.Issue

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
		k2 := strings.Replace(k, BaseURLRepoAPI, BaseURLRepoHTML, 1)
		hURLs[k2] = v
	}
	return hURLs
}

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

func (iss Issues) TableRepos(name string, htmlURLs bool) *table.Table {
	h := histogram.NewHistogram(name)
	h.MapAdd(iss.RepositoryIssueCounts(htmlURLs))
	tblRepo := h.Table("Repo", "Issue Count")
	tblRepo.Name = h.Name
	tblRepo.FormatMap[0] = table.FormatURL
	for i, row := range tblRepo.Rows {
		if len(row) == 0 {
			continue
		} else if repoURL := strings.TrimSpace(row[0]); len(repoURL) == 0 {
			continue
		} else {
			row[0] = "[" + repoURL + "](" + repoURL + ")"
			tblRepo.Rows[i] = row
		}
	}
	return tblRepo
}

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

type Issue struct {
	*github.Issue
}

func (is *Issue) AuthorUserID() (int64, error) {
	if is.Issue == nil {
		return -1, ErrIssueIsNotSet
	} else if is.Issue.User == nil {
		return -1, ErrUserIsNotSet
	} else if is.Issue.User.ID == nil {
		return -1, ErrUserLoginIsNotSet
	} else {
		return pointer.Dereference(is.Issue.User.ID), nil
	}
}

func (is *Issue) AuthorUsername() (string, error) {
	if is.Issue == nil {
		return "", ErrIssueIsNotSet
	} else if is.Issue.User == nil {
		return "", ErrUserIsNotSet
	} else if is.Issue.User.Login == nil {
		return "", ErrUserLoginIsNotSet
	} else {
		return pointer.Dereference(is.Issue.User.Login), nil
	}
}

func (is *Issue) CreatedTime() (time.Time, error) {
	if is.Issue == nil {
		return time.Time{}, ErrIssueIsNotSet
	} else if is.Issue.CreatedAt == nil {
		return time.Time{}, ErrIssueCreatedAtIsNotSet
	} else if dt := is.Issue.CreatedAt.GetTime(); dt == nil || dt.IsZero() {
		return time.Time{}, ErrIssueCreatedAtGetTimeIsNotSet
	} else {
		return *dt, nil
	}
}

func (is *Issue) CreatedAge() (time.Duration, error) {
	if dt, err := is.CreatedTime(); err != nil {
		return -1, err
	} else {
		return time.Since(dt), nil
	}
}

func (is *Issue) MustAuthorUsername() string {
	if username, err := is.AuthorUsername(); err != nil {
		return ""
	} else {
		return username
	}
}

func (is *Issue) MustAuthorUserID() int64 {
	if userID, err := is.AuthorUserID(); err != nil {
		return -1
	} else {
		return userID
	}
}
