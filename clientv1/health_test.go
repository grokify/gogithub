package clientv1

import (
	"context"
	"net/http"
	"net/url"
	"testing"
)

func TestCountPullRequests(t *testing.T) {
	tests := []struct {
		name  string
		state string
		body  string
		link  string
		want  int
	}{
		{
			name:  "multiple pages",
			state: "open",
			body:  `[{"number":7}]`,
			link:  `<https://api.github.com/repos/o/r/pulls?per_page=1&page=2>; rel="next", <https://api.github.com/repos/o/r/pulls?per_page=1&page=42>; rel="last"`,
			want:  42,
		},
		{
			name: "single item, no link header",
			body: `[{"number":1}]`,
			want: 1,
		},
		{
			name: "none",
			body: `[]`,
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var query url.Values
			c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/repos/o/r/pulls" {
					t.Errorf("path = %q, want /repos/o/r/pulls", r.URL.Path)
				}
				query = r.URL.Query()
				if tt.link != "" {
					w.Header().Set("Link", tt.link)
				}
				writeJSON(t, w, tt.body)
			}))

			got, err := c.CountPullRequests(context.Background(), "o", "r", tt.state)
			if err != nil {
				t.Fatalf("CountPullRequests() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("CountPullRequests() = %d, want %d", got, tt.want)
			}
			if query.Get("per_page") != "1" {
				t.Errorf("per_page = %q, want 1", query.Get("per_page"))
			}
			wantState := tt.state
			if wantState == "" {
				wantState = "open"
			}
			if query.Get("state") != wantState {
				t.Errorf("state = %q, want %q", query.Get("state"), wantState)
			}
		})
	}
}

func TestCountPullRequestsError(t *testing.T) {
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		writeJSON(t, w, `{"message":"Not Found"}`)
	}))
	if _, err := c.CountPullRequests(context.Background(), "o", "missing", ""); err == nil {
		t.Error("CountPullRequests() error = nil, want error")
	}
}

func TestListRepositoryWorkflowRuns(t *testing.T) {
	var query url.Values
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/o/r/actions/runs" {
			t.Errorf("path = %q, want /repos/o/r/actions/runs", r.URL.Path)
		}
		query = r.URL.Query()
		writeJSON(t, w, `{"total_count":2,"workflow_runs":[
			{"id":11,"name":"CI","workflow_id":1,"run_number":5,"event":"push","status":"completed","conclusion":"success","head_branch":"main","head_sha":"abc"},
			{"id":10,"name":"Lint","workflow_id":2,"run_number":3,"event":"push","status":"in_progress","head_branch":"main","head_sha":"abc"}
		]}`)
	}))

	runs, err := c.ListRepositoryWorkflowRuns(context.Background(), "o", "r", &ListWorkflowRunsOptions{
		Branch:  "main",
		PerPage: 100,
	})
	if err != nil {
		t.Fatalf("ListRepositoryWorkflowRuns() error = %v", err)
	}
	if len(runs) != 2 {
		t.Fatalf("len(runs) = %d, want 2", len(runs))
	}
	if runs[0].WorkflowID != 1 || runs[0].Conclusion != "success" {
		t.Errorf("runs[0] = %+v, want workflow 1 success", runs[0])
	}
	if runs[1].WorkflowID != 2 || runs[1].Status != "in_progress" {
		t.Errorf("runs[1] = %+v, want workflow 2 in_progress", runs[1])
	}
	if query.Get("branch") != "main" {
		t.Errorf("branch = %q, want main", query.Get("branch"))
	}
	if query.Get("per_page") != "100" {
		t.Errorf("per_page = %q, want 100", query.Get("per_page"))
	}
}

func TestListRepositoryWorkflowRunsNilOptions(t *testing.T) {
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "" {
			t.Errorf("query = %q, want none", r.URL.RawQuery)
		}
		writeJSON(t, w, `{"total_count":0,"workflow_runs":[]}`)
	}))
	runs, err := c.ListRepositoryWorkflowRuns(context.Background(), "o", "r", nil)
	if err != nil {
		t.Fatalf("ListRepositoryWorkflowRuns() error = %v", err)
	}
	if len(runs) != 0 {
		t.Errorf("len(runs) = %d, want 0", len(runs))
	}
}
