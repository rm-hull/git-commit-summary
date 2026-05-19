package version

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/cockroachdb/errors"
	"golang.org/x/mod/semver"
)

type Origin struct {
	VCS  string `json:"VCS"`
	URL  string `json:"URL"`
	Hash string `json:"Hash"`
	Ref  string `json:"Ref"`
}

type LatestResponse struct {
	Version string `json:"Version"`
	Time    string `json:"Time"`
	Origin  Origin `json:"Origin"`
}

var (
	latestURL  = "https://proxy.golang.org/github.com/rm-hull/git-commit-summary/@latest"
	httpClient = http.Client{Timeout: 2 * time.Second}
)

// CheckLatest fetches the latest version from proxy.golang.org and compares it with the current version.
// If a newer version is available, it returns the version string; otherwise, it returns an empty string.
func CheckLatest(currentVersion string) (string, error) {
	if currentVersion == "devel" {
		return "", nil
	}

	req, err := http.NewRequest("GET", latestURL, nil)
	if err != nil {
		return "", errors.Wrapf(err, "failed to create HTTP request")
	}
	req.Header.Set("Accept", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", errors.Wrapf(err, "failed to perform HTTP request")

	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return "", errors.Errorf("unexpected status code: %s", resp.Status)
	}

	var latest LatestResponse
	if err := json.NewDecoder(resp.Body).Decode(&latest); err != nil {
		return "", err
	}

	// Ensure currentVersion has 'v' prefix if it's missing, as semver package requires it.
	if len(currentVersion) > 0 && currentVersion[0] != 'v' {
		currentVersion = "v" + currentVersion
	}

	if semver.Compare(currentVersion, latest.Version) < 0 {
		return latest.Version, nil
	}

	return "", nil
}
