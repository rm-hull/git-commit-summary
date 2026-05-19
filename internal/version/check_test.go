package version

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCheckLatestReturnsLatestWhenNewer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		body, err := json.Marshal(LatestResponse{
			Version: "v0.2.0",
			Time:    "2026-05-19T12:00:00Z",
			Origin: Origin{
				VCS:  "git",
				URL:  "https://github.com/rm-hull/git-commit-summary",
				Hash: "abc123",
				Ref:  "refs/tags/v0.2.0",
			},
		})
		require.NoError(t, err)
		_, _ = w.Write(body)
	}))
	defer server.Close()

	oldURL := latestURL
	oldClient := httpClient
	latestURL = server.URL
	httpClient = http.Client{Transport: http.DefaultTransport}
	defer func() {
		latestURL = oldURL
		httpClient = oldClient
	}()

	latest, err := CheckLatest("0.1.0")
	require.NoError(t, err)
	require.Equal(t, "v0.2.0", latest)
}

func TestCheckLatestReturnsEmptyWhenUpToDate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		body, err := json.Marshal(LatestResponse{
			Version: "v1.0.0",
			Time:    "2026-05-19T12:00:00Z",
			Origin: Origin{
				VCS:  "git",
				URL:  "https://github.com/rm-hull/git-commit-summary",
				Hash: "abc123",
				Ref:  "refs/tags/v1.0.0",
			},
		})
		require.NoError(t, err)
		_, _ = w.Write(body)
	}))
	defer server.Close()

	oldURL := latestURL
	oldClient := httpClient
	latestURL = server.URL
	httpClient = http.Client{Transport: http.DefaultTransport}
	defer func() {
		latestURL = oldURL
		httpClient = oldClient
	}()

	latest, err := CheckLatest("1.0.0")
	require.NoError(t, err)
	require.Empty(t, latest)
}

func TestCheckLatestReturnsEmptyForDevel(t *testing.T) {
	latest, err := CheckLatest("devel")
	require.NoError(t, err)
	require.Empty(t, latest)
}

func TestCheckLatestReturnsErrorOnBadStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	oldURL := latestURL
	oldClient := httpClient
	latestURL = server.URL
	httpClient = http.Client{Transport: http.DefaultTransport}
	defer func() {
		latestURL = oldURL
		httpClient = oldClient
	}()

	_, err := CheckLatest("0.1.0")
	require.Error(t, err)
}
