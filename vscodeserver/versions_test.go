// SPDX-License-Identifier: MIT
// © 2026 Hannes Stauss (scalarion@nimblescape.com)
// Licensed under the MIT License. See LICENSE in the repository root for details.

// Tests of the VS Code release source (versions.go) and of the declaration
// of the layer vscode-server, with example answers of the update service.

package vscodeserver

import (
	"reflect"
	"strings"
	"testing"

	"github.com/majikmate/devcontainer-core/pkg/layer"
)

func TestParseReleases(t *testing.T) {
	releases, err := parseReleases([]byte(`["1.105.1","1.105.0","1.104.3-insider","1.104.2"]`))
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, r := range releases {
		got = append(got, r.Version)
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

// TestResolve: without pin and channel, the general rule chooses the newest
// release.
func TestResolve(t *testing.T) {
	s := *vscodeSource
	s.Releases = func() ([]layer.Release, error) {
		return []layer.Release{{Version: "1.105.1"}, {Version: "1.105.0"}, {Version: "1.104.2"}}, nil
	}
	tool := layer.Tool{Name: "vscode-server", Source: &s}
	c, err := tool.Effective(layer.Config{Pin: vscodeServerPin, Channel: vscodeServerChannel})
	if err != nil {
		t.Fatal(err)
	}
	if v, err := tool.Resolve(c, ""); err != nil || v != "1.105.1" {
		t.Errorf("resolve = %q, %v, want 1.105.1", v, err)
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
	if _, err := parseBuild([]byte(strings.Replace(answer, commitID, "1.105.1", 1)), "1.105.1"); err == nil {
		t.Error("no commit: no error")
	}
}

// TestLayer: the layer is registered with the release choice constants.
func TestLayer(t *testing.T) {
	l, err := layer.Get("vscode-server")
	if err != nil {
		t.Fatal(err)
	}
	if len(l.Tools) != 1 || l.Tools[0].Arg != "VSCODE_SERVER_VERSION" || l.Tools[0].Source != vscodeSource ||
		l.Tools[0].Version != (layer.Config{Pin: vscodeServerPin, Channel: vscodeServerChannel}) {
		t.Errorf("tools = %+v", l.Tools)
	}
}
