package fetcher

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

// Fetcher wraps an HTTP client.
type Fetcher struct {
	client *http.Client
}

// NewFetcher constructs a new client wrapper with a given client.
func NewFetcher(client *http.Client) *Fetcher {
	return &Fetcher{
		client: client,
	}
}

type FetchOptionsFn func(opt *FetchOptions)

// WithUsername provides optional username to the URL fetch.
func WithUsername(username string) FetchOptionsFn {
	return func(opt *FetchOptions) {
		opt.username = username
	}
}

// WithPassword provides optional password to the URL fetch.
func WithPassword(password string) FetchOptionsFn {
	return func(opt *FetchOptions) {
		opt.password = password
	}
}

// WithToken provides optional token to the URL fetch.
func WithToken(token string) FetchOptionsFn {
	return func(opt *FetchOptions) {
		opt.token = token
	}
}

type FetchOptions struct {
	username string
	password string
	token    string
}

// Fetch constructs a request and does a client.Do with it.
func (f *Fetcher) Fetch(ctx context.Context, url string, opts ...FetchOptionsFn) ([]byte, error) {
	opt := &FetchOptions{}
	for _, fn := range opts {
		fn(opt)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate request for url '%s': %w", url, err)
	}

	if opt.username != "" && opt.password != "" {
		req.SetBasicAuth(opt.username, opt.password)
	}

	if opt.token != "" {
		req.Header.Add("Authorization", "Bearer "+opt.token)
	}

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch data: %w", err)
	}

	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			err = fmt.Errorf("failed to close with %w after %w", closeErr, err)
		}
	}()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("failed to fetch url content with status code %d", resp.StatusCode)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %w", err)
	}

	return content, nil
}
