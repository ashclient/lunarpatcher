package main

// Roblox game artwork is fetched from Roblox's public, unauthenticated thumbnail API
// when the user applies the patch. The resulting small, verified images are embedded
// into the modified renderer so that file:// windows do not depend on Roblox API CORS.
// We never request or retain Roblox account cookies or credentials.

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"
)

var robloxCatalogPlaces = []string{
	"4924922222", "2753915549", "6516141723", "920587237",
	"142823291", "1962086868", "189707", "17625359962",
}

type robloxIconRecord struct {
	TargetID int64  `json:"targetId"`
	State    string `json:"state"`
	ImageURL string `json:"imageUrl"`
}

type robloxIconResult struct {
	Data []robloxIconRecord `json:"data"`
}

// fetchRobloxArtwork is intentionally best-effort: an offline computer can still
// patch Lunar, while failed images fall back to neutral icon placeholders.
func fetchRobloxArtwork(ctx context.Context) map[string]string {
	ctx, cancel := context.WithTimeout(ctx, 11*time.Second)
	defer cancel()
	client := &http.Client{Timeout: 8 * time.Second}
	artwork, _ := fetchRobloxArtworkWithClient(ctx, client, "https://thumbnails.roblox.com", robloxCatalogPlaces)
	return artwork
}

func fetchRobloxArtworkWithClient(ctx context.Context, client *http.Client, apiHost string, places []string) (map[string]string, error) {
	icons := map[string]string{}
	if len(places) == 0 {
		return icons, nil
	}
	if len(places) > 24 {
		return icons, fmt.Errorf("too many icon requests")
	}
	for _, id := range places {
		if !validRobloxPlaceID(id) {
			return icons, fmt.Errorf("invalid place ID")
		}
	}
	params := url.Values{}
	params.Set("placeIds", strings.Join(places, ","))
	params.Set("returnPolicy", "PlaceHolder")
	params.Set("size", "512x512")
	params.Set("format", "Png")
	params.Set("isCircular", "false")
	endpoint := strings.TrimRight(apiHost, "/") + "/v1/places/gameicons?" + params.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return icons, err
	}
	req.Header.Set("Accept", "application/json")
	response, err := client.Do(req)
	if err != nil {
		return icons, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return icons, fmt.Errorf("Roblox thumbnails HTTP %d", response.StatusCode)
	}
	raw, err := io.ReadAll(io.LimitReader(response.Body, 512*1024))
	if err != nil {
		return icons, err
	}
	var payload robloxIconResult
	if err := json.Unmarshal(raw, &payload); err != nil {
		return icons, err
	}
	expected := map[string]bool{}
	for _, p := range places {
		expected[p] = true
	}
	sem := make(chan struct{}, 4)
	var wg sync.WaitGroup
	var mu sync.Mutex
	for _, record := range payload.Data {
		r := record
		id := fmt.Sprint(r.TargetID)
		if !expected[id] || r.State != "Completed" || !validRobloxCDNURL(r.ImageURL) {
			continue
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			select {
			case sem <- struct{}{}:
			case <-ctx.Done():
				return
			}
			defer func() { <-sem }()
			img, err := downloadRobloxImage(ctx, client, r.ImageURL)
			if err != nil {
				return
			}
			mu.Lock()
			icons[id] = img
			mu.Unlock()
		}()
	}
	wg.Wait()
	return icons, nil
}

func validRobloxPlaceID(id string) bool {
	if len(id) == 0 || len(id) > 18 || id[0] == '0' {
		return false
	}
	for _, c := range id {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func validRobloxCDNURL(s string) bool {
	u, err := url.Parse(s)
	if err != nil || u.Scheme != "https" || u.User != nil || u.Port() != "" {
		return false
	}
	host := strings.ToLower(u.Hostname())
	return host == "tr.rbxcdn.com" || strings.HasSuffix(host, ".rbxcdn.com")
}

func downloadRobloxImage(ctx context.Context, client *http.Client, s string) (string, error) {
	if !validRobloxCDNURL(s) {
		return "", fmt.Errorf("untrusted thumbnail URL")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s, nil)
	if err != nil {
		return "", err
	}
	response, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		return "", fmt.Errorf("thumbnail HTTP %d", response.StatusCode)
	}
	if response.ContentLength > 650*1024 {
		return "", fmt.Errorf("thumbnail too large")
	}
	// Followed redirects must not leave Roblox's CDN.
	if response.Request != nil && response.Request.URL != nil && !validRobloxCDNURL(response.Request.URL.String()) {
		return "", fmt.Errorf("thumbnail redirected to another host")
	}
	b, err := io.ReadAll(io.LimitReader(response.Body, 650*1024+1))
	if err != nil {
		return "", err
	}
	if len(b) > 650*1024 {
		return "", fmt.Errorf("thumbnail too large")
	}
	mime := http.DetectContentType(b)
	if mime != "image/png" && mime != "image/jpeg" && mime != "image/webp" {
		return "", fmt.Errorf("invalid thumbnail content type: %s", mime)
	}
	return "data:" + mime + ";base64," + base64.StdEncoding.EncodeToString(b), nil
}

func sortedArtworkIDs(art map[string]string) []string {
	ids := make([]string, 0, len(art))
	for id := range art {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}
