// SPDX-License-Identifier: MIT
// © 2026 Hannes Stauss (scalarion@nimblescape.com)
// Licensed under the MIT License. See LICENSE in the repository root for details.

// Tests of the Node.js release source (versions.go) and of the declaration of
// the layer node, with excerpts of the Node.js release index and schedule.

package node

import (
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/majikmate/devcontainer-core/pkg/layer"
)

// Excerpts of https://nodejs.org/dist/index.json and of the Node.js release schedule
const (
	nodeReleases = `[
 {"version": "v26.1.0", "lts": false},
 {"version": "v24.11.2", "lts": "Krypton"},
 {"version": "v24.11.1", "lts": "Krypton"},
 {"version": "v24.9.0", "lts": false},
 {"version": "v22.21.0", "lts": "Jod"}
]`
	nodeSchedule = `{
 "v20": {"start": "2023-04-18", "lts": "2023-10-24", "maintenance": "2024-10-22", "end": "2026-04-30", "codename": "Iron"},
 "v22": {"start": "2024-04-24", "lts": "2024-10-29", "maintenance": "2025-10-21", "end": "2027-04-30", "codename": "Jod"},
 "v24": {"start": "2025-05-06", "lts": "2025-10-28", "maintenance": "2026-10-20", "end": "2028-04-30", "codename": "Krypton"},
 "v25": {"start": "2025-10-15", "maintenance": "2026-04-01", "end": "2026-06-01"},
 "v26": {"start": "2026-04-22", "lts": "2026-10-28", "maintenance": "2027-10-20", "end": "2029-04-30", "codename": ""}
}`
)

// resolve returns the version that the general rule chooses from releases.
func resolve(t *testing.T, source *layer.Source, releases []layer.Release, config layer.Config) string {
	t.Helper()
	s := *source
	s.Releases = func() ([]layer.Release, error) { return releases, nil }
	tool := layer.Tool{Name: "test", Source: &s}
	c, err := tool.Effective(config)
	if err != nil {
		t.Fatal(err)
	}
	v, err := tool.Resolve(c, "")
	if err != nil {
		return "error: " + err.Error()
	}
	return v
}

func TestNodeSource(t *testing.T) {
	releases, err := parseNodeReleases([]byte(nodeReleases))
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		config layer.Config
		want   string
	}{
		{layer.Config{Pin: "24", Channel: "lts"}, "v24.11.2"}, // line 24, channel lts
		{layer.Config{}, "v24.11.2"},                          // default channel lts
		{layer.Config{Channel: "current"}, "v26.1.0"},         // the newest release
		{layer.Config{Pin: "22"}, "v22.21.0"},
	}
	for _, c := range cases {
		if got := resolve(t, nodeSource, releases, c.config); got != c.want {
			t.Errorf("%+v: %s, want %s", c.config, got, c.want)
		}
	}
	// No LTS release of line 26 yet
	if got := resolve(t, nodeSource, releases, layer.Config{Pin: "26", Channel: "lts"}); !strings.HasPrefix(got, "error:") {
		t.Errorf("26, lts: %s, want an error", got)
	}
	for lts, want := range map[string]string{"Krypton\n": "LTS Krypton", "": "Current", "undefined": "Current"} {
		if got := nodeChannelName(lts); got != want {
			t.Errorf("nodeChannelName(%q) = %q, want %q", lts, got, want)
		}
	}
}

func TestNodeSupport(t *testing.T) {
	today, _ := time.Parse("2006-01-02", "2026-09-27")
	if err := nodeSupport([]byte(nodeSchedule), "24", today); err != nil {
		t.Errorf("24: %v", err)
	}
	if err := nodeSupport([]byte(nodeSchedule), "19", today); err == nil {
		t.Error("19 (not in the schedule): no error")
	}
	err := nodeSupport([]byte(nodeSchedule), "20", today)
	var eol *layer.EndOfLifeError
	if !errors.As(err, &eol) {
		t.Fatalf("20: err = %v, want an EndOfLifeError", err)
	}
	if eol.Since != "support ended on 2026-04-30" || !reflect.DeepEqual(eol.Supported, []string{"22", "24", "26"}) {
		t.Errorf("20: %+v", eol)
	}
}

// TestLayer checks that the layer uses the release choice of the constants
// and switches off the npm upgrade message.
func TestLayer(t *testing.T) {
	l, err := layer.Get("node")
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]layer.Config{
		"node": {Pin: nodePin, Channel: nodeChannel},
		"nvm":  {Pin: nvmPin, Channel: nvmChannel},
	}
	for _, tool := range l.Tools {
		if tool.Version != want[tool.Name] {
			t.Errorf("tool %s: configuration %+v, want %+v", tool.Name, tool.Version, want[tool.Name])
		}
	}
	if env, _ := l.Metadata["containerEnv"].(map[string]any); env["NPM_CONFIG_UPDATE_NOTIFIER"] == nil {
		t.Error("containerEnv NPM_CONFIG_UPDATE_NOTIFIER is missing")
	}
}
