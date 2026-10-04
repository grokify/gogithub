package etagcache

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

// etagServer serves a body per path with an ETag, answering 304 to a
// matching If-None-Match. Bodies can be changed to simulate updates.
type etagServer struct {
	mu       sync.Mutex
	bodies   map[string]string
	versions map[string]int
	hits     atomic.Int32
	full     atomic.Int32
}

func newETagServer() *etagServer {
	return &etagServer{bodies: map[string]string{}, versions: map[string]int{}}
}

func (s *etagServer) set(path, body string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.bodies[path] = body
	s.versions[path]++
}

func (s *etagServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.hits.Add(1)
	s.mu.Lock()
	body, ok := s.bodies[r.URL.Path]
	etag := fmt.Sprintf(`"v%d-%s"`, s.versions[r.URL.Path], r.Header.Get("Authorization"))
	s.mu.Unlock()
	w.Header().Set("X-RateLimit-Remaining", fmt.Sprint(5000-s.hits.Load()))
	if !ok {
		http.Error(w, `{"message":"Not Found"}`, http.StatusNotFound)
		return
	}
	if r.URL.Query().Get("noetag") == "" {
		w.Header().Set("ETag", etag)
	}
	if r.Header.Get("If-None-Match") == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	s.full.Add(1)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Link", `<https://api.github.com/x?page=2>; rel="next"`)
	if _, err := io.WriteString(w, body); err != nil {
		panic(err)
	}
}

// newClient returns a transport, the server behind it, the server's URL,
// and an HTTP client using the transport.
func newClient(t *testing.T) (*Transport, *etagServer, string, *http.Client) {
	t.Helper()
	srv := newETagServer()
	ts := httptest.NewServer(srv)
	t.Cleanup(ts.Close)
	transport := NewTransport(nil)
	transport.OnStoreError = func(err error) { t.Errorf("store error: %v", err) }
	return transport, srv, ts.URL, &http.Client{Transport: transport}
}

func get(t *testing.T, c *http.Client, url, token string) (*http.Response, string) {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		t.Fatal(err)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := c.Do(req)
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if err := resp.Body.Close(); err != nil {
		t.Fatal(err)
	}
	return resp, string(body)
}

func TestTransportServesFromCacheOn304(t *testing.T) {
	transport, srv, base, c := newClient(t)
	srv.set("/repos/o/r", `{"id":1}`)
	url := base + "/repos/o/r"

	resp, body := get(t, c, url, "tok")
	if resp.StatusCode != 200 || body != `{"id":1}` {
		t.Fatalf("first GET = %d %q", resp.StatusCode, body)
	}

	resp, body = get(t, c, url, "tok")
	if resp.StatusCode != 200 || body != `{"id":1}` {
		t.Fatalf("second GET = %d %q, want cached 200", resp.StatusCode, body)
	}
	if resp.Header.Get("Link") == "" || resp.Header.Get("Content-Type") != "application/json" {
		t.Errorf("cached headers missing: %v", resp.Header)
	}
	if got := resp.Header.Get("X-RateLimit-Remaining"); got != "4998" {
		t.Errorf("X-RateLimit-Remaining = %q, want the 304's fresh value 4998", got)
	}
	if resp.ContentLength != int64(len(body)) {
		t.Errorf("ContentLength = %d, want %d", resp.ContentLength, len(body))
	}
	if srv.full.Load() != 1 {
		t.Errorf("full responses = %d, want 1", srv.full.Load())
	}

	stats := transport.Stats()
	if stats.Requests != 2 || stats.Hits != 1 || stats.Misses != 1 || stats.Uncacheable != 0 {
		t.Errorf("Stats = %+v, want 2 requests, 1 hit, 1 miss", stats)
	}
	if !strings.Contains(stats.String(), "1 served from cache") {
		t.Errorf("Stats.String() = %q", stats.String())
	}
}

func TestTransportRefetchesWhenChanged(t *testing.T) {
	transport, srv, base, c := newClient(t)
	url := base + "/repos/o/r"
	srv.set("/repos/o/r", `v1`)
	get(t, c, url, "tok")

	srv.set("/repos/o/r", `v2`)
	resp, body := get(t, c, url, "tok")
	if resp.StatusCode != 200 || body != "v2" {
		t.Fatalf("GET after change = %d %q, want fresh v2", resp.StatusCode, body)
	}
	// The new version is now cached.
	_, body = get(t, c, url, "tok")
	if body != "v2" || srv.full.Load() != 2 {
		t.Errorf("third GET body = %q, full responses = %d; want v2 served from cache", body, srv.full.Load())
	}
	if s := transport.Stats(); s.Hits != 1 || s.Misses != 2 {
		t.Errorf("Stats = %+v, want 1 hit, 2 misses", s)
	}
}

func TestTransportKeysByCredential(t *testing.T) {
	transport, srv, base, c := newClient(t)
	url := base + "/user"
	srv.set("/user", `me`)

	get(t, c, url, "alice")
	get(t, c, url, "bob")
	if srv.full.Load() != 2 {
		t.Errorf("full responses = %d, want 2: a second credential must not hit the first's entry", srv.full.Load())
	}
	get(t, c, url, "alice")
	if s := transport.Stats(); s.Hits != 1 {
		t.Errorf("Stats = %+v, want 1 hit for alice's repeat", s)
	}
}

func TestTransportBypasses(t *testing.T) {
	transport, srv, base, c := newClient(t)
	srv.set("/repos/o/r", `x`)

	t.Run("POST", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPost, base+"/repos/o/r", strings.NewReader("{}"))
		if err != nil {
			t.Fatal(err)
		}
		resp, err := c.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		if err := resp.Body.Close(); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("404", func(t *testing.T) {
		resp, _ := get(t, c, base+"/missing", "tok")
		if resp.StatusCode != 404 {
			t.Errorf("status = %d, want 404 passed through", resp.StatusCode)
		}
	})
	t.Run("no ETag", func(t *testing.T) {
		get(t, c, base+"/repos/o/r?noetag=1", "tok")
		get(t, c, base+"/repos/o/r?noetag=1", "tok")
	})
	if s := transport.Stats(); s.Uncacheable != 4 || s.Hits != 0 {
		t.Errorf("Stats = %+v, want 4 uncacheable, 0 hits", s)
	}
}

func TestTransportPreservesCallerConditional(t *testing.T) {
	_, srv, base, c := newClient(t)
	url := base + "/repos/o/r"
	srv.set("/repos/o/r", `x`)
	get(t, c, url, "tok")

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer tok")
	req.Header.Set("If-None-Match", `"something-else"`)
	resp, err := c.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if err := resp.Body.Close(); err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 || srv.full.Load() != 2 {
		t.Errorf("status = %d, full = %d; the caller's own validator must be sent as is", resp.StatusCode, srv.full.Load())
	}
}

func TestTransportNetworkError(t *testing.T) {
	transport := NewTransport(nil)
	c := &http.Client{Transport: transport}
	if _, err := c.Get("http://127.0.0.1:1/unreachable"); err == nil {
		t.Error("expected a transport error")
	}
}

type failingStore struct{ err error }

func (f failingStore) Get(string) (*Entry, error) { return nil, f.err }
func (f failingStore) Set(string, *Entry) error   { return f.err }

func TestTransportStoreErrors(t *testing.T) {
	boom := errors.New("boom")
	srv := newETagServer()
	srv.set("/x", "x")
	ts := httptest.NewServer(srv)
	t.Cleanup(ts.Close)

	var reported []error
	transport := NewTransportWithStore(nil, failingStore{err: boom})
	transport.OnStoreError = func(err error) { reported = append(reported, err) }
	c := &http.Client{Transport: transport}

	resp, err := c.Get(ts.URL + "/x")
	if err != nil {
		t.Fatalf("request must succeed despite store errors: %v", err)
	}
	if err := resp.Body.Close(); err != nil {
		t.Fatal(err)
	}
	if len(reported) != 2 || !errors.Is(reported[0], boom) || !errors.Is(reported[1], boom) {
		t.Errorf("reported = %v, want Get and Set errors", reported)
	}
}

func TestKey(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "https://api.github.com/user?x=1", nil)
	if err != nil {
		t.Fatal(err)
	}
	if got := Key(req); got != "GET https://api.github.com/user?x=1" {
		t.Errorf("Key() without auth = %q", got)
	}
	req.Header.Set("Authorization", "Bearer secret")
	withAuth := Key(req)
	if !strings.HasPrefix(withAuth, "GET https://api.github.com/user?x=1 ") || strings.Contains(withAuth, "secret") {
		t.Errorf("Key() with auth = %q, want URL plus a hash, never the token", withAuth)
	}
}

func TestMemoryStoreLRU(t *testing.T) {
	s := NewMemoryStore(2)
	for _, k := range []string{"a", "b"} {
		if err := s.Set(k, &Entry{ETag: k}); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := s.Get("a"); err != nil { // a is now most recent
		t.Fatal(err)
	}
	if err := s.Set("c", &Entry{ETag: "c"}); err != nil {
		t.Fatal(err)
	}
	if s.Len() != 2 {
		t.Errorf("Len() = %d, want 2", s.Len())
	}
	if e, _ := s.Get("b"); e != nil {
		t.Error("b should have been evicted as least recently used")
	}
	for _, k := range []string{"a", "c"} {
		if e, _ := s.Get(k); e == nil || e.ETag != k {
			t.Errorf("Get(%q) = %v, want entry", k, e)
		}
	}
	// Updating an existing key keeps the count.
	if err := s.Set("a", &Entry{ETag: "a2"}); err != nil {
		t.Fatal(err)
	}
	if e, _ := s.Get("a"); s.Len() != 2 || e.ETag != "a2" {
		t.Errorf("after update: Len() = %d, a = %v", s.Len(), e)
	}
}

func TestMemoryStoreUnlimited(t *testing.T) {
	s := NewMemoryStore(0)
	for i := range 100 {
		if err := s.Set(fmt.Sprint(i), &Entry{}); err != nil {
			t.Fatal(err)
		}
	}
	if s.Len() != 100 {
		t.Errorf("Len() = %d, want 100", s.Len())
	}
}

func TestFileStore(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "cache")
	s, err := NewFileStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	if s.Dir() != dir {
		t.Errorf("Dir() = %q, want %q", s.Dir(), dir)
	}
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != 0o700 {
		t.Errorf("dir perm = %o, want 700", perm)
	}

	if e, err := s.Get("missing"); err != nil || e != nil {
		t.Errorf("Get(missing) = %v, %v; want nil, nil", e, err)
	}

	entry := &Entry{ETag: `"abc"`, Header: http.Header{"Link": {"x"}}, Body: []byte(`{"a":1}`)}
	if err := s.Set("GET https://api.github.com/x", entry); err != nil {
		t.Fatal(err)
	}
	got, err := s.Get("GET https://api.github.com/x")
	if err != nil {
		t.Fatal(err)
	}
	if got == nil || got.ETag != entry.ETag || string(got.Body) != string(entry.Body) || got.Header.Get("Link") != "x" {
		t.Errorf("Get() = %+v, want %+v", got, entry)
	}

	files, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 {
		t.Fatalf("files = %v, want one entry and no temp files", files)
	}
	finfo, err := files[0].Info()
	if err != nil {
		t.Fatal(err)
	}
	if perm := finfo.Mode().Perm(); perm != 0o600 {
		t.Errorf("file perm = %o, want 600", perm)
	}

	// A corrupt file is a miss, not an error.
	if err := os.WriteFile(filepath.Join(dir, files[0].Name()), []byte("not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	if e, err := s.Get("GET https://api.github.com/x"); err != nil || e != nil {
		t.Errorf("Get(corrupt) = %v, %v; want nil, nil", e, err)
	}
}

func TestFileStoreSurvivesTransports(t *testing.T) {
	dir := t.TempDir()
	srv := newETagServer()
	srv.set("/x", "x")
	ts := httptest.NewServer(srv)
	t.Cleanup(ts.Close)

	for i := range 2 {
		store, err := NewFileStore(dir)
		if err != nil {
			t.Fatal(err)
		}
		transport := NewTransportWithStore(nil, store)
		c := &http.Client{Transport: transport}
		resp, err := c.Get(ts.URL + "/x")
		if err != nil {
			t.Fatal(err)
		}
		if err := resp.Body.Close(); err != nil {
			t.Fatal(err)
		}
		if i == 1 && transport.Stats().Hits != 1 {
			t.Errorf("second process: Stats = %+v, want a hit from the on-disk entry", transport.Stats())
		}
	}
	if srv.full.Load() != 1 {
		t.Errorf("full responses = %d, want 1", srv.full.Load())
	}
}

func TestNewFileStoreError(t *testing.T) {
	file := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(file, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := NewFileStore(filepath.Join(file, "sub")); err == nil {
		t.Error("NewFileStore under a regular file should fail")
	}
}
