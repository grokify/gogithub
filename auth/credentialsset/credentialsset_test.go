package credentialsset

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const testAccount = "github-app"

// tokenServer fakes an OAuth token endpoint and records the forms it receives.
type tokenServer struct {
	t     *testing.T
	forms []url.Values
	url   string
}

func newTokenServer(t *testing.T) *tokenServer {
	t.Helper()
	s := &tokenServer{t: t}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Errorf("parse token form: %v", err)
		}
		s.forms = append(s.forms, r.PostForm)
		w.Header().Set("Content-Type", "application/json")
		if r.PostForm.Get("code") != "good-code" {
			w.WriteHeader(http.StatusBadRequest)
			s.write(w, `{"error":"bad_verification_code"}`)
			return
		}
		s.write(w, `{"access_token":"oauth-token","token_type":"bearer","scope":"repo,read:org"}`)
	}))
	t.Cleanup(srv.Close)
	s.url = srv.URL
	return s
}

func (s *tokenServer) write(w http.ResponseWriter, body string) {
	if _, err := fmt.Fprint(w, body); err != nil {
		s.t.Errorf("write response: %v", err)
	}
}

// writeCredentialsSet writes a credentials set file holding one account and
// returns its path.
func writeCredentialsSet(t *testing.T, account map[string]any) string {
	t.Helper()
	b, err := json.Marshal(map[string]any{
		"credentials": map[string]any{testAccount: account},
	})
	if err != nil {
		t.Fatalf("marshal credentials set: %v", err)
	}
	filename := filepath.Join(t.TempDir(), "credentials.json")
	if err := os.WriteFile(filename, b, 0600); err != nil {
		t.Fatalf("write credentials set: %v", err)
	}
	return filename
}

// authCodeAccount returns an authorization code account that exchanges codes
// at tokenURL.
func authCodeAccount(tokenURL string) map[string]any {
	return map[string]any{
		"type": "oauth2",
		"oauth2": map[string]any{
			"clientID":     "client-abc",
			"clientSecret": "secret-xyz",
			"redirectURL":  RedirectURL,
			"scope":        []string{"repo", "read:org"},
			"grantType":    "authorization_code",
			"endpoint": map[string]any{
				"AuthURL":  "https://github.example/login/oauth/authorize",
				"TokenURL": tokenURL,
			},
		},
	}
}

func failPrompt(t *testing.T) AuthCodePrompt {
	return func(context.Context, string, string) (string, error) {
		t.Error("prompt called, want no prompt")
		return "", nil
	}
}

func TestNewTokenAuthorizationCode(t *testing.T) {
	srv := newTokenServer(t)
	filename := writeCredentialsSet(t, authCodeAccount(srv.url))

	var gotURL, gotState string
	token, err := NewToken(context.Background(), filename, testAccount,
		func(_ context.Context, authURL, state string) (string, error) {
			gotURL, gotState = authURL, state
			return "  good-code\n", nil
		})
	if err != nil {
		t.Fatalf("NewToken() error = %v", err)
	}

	if token.AccessToken != "oauth-token" {
		t.Errorf("AccessToken = %q, want %q", token.AccessToken, "oauth-token")
	}

	u, err := url.Parse(gotURL)
	if err != nil {
		t.Fatalf("parse auth URL %q: %v", gotURL, err)
	}
	if got, want := u.Host+u.Path, "github.example/login/oauth/authorize"; got != want {
		t.Errorf("auth URL = %q, want %q", got, want)
	}
	q := u.Query()
	for param, want := range map[string]string{
		"client_id":     "client-abc",
		"redirect_uri":  RedirectURL,
		"scope":         "repo read:org",
		"response_type": "code",
		"state":         gotState,
	} {
		if got := q.Get(param); got != want {
			t.Errorf("auth URL %s = %q, want %q", param, got, want)
		}
	}
	if len(gotState) != 2*stateBytes {
		t.Errorf("state = %q, want %d hex characters", gotState, 2*stateBytes)
	}

	if len(srv.forms) != 1 {
		t.Fatalf("token requests = %d, want 1", len(srv.forms))
	}
	if got := srv.forms[0].Get("code"); got != "good-code" {
		t.Errorf("exchanged code = %q, want %q (trimmed)", got, "good-code")
	}
	if got := srv.forms[0].Get("redirect_uri"); got != RedirectURL {
		t.Errorf("exchange redirect_uri = %q, want %q", got, RedirectURL)
	}
}

func TestNewTokenStateIsRandom(t *testing.T) {
	srv := newTokenServer(t)
	filename := writeCredentialsSet(t, authCodeAccount(srv.url))

	var states []string
	for range 2 {
		_, err := NewToken(context.Background(), filename, testAccount,
			func(_ context.Context, _, state string) (string, error) {
				states = append(states, state)
				return "good-code", nil
			})
		if err != nil {
			t.Fatalf("NewToken() error = %v", err)
		}
	}
	if states[0] == states[1] {
		t.Errorf("state reused across calls: %q", states[0])
	}
}

func TestNewTokenStoredToken(t *testing.T) {
	srv := newTokenServer(t)

	tests := []struct {
		name  string
		store func(account map[string]any)
	}{
		{"on account", func(account map[string]any) {
			account["token"] = map[string]any{"access_token": "stored-token"}
		}},
		{"on oauth2 credentials", func(account map[string]any) {
			oauth2Creds, ok := account["oauth2"].(map[string]any)
			if !ok {
				t.Fatal("account has no oauth2 credentials")
			}
			oauth2Creds["token"] = map[string]any{"access_token": "stored-token"}
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			account := authCodeAccount(srv.url)
			tt.store(account)
			filename := writeCredentialsSet(t, account)

			token, err := NewToken(context.Background(), filename, testAccount, failPrompt(t))
			if err != nil {
				t.Fatalf("NewToken() error = %v", err)
			}
			if token.AccessToken != "stored-token" {
				t.Errorf("AccessToken = %q, want %q", token.AccessToken, "stored-token")
			}
		})
	}
	if len(srv.forms) != 0 {
		t.Errorf("token requests = %d, want 0 when a token is stored", len(srv.forms))
	}
}

func TestNewTokenErrors(t *testing.T) {
	srv := newTokenServer(t)
	errPrompt := errors.New("prompt failed")
	prompt := func(code string, err error) AuthCodePrompt {
		return func(context.Context, string, string) (string, error) { return code, err }
	}

	tests := []struct {
		name         string
		account      map[string]any
		lookup       string
		prompt       AuthCodePrompt
		wantErr      error
		wantContains string
	}{
		{
			name:    "prompt error",
			account: authCodeAccount(srv.url),
			prompt:  prompt("", errPrompt),
			wantErr: errPrompt,
		},
		{
			name:    "empty code",
			account: authCodeAccount(srv.url),
			prompt:  prompt(" \n", nil),
			wantErr: ErrNoAuthCode,
		},
		{
			name:    "nil prompt",
			account: authCodeAccount(srv.url),
			wantErr: ErrNoAuthCode,
		},
		{
			name:         "rejected code",
			account:      authCodeAccount(srv.url),
			prompt:       prompt("bad-code", nil),
			wantContains: "exchange authorization code",
		},
		{
			name: "not oauth2",
			account: map[string]any{
				"type":  "basic",
				"basic": map[string]any{"username": "user"},
			},
			prompt:  failPrompt(t),
			wantErr: ErrNotOAuth2,
		},
		{
			name:         "unknown account lists valid accounts",
			account:      authCodeAccount(srv.url),
			lookup:       "missing",
			prompt:       failPrompt(t),
			wantContains: testAccount,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filename := writeCredentialsSet(t, tt.account)
			lookup := tt.lookup
			if lookup == "" {
				lookup = testAccount
			}

			_, err := NewToken(context.Background(), filename, lookup, tt.prompt)
			if err == nil {
				t.Fatal("NewToken() error = nil, want error")
			}
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("NewToken() error = %v, want %v", err, tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantContains) {
				t.Errorf("NewToken() error = %q, want it to contain %q", err.Error(), tt.wantContains)
			}
		})
	}
}

func TestNewTokenMissingFile(t *testing.T) {
	filename := filepath.Join(t.TempDir(), "missing.json")
	if _, err := NewToken(context.Background(), filename, testAccount, failPrompt(t)); err == nil {
		t.Fatal("NewToken() error = nil, want error")
	}
}

func TestNewClient(t *testing.T) {
	account := authCodeAccount("https://github.example/login/oauth/access_token")
	account["token"] = map[string]any{"access_token": "stored-token"}
	filename := writeCredentialsSet(t, account)

	client, err := NewClient(context.Background(), filename, testAccount, failPrompt(t))
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	if client == nil {
		t.Fatal("NewClient() = nil, want client")
	}
}

func TestPromptReadWriter(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantCode string
		wantErr  bool
	}{
		{"code with newline", "abc123\n", "abc123\n", false},
		{"code without newline", "abc123", "abc123", false},
		{"no input", "", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			prompt := PromptReadWriter(&out, strings.NewReader(tt.input))

			code, err := prompt(context.Background(), "https://github.example/authorize?state=s1", "s1")
			if (err != nil) != tt.wantErr {
				t.Fatalf("prompt() error = %v, wantErr %v", err, tt.wantErr)
			}
			if code != tt.wantCode {
				t.Errorf("code = %q, want %q", code, tt.wantCode)
			}
			for _, want := range []string{"https://github.example/authorize?state=s1", "state s1"} {
				if !strings.Contains(out.String(), want) {
					t.Errorf("instructions = %q, want them to contain %q", out.String(), want)
				}
			}
		})
	}
}
