package main

import "testing"

func TestVersions(t *testing.T) {
	for _, v := range []struct {
		remote, current string
		want            bool
	}{{"v1.1.0", "1.0.0", true}, {"v1.0.0", "1.0.0", false}, {"v1.0.1", "1.0.9", false}, {"v2.0.0", "1.9.9", true}, {"oops", "1.0.0", false}} {
		if got := newerVersion(v.remote, v.current); got != v.want {
			t.Errorf("%s %s: %v", v.remote, v.current, got)
		}
	}
}
