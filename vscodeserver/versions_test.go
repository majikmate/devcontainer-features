// SPDX-License-Identifier: MIT
// © 2026 Hannes Stauss (scalarion@nimblescape.com)
// Licensed under the MIT License. See LICENSE in the repository root for details.

// Tests of the release list, the choice of the releases and the answers of
// the update service (versions.go), with example answers.

package vscodeserver

import (
	"reflect"
	"strings"
	"testing"
)

func TestParseReleases(t *testing.T) {
	got, err := parseReleases([]byte(`["1.105.1","1.105.0","1.104.3-insider","1.104.2"]`))
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"1.105.1", "1.105.0", "1.104.2"}; !reflect.DeepEqual(got, want) {
		t.Errorf("releases = %v, want %v", got, want)
	}
	if _, err := parseReleases([]byte(`[]`)); err == nil {
		t.Error("empty list: no error")
	}
	if _, err := parseReleases([]byte(`{`)); err == nil {
		t.Error("broken JSON: no error")
	}
}

// TestNewestPerMinor: the newest patch release of each minor version, at
// most keep releases.
func TestNewestPerMinor(t *testing.T) {
	releases := []string{"1.139.1", "1.139.0", "1.138.2", "1.138.1", "1.138.0", "1.137.0", "1.136.3"}
	if got, want := newestPerMinor(releases, 3), []string{"1.139.1", "1.138.2", "1.137.0"}; !reflect.DeepEqual(got, want) {
		t.Errorf("keep 3 = %v, want %v", got, want)
	}
	if got, want := newestPerMinor(releases, 1), []string{"1.139.1"}; !reflect.DeepEqual(got, want) {
		t.Errorf("keep 1 = %v, want %v", got, want)
	}
	if got := newestPerMinor(releases[:2], 3); !reflect.DeepEqual(got, []string{"1.139.1"}) {
		t.Errorf("one minor version = %v", got)
	}
}

func TestParseBuild(t *testing.T) {
	sha := strings.Repeat("ab", 32)
	commitID := strings.Repeat("0f", 20)
	answer := `{"url":"https://example.com/vscode-server-linux-x64.tar.gz","name":"1.105.1","version":"` + commitID +
		`","productVersion":"1.105.1","hash":"x","timestamp":1,"sha256hash":"` + sha + `","supportsFastUpdate":true}`
	b, err := parseBuild([]byte(answer), "1.105.1")
	if err != nil {
		t.Fatal(err)
	}
	want := build{URL: "https://example.com/vscode-server-linux-x64.tar.gz", SHA256: sha, Commit: commitID, ProductVersion: "1.105.1"}
	if b != want {
		t.Errorf("build = %+v, want %+v", b, want)
	}
	// Another release, no checksum or no commit: errors
	if _, err := parseBuild([]byte(answer), "1.105.0"); err == nil {
		t.Error("other release: no error")
	}
	if _, err := parseBuild([]byte(strings.Replace(answer, sha, "", 1)), "1.105.1"); err == nil {
		t.Error("no checksum: no error")
	}
	// The commit becomes a folder name: a path is not a commit
	if _, err := parseBuild([]byte(strings.Replace(answer, commitID, "../../etc", 1)), "1.105.1"); err == nil {
		t.Error("path as commit: no error")
	}
}
