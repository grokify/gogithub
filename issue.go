package gogithub

import (
	"errors"
	"strconv"
	"time"

	"github.com/google/go-github/v56/github"
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

func (iss Issues) Table() (*table.Table, error) {
	tbl := table.NewTable("")
	tbl.Columns = []string{
		"Author",
		"Title",
		"HTML URL",
		"State",
		"Created",
		"Age (Days)",
	}
	tbl.FormatMap = map[int]string{
		2: table.FormatURL,
		4: table.FormatTime,
		5: table.FormatInt,
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
	} else if dt := is.Issue.CreatedAt.GetTime(); dt == nil {
		return time.Time{}, ErrIssueCreatedAtGetTimeIsNotSet
	} else {
		return *dt, nil
	}
}

func (is *Issue) CreatedAge() (time.Duration, error) {
	if dt, err := is.CreatedTime(); err != nil {
		return -1, err
	} else {
		return time.Now().Sub(dt), nil
	}
}

func (is *Issue) MustAuthorUsername() string {
	if username, err := is.AuthorUsername(); err != nil {
		return ""
	} else {
		return username
	}
}

func DereferenceSlice[S ~[]*E, E comparable](s S) []E {
	deref := []E{}
	for _, e := range s {
		deref = append(deref, *e)
	}
	return deref
}
