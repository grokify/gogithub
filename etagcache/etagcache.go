// Package etagcache provides an HTTP transport that makes conditional
// requests against the GitHub API.
//
// GitHub returns an ETag with most GET responses and answers a repeat request
// carrying If-None-Match with 304 Not Modified when nothing changed. A 304
// does not count against the rate limit. Transport remembers the ETag and
// body of each successful GET, sends the conditional headers on the next
// request for the same URL, and when GitHub answers 304, returns the cached
// response as if it were a fresh 200, so callers such as go-github never see
// the 304.
//
// Use it as the base transport of an authenticated client:
//
//	cache := etagcache.NewTransport(nil)
//	client, err := clientv1.NewClientWithOptions(ctx, clientv1.ClientOptions{
//	    Token:     token,
//	    Transport: cache,
//	})
//	// ... make requests, then:
//	fmt.Println(cache.Stats())
//
// Entries are keyed by URL and by a hash of the Authorization header, so a
// store shared by clients with different tokens never serves one token's
// response to another. Cached bodies may contain private data; FileStore
// writes them with owner-only permissions.
package etagcache

import (
	"bytes"
	"container/list"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
)

// Entry is a cached response.
type Entry struct {
	ETag         string      `json:"etag"`
	LastModified string      `json:"lastModified,omitempty"`
	Header       http.Header `json:"header"`
	Body         []byte      `json:"body"`
}

// Store persists entries. Implementations must be safe for concurrent use.
type Store interface {
	// Get returns the entry for key, or nil when there is none.
	Get(key string) (*Entry, error)
	// Set stores the entry for key, replacing any existing one.
	Set(key string, entry *Entry) error
}

// Stats counts what the transport did.
type Stats struct {
	// Requests is the number of requests passed through the transport.
	Requests int64
	// Hits is the number of requests GitHub answered with 304 Not Modified;
	// these did not count against the rate limit.
	Hits int64
	// Misses is the number of cacheable requests that were sent without a
	// cached entry, or whose entry was stale.
	Misses int64
	// Uncacheable is the number of requests that bypassed the cache: methods
	// other than GET and HEAD, or responses without an ETag.
	Uncacheable int64
}

func (s Stats) String() string {
	return fmt.Sprintf("%d requests, %d served from cache, %d fetched, %d uncacheable", s.Requests, s.Hits, s.Misses, s.Uncacheable)
}

// Transport makes conditional requests using a Store. The zero value is not
// usable; use NewTransport.
type Transport struct {
	// Next performs the requests. nil means http.DefaultTransport.
	Next http.RoundTripper
	// Store holds cached responses.
	Store Store
	// OnStoreError is called when the store fails to read or write; the
	// request still proceeds without the cache. nil ignores store errors.
	OnStoreError func(err error)

	requests, hits, misses, uncacheable atomic.Int64
}

// NewTransport returns a Transport over next (nil means
// http.DefaultTransport) backed by a new MemoryStore.
func NewTransport(next http.RoundTripper) *Transport {
	return &Transport{Next: next, Store: NewMemoryStore(0)}
}

// NewTransportWithStore returns a Transport over next (nil means
// http.DefaultTransport) backed by store.
func NewTransportWithStore(next http.RoundTripper, store Store) *Transport {
	return &Transport{Next: next, Store: store}
}

// Stats returns the counts so far.
func (t *Transport) Stats() Stats {
	return Stats{
		Requests:    t.requests.Load(),
		Hits:        t.hits.Load(),
		Misses:      t.misses.Load(),
		Uncacheable: t.uncacheable.Load(),
	}
}

// RoundTrip implements http.RoundTripper.
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.requests.Add(1)
	next := t.Next
	if next == nil {
		next = http.DefaultTransport
	}
	if (req.Method != http.MethodGet && req.Method != http.MethodHead) || t.Store == nil {
		t.uncacheable.Add(1)
		return next.RoundTrip(req)
	}

	key := Key(req)
	cached, err := t.Store.Get(key)
	if err != nil {
		t.storeError(err)
		cached = nil
	}
	if cached != nil && req.Header.Get("If-None-Match") == "" && req.Header.Get("If-Modified-Since") == "" {
		req = req.Clone(req.Context())
		req.Header.Set("If-None-Match", cached.ETag)
		if cached.LastModified != "" {
			req.Header.Set("If-Modified-Since", cached.LastModified)
		}
	}

	resp, err := next.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusNotModified && cached != nil {
		t.hits.Add(1)
		return t.fromCache(resp, cached), nil
	}

	if resp.StatusCode != http.StatusOK || resp.Header.Get("ETag") == "" {
		t.uncacheable.Add(1)
		return resp, nil
	}
	t.misses.Add(1)

	body, err := io.ReadAll(resp.Body)
	closeErr := resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("etagcache: read response body: %w", err)
	}
	if closeErr != nil {
		return nil, fmt.Errorf("etagcache: close response body: %w", closeErr)
	}
	resp.Body = io.NopCloser(bytes.NewReader(body))

	if err := t.Store.Set(key, &Entry{
		ETag:         resp.Header.Get("ETag"),
		LastModified: resp.Header.Get("Last-Modified"),
		Header:       resp.Header.Clone(),
		Body:         body,
	}); err != nil {
		t.storeError(err)
	}
	return resp, nil
}

// fromCache builds a 200 response from a cached entry and the 304 that
// validated it. Headers come from the cached response, except those GitHub
// sends fresh with the 304: rate limit and request identification.
func (t *Transport) fromCache(notModified *http.Response, cached *Entry) *http.Response {
	header := cached.Header.Clone()
	for name, values := range notModified.Header {
		if isFreshHeader(name) {
			header[name] = values
		}
	}
	// The 304 body is empty by definition; release the connection.
	if notModified.Body != nil {
		if err := notModified.Body.Close(); err != nil {
			t.storeError(fmt.Errorf("etagcache: close 304 body: %w", err))
		}
	}
	return &http.Response{
		Status:        "200 OK",
		StatusCode:    http.StatusOK,
		Proto:         notModified.Proto,
		ProtoMajor:    notModified.ProtoMajor,
		ProtoMinor:    notModified.ProtoMinor,
		Header:        header,
		Body:          io.NopCloser(bytes.NewReader(cached.Body)),
		ContentLength: int64(len(cached.Body)),
		Request:       notModified.Request,
		TLS:           notModified.TLS,
	}
}

// isFreshHeader reports whether a header on a 304 should replace the cached
// one. Rate limit headers describe the current request; Date and request IDs
// identify it.
func isFreshHeader(name string) bool {
	canonical := http.CanonicalHeaderKey(name)
	switch canonical {
	case "Date", "X-Github-Request-Id", "Etag", "Last-Modified":
		return true
	}
	return len(canonical) > len("X-Ratelimit-") && canonical[:len("X-Ratelimit-")] == "X-Ratelimit-"
}

func (t *Transport) storeError(err error) {
	if t.OnStoreError != nil {
		t.OnStoreError(err)
	}
}

// Key returns the cache key for a request: the method and URL, plus a hash
// of the Authorization header so that responses are never shared between
// credentials.
func Key(req *http.Request) string {
	auth := req.Header.Get("Authorization")
	if auth == "" {
		return req.Method + " " + req.URL.String()
	}
	sum := sha256.Sum256([]byte(auth))
	return req.Method + " " + req.URL.String() + " " + hex.EncodeToString(sum[:8])
}

// MemoryStore keeps entries in memory, evicting the least recently used
// when MaxEntries is exceeded. It is safe for concurrent use.
type MemoryStore struct {
	mu         sync.Mutex
	maxEntries int
	entries    map[string]*list.Element
	// order tracks recency: front is the least recently used.
	order *list.List
}

type memoryEntry struct {
	key   string
	entry *Entry
}

// NewMemoryStore returns a MemoryStore holding at most maxEntries entries;
// zero or negative means unlimited.
func NewMemoryStore(maxEntries int) *MemoryStore {
	return &MemoryStore{
		maxEntries: maxEntries,
		entries:    make(map[string]*list.Element),
		order:      list.New(),
	}
}

// Get implements Store.
func (s *MemoryStore) Get(key string) (*Entry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	el, ok := s.entries[key]
	if !ok {
		return nil, nil
	}
	s.order.MoveToBack(el)
	return el.Value.(*memoryEntry).entry, nil
}

// Set implements Store.
func (s *MemoryStore) Set(key string, entry *Entry) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if el, ok := s.entries[key]; ok {
		el.Value.(*memoryEntry).entry = entry
		s.order.MoveToBack(el)
		return nil
	}
	s.entries[key] = s.order.PushBack(&memoryEntry{key: key, entry: entry})
	for s.maxEntries > 0 && s.order.Len() > s.maxEntries {
		oldest := s.order.Front()
		s.order.Remove(oldest)
		delete(s.entries, oldest.Value.(*memoryEntry).key)
	}
	return nil
}

// Len returns the number of entries.
func (s *MemoryStore) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.order.Len()
}
