package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
)

type fakeRobloxTransport struct{ requests []string }

func (f *fakeRobloxTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	f.requests = append(f.requests, req.URL.Host+req.URL.Path)
	var data []byte
	switch req.URL.Host {
	case "thumbnails.roblox.com":
		data = []byte(`{"data":[{"targetId":2753915549,"state":"Completed","imageUrl":"https://tr.rbxcdn.com/gameicon/512/512/Image/Png"},{"targetId":4924922222,"state":"Blocked","imageUrl":"https://tr.rbxcdn.com/blocked"},{"targetId":1,"state":"Completed","imageUrl":"https://evil.example/image"}]}`)
	case "tr.rbxcdn.com":
		sample, err := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+k3X8AAAAASUVORK5CYII=")
		if err != nil {
			return nil, err
		}
		data = sample
	default:
		return nil, fmt.Errorf("unexpected host %s", req.URL.Host)
	}
	return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(string(data))), ContentLength: int64(len(data)), Header: http.Header{"Content-Type": []string{"application/octet-stream"}}, Request: req}, nil
}
func TestOfficialRobloxArtworkFromPlaceIds(t *testing.T) {
	f := &fakeRobloxTransport{}
	client := &http.Client{Transport: f}
	art, err := fetchRobloxArtworkWithClient(context.Background(), client, "https://thumbnails.roblox.com", []string{"2753915549", "4924922222"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(art["2753915549"], "data:image/png;base64,") {
		t.Fatalf("missing official icon: %+v", art)
	}
	if len(art) != 1 {
		t.Fatalf("blocked icon should be omitted: %v", sortedArtworkIDs(art))
	}
	if len(f.requests) != 2 {
		t.Fatalf("should request API once, CDN once; got %v", f.requests)
	}
}
func TestOfficialRobloxArtworkURLSafety(t *testing.T) {
	for _, value := range []string{"http://tr.rbxcdn.com/foo", "https://tr.rbxcdn.com.evil.org/foo", "https://evil.org/foo", "https://user:pass@tr.rbxcdn.com/foo"} {
		if validRobloxCDNURL(value) {
			t.Errorf("allowed untrusted image %s", value)
		}
	}
	if !validRobloxCDNURL("https://tr.rbxcdn.com/something/512/512/Image/Png") {
		t.Fatal("official CDN disallowed")
	}
}
func TestRobloxArtworkEmbeddedInRenderer(t *testing.T) {
	art := map[string]string{"2753915549": "data:image/png;base64,YQ=="}
	if !strings.Contains(robloxWidgetJS, "__ROBLOX_ART__") {
		t.Fatal("artwork injection placeholder missing")
	}
	if !strings.Contains(robloxWidgetJS, "imageFor(primary.id)") {
		t.Fatal("featured art not wired")
	}
	if len(art) != 1 {
		t.Fatal("missing fixture")
	}
}

func TestPatchedRobloxImageDataInGeneratedASAR(t *testing.T) {
	raw, err := os.ReadFile("/mnt/data/app.asar")
	if os.IsNotExist(err) {
		t.Skip("original ASAR fixture not mounted")
	}
	if err != nil {
		t.Fatal(err)
	}
	img := "data:image/png;base64,YQ=="
	patched, _, err := patchSourceWithArtwork(raw, "ads", "#D676A5", true, map[string]string{"2753915549": img})
	if err != nil {
		t.Fatal(err)
	}
	a, err := parseAsar(patched)
	if err != nil {
		t.Fatal(err)
	}
	source, ok := a.entry("dist/assets/main.js")
	if !ok {
		t.Fatal("missing injected renderer")
	}
	if !bytes.Contains(source, []byte(img)) {
		t.Fatal("real image is not embedded in the actual ASAR")
	}
	if bytes.Contains(source, []byte("__ROBLOX_ART__")) {
		t.Fatal("unresolved image placeholder")
	}
}
