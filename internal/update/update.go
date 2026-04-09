package update

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/uqpay/uqpay-cli/internal/build"
)

const (
	repo             = "uqpay/uqpay-cli"
	checkIntervalHrs = 4
	npmPackage       = "@uqpay/cli"
)

type cache struct {
	LastCheck     int64  `json:"last_check"`
	LatestVersion string `json:"latest_version"`
	LastNotified  int64  `json:"last_notified"` // unix timestamp of last notification
}

func cachePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".uqpay", "version-check.json")
}

func loadCache() *cache {
	data, err := os.ReadFile(cachePath())
	if err != nil {
		return nil
	}
	var c cache
	if err := json.Unmarshal(data, &c); err != nil {
		return nil
	}
	return &c
}

func saveCache(c *cache) {
	data, _ := json.Marshal(c)
	os.MkdirAll(filepath.Dir(cachePath()), 0700)
	os.WriteFile(cachePath(), data, 0600)
}

func fetchLatest() (string, error) {
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get("https://api.github.com/repos/" + repo + "/releases/latest")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("status %d", resp.StatusCode)
	}
	var result struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return strings.TrimPrefix(result.TagName, "v"), nil
}

// compareVersions returns >0 if a > b, 0 if equal, <0 if a < b.
func compareVersions(a, b string) int {
	pa := strings.SplitN(a, ".", 3)
	pb := strings.SplitN(b, ".", 3)
	for i := 0; i < 3; i++ {
		var na, nb int
		if i < len(pa) {
			na, _ = strconv.Atoi(pa[i])
		}
		if i < len(pb) {
			nb, _ = strconv.Atoi(pb[i])
		}
		if na != nb {
			return na - nb
		}
	}
	return 0
}

// CheckForUpdate prints an update notice to stderr if a newer version is available.
// Non-blocking: uses cached result or fires a background goroutine.
func CheckForUpdate() {
	current := build.Version
	if current == "dev" || strings.Contains(current, "dirty") {
		return // skip for dev builds
	}

	now := time.Now().Unix()
	oneDayAgo := now - 86400

	c := loadCache()
	if c != nil && now-c.LastCheck < int64(checkIntervalHrs*3600) {
		if compareVersions(c.LatestVersion, current) > 0 && c.LastNotified < oneDayAgo {
			printNotice(current, c.LatestVersion)
			c.LastNotified = now
			saveCache(c)
		}
		return
	}

	// Background check
	go func() {
		latest, err := fetchLatest()
		if err != nil || latest == "" {
			return
		}
		var lastNotified int64
		if c != nil && c.LatestVersion == latest {
			lastNotified = c.LastNotified
		}
		newCache := &cache{LastCheck: now, LatestVersion: latest, LastNotified: lastNotified}
		if compareVersions(latest, current) > 0 && lastNotified < oneDayAgo {
			printNotice(current, latest)
			newCache.LastNotified = now
		}
		saveCache(newCache)
	}()
}

func printNotice(current, latest string) {
	fmt.Fprintf(os.Stderr, "\n  Update available: %s → %s\n  Run: uqpay upgrade\n\n", current, latest)
}

// LatestVersion fetches and returns the latest version (blocking).
func LatestVersion() (string, error) {
	return fetchLatest()
}

// NpmPackage returns the npm package name.
func NpmPackage() string {
	return npmPackage
}
