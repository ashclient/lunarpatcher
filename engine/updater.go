package main

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const releasesAPI = "https://api.github.com/repos/ashclient/lunarpatcher/releases/latest"
const setupAsset = "LunarPatcher_Setup.exe"
const checksumAsset = "LunarPatcher_Setup.exe.sha256"
const maxInstallerSize = 100 << 20

var safeTag = regexp.MustCompile(`^v?\d+\.\d+\.\d+(?:[-+][0-9A-Za-z.-]+)?$`)
var hexDigest = regexp.MustCompile(`(?i)^[a-f0-9]{64}$`)

type githubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}
type githubRelease struct {
	TagName    string        `json:"tag_name"`
	HTMLURL    string        `json:"html_url"`
	Draft      bool          `json:"draft"`
	Prerelease bool          `json:"prerelease"`
	Assets     []githubAsset `json:"assets"`
}

var updateClient = &http.Client{Timeout: 65 * time.Second, CheckRedirect: func(req *http.Request, via []*http.Request) error {
	if len(via) > 7 {
		return errors.New("too many redirects")
	}
	if req.URL.Scheme != "https" {
		return errors.New("non-HTTPS update redirect")
	}
	return nil
}}

func request(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "LunarPatcher/"+appVersion)
	req.Header.Set("Accept", "application/vnd.github+json")
	res, err := updateClient.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		res.Body.Close()
		return nil, fmt.Errorf("GitHub returned %s", res.Status)
	}
	return res, nil
}
func latestRelease(ctx context.Context) (githubRelease, error) {
	var r githubRelease
	res, err := request(ctx, releasesAPI)
	if err != nil {
		return r, err
	}
	defer res.Body.Close()
	err = json.NewDecoder(io.LimitReader(res.Body, 1<<20)).Decode(&r)
	if err != nil {
		return r, err
	}
	if r.Draft || r.Prerelease || !safeTag.MatchString(r.TagName) {
		return r, errors.New("no stable release")
	}
	return r, nil
}
func newerVersion(remote, current string) bool {
	trim := func(s string) []int {
		s = strings.TrimPrefix(s, "v")
		s = strings.SplitN(s, "-", 2)[0]
		s = strings.SplitN(s, "+", 2)[0]
		parts := strings.Split(s, ".")
		if len(parts) != 3 {
			return nil
		}
		out := make([]int, 3)
		for i, p := range parts {
			n, e := strconv.Atoi(p)
			if e != nil || n < 0 {
				return nil
			}
			out[i] = n
		}
		return out
	}
	a, b := trim(remote), trim(current)
	if len(a) != 3 || len(b) != 3 {
		return false
	}
	for i := 0; i < 3; i++ {
		if a[i] > b[i] {
			return true
		}
		if a[i] < b[i] {
			return false
		}
	}
	return false
}
func assetURL(r githubRelease, name string) (string, error) {
	for _, a := range r.Assets {
		if a.Name == name {
			expected := "https://github.com/ashclient/lunarpatcher/releases/download/" + r.TagName + "/" + name
			if a.BrowserDownloadURL != expected {
				return "", errors.New("unexpected release asset URL")
			}
			if a.Size < 1 || a.Size > maxInstallerSize {
				return "", errors.New("unexpected release asset size")
			}
			return expected, nil
		}
	}
	return "", fmt.Errorf("release asset %s not found", name)
}
func fetchLimited(ctx context.Context, url string, limit int64) ([]byte, error) {
	res, err := request(ctx, url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	b, err := io.ReadAll(io.LimitReader(res.Body, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(b)) > limit {
		return nil, errors.New("update exceeds size limit")
	}
	return b, nil
}
func downloadInstaller(ctx context.Context, r githubRelease) (string, error) {
	if !newerVersion(r.TagName, appVersion) {
		return "", errors.New("no newer version available")
	}
	link, err := assetURL(r, setupAsset)
	if err != nil {
		return "", err
	}
	hashLink, err := assetURL(r, checksumAsset)
	if err != nil {
		return "", err
	}
	checksum, err := fetchLimited(ctx, hashLink, 512)
	if err != nil {
		return "", err
	}
	expected := strings.Fields(string(checksum))
	if len(expected) == 0 || !hexDigest.MatchString(expected[0]) {
		return "", errors.New("invalid release checksum")
	}
	data, err := fetchLimited(ctx, link, maxInstallerSize)
	if err != nil {
		return "", err
	}
	computed := sha256.Sum256(data)
	declared, _ := hex.DecodeString(expected[0])
	if subtle.ConstantTimeCompare(computed[:], declared) != 1 {
		return "", errors.New("download SHA-256 mismatch")
	}
	dir := filepath.Join(os.TempDir(), "LunarPatcherUpdate", r.TagName)
	if err = os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	path := filepath.Join(dir, setupAsset)
	temp := path + ".part"
	if err = os.WriteFile(temp, data, 0600); err != nil {
		return "", err
	}
	if err = os.Rename(temp, path); err != nil {
		os.Remove(temp)
		return "", err
	}
	return path, nil
}
