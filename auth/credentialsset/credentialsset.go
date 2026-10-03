// Package credentialsset creates GitHub clients from a goauth credentials
// set file, using an OAuth app instead of a personal access token.
//
// A token issued to an OAuth app is accepted by organizations that forbid
// personal access tokens, provided the organization allows the OAuth app.
//
// It lives in its own package so that importing gogithub/auth does not pull
// in goauth and its dependencies.
package credentialsset

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/grokify/goauth"
	"github.com/grokify/goauth/authutil"
	"golang.org/x/oauth2"

	"github.com/grokify/gogithub/clientv1"
)

// RedirectURL is a hosted redirect page that displays the authorization code
// and state, for applications that cannot receive a redirect themselves.
// Register it as the OAuth app's authorization callback URL to use it.
const RedirectURL = "https://grokify.github.io/goauth/oauth2callback/"

// stateBytes is the number of random bytes in a generated OAuth state.
const stateBytes = 16

// ErrNotOAuth2 indicates the account's credentials are not OAuth 2.0 credentials.
var ErrNotOAuth2 = errors.New("credentials are not oauth2 credentials")

// ErrNoAuthCode indicates the prompt returned an empty authorization code.
var ErrNoAuthCode = errors.New("no authorization code provided")

// AuthCodePrompt asks the user to authorize the application at authURL and
// returns the authorization code shown on the redirect page. The state on the
// redirect page should match state.
type AuthCodePrompt func(ctx context.Context, authURL, state string) (code string, err error)

// PromptReadWriter returns an AuthCodePrompt that writes instructions to w
// and reads the authorization code from r. Pass os.Stderr as w to keep
// instructions out of a command's standard output.
func PromptReadWriter(w io.Writer, r io.Reader) AuthCodePrompt {
	return func(_ context.Context, authURL, state string) (string, error) {
		if _, err := fmt.Fprintf(w,
			"Open this URL in your browser and authorize the application:\n\n  %s\n\n"+
				"Confirm the redirect page shows state %s, then enter the authorization code: ",
			authURL, state); err != nil {
			return "", err
		}
		code, err := bufio.NewReader(r).ReadString('\n')
		if err != nil && (!errors.Is(err, io.EOF) || code == "") {
			return "", fmt.Errorf("read authorization code: %w", err)
		}
		return code, nil
	}
}

// NewClient creates a GitHub client for an account in a goauth credentials
// set file. See NewToken for how the token is obtained.
func NewClient(ctx context.Context, filename, account string, prompt AuthCodePrompt) (clientv1.Client, error) {
	token, err := NewToken(ctx, filename, account, prompt)
	if err != nil {
		return nil, err
	}
	return clientv1.NewClient(ctx, token.AccessToken)
}

// NewToken returns an OAuth token for an account in a goauth credentials set
// file.
//
// A valid token stored with the account is returned as is. Otherwise, for
// the authorization code grant, prompt is called with the authorization URL
// and the code it returns is exchanged for a token. The token is not written
// back to the file.
func NewToken(ctx context.Context, filename, account string, prompt AuthCodePrompt) (*oauth2.Token, error) {
	creds, err := goauth.NewCredentialsFromSetFile(filename, account, true)
	if err != nil {
		return nil, fmt.Errorf("read credentials for account %q: %w", account, err)
	}
	if creds.Type != goauth.TypeOAuth2 || creds.OAuth2 == nil {
		return nil, fmt.Errorf("account %q has type %q: %w", account, creds.Type, ErrNotOAuth2)
	}
	if token := storedToken(creds); token != nil {
		return token, nil
	}
	if !creds.OAuth2.IsGrantType(authutil.GrantTypeAuthorizationCode) {
		return creds.NewToken(ctx)
	}
	if prompt == nil {
		return nil, fmt.Errorf("account %q requires authorization: %w", account, ErrNoAuthCode)
	}

	state, err := newState()
	if err != nil {
		return nil, err
	}
	code, err := prompt(ctx, creds.OAuth2.AuthCodeURL(state, nil), state)
	if err != nil {
		return nil, err
	}
	code = strings.TrimSpace(code)
	if code == "" {
		return nil, ErrNoAuthCode
	}
	token, err := creds.OAuth2.Exchange(ctx, code, nil)
	if err != nil {
		return nil, fmt.Errorf("exchange authorization code: %w", err)
	}
	return token, nil
}

// storedToken returns a valid token saved with the credentials, or nil.
func storedToken(creds goauth.Credentials) *oauth2.Token {
	for _, token := range []*oauth2.Token{creds.Token, creds.OAuth2.Token} {
		if token.Valid() {
			return token
		}
	}
	return nil
}

// newState returns a random OAuth state value.
func newState() (string, error) {
	b := make([]byte, stateBytes)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate oauth state: %w", err)
	}
	return hex.EncodeToString(b), nil
}
