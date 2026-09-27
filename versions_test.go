package features

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
		{layer.Config{Pin: "24", Channel: "lts"}, "v24.11.2"}, // the configuration of the feature
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

func TestDenoSource(t *testing.T) {
	tags := "abc\trefs/tags/v2.9.3\nabc\trefs/tags/v2.9.7\nabc\trefs/tags/v2.9.8\nabc\trefs/tags/v2.10.0-rc.1\nabc\trefs/tags/v1.46.3\nabc\trefs/tags/v20.0.0\n"
	// v2.9.8 is tagged, but not released on dl.deno.land yet
	releases := parseDenoReleases(tags, "v2.9.7", "v2.9.3")
	var got []string
	for _, r := range releases {
		got = append(got, r.Version+" "+strings.Join(r.Channels, ","))
	}
	if want := []string{"v2.9.7 stable", "v2.9.3 stable,lts", "v1.46.3 stable"}; !reflect.DeepEqual(got, want) {
		t.Errorf("releases = %v, want %v", got, want)
	}
	cases := []struct {
		config layer.Config
		want   string
	}{
		{layer.Config{Pin: "2", Channel: "lts"}, "v2.9.3"}, // the configuration of the feature: exactly the LTS release
		{layer.Config{Pin: "2", Channel: "stable"}, "v2.9.7"},
		{layer.Config{Pin: "1", Channel: "stable"}, "v1.46.3"},
	}
	for _, c := range cases {
		if got := resolve(t, denoSource, releases, c.config); got != c.want {
			t.Errorf("%+v: %s, want %s", c.config, got, c.want)
		}
	}
}

func TestDenoSupport(t *testing.T) {
	if err := denoSupport("v2.5.6", "2"); err != nil {
		t.Errorf("2 with v2.5.6: %v", err)
	}
	if err := denoSupport("v2.5.6", "3"); err == nil || errors.As(err, new(*layer.EndOfLifeError)) {
		t.Errorf("3 (not released): err = %v", err)
	}
	err := denoSupport("v3.0.1", "2")
	var eol *layer.EndOfLifeError
	if !errors.As(err, &eol) {
		t.Fatalf("2 with v3.0.1: err = %v, want an EndOfLifeError", err)
	}
	if !strings.Contains(eol.Since, "Deno v3.0.1 was released") || !reflect.DeepEqual(eol.Supported, []string{"3"}) {
		t.Errorf("2 with v3.0.1: %+v", eol)
	}
}

func TestGoDirective(t *testing.T) {
	mod := "module golang.org/x/tools/gopls\n\ngo 1.28.0\n\ntoolchain go1.28.2\n"
	if got := goDirective(mod); got != "1.28.0" {
		t.Errorf("go directive = %q", got)
	}
	if got := goDirective("module x\n"); got != "" {
		t.Errorf("no go directive = %q", got)
	}
}

// TestVersionsDeclared checks the release choices of the features: Go 1.27,
// Node.js 24 LTS, Deno 2 LTS; all other tools the newest release. Every tool
// has a source; the Go tools follow Go.
func TestVersionsDeclared(t *testing.T) {
	want := map[string]layer.Config{
		"go":   {Pin: "1.27"},
		"node": {Pin: "24", Channel: "lts"},
		"deno": {Pin: "2", Channel: "lts"},
	}
	for _, name := range names {
		l, err := layer.Get(name)
		if err != nil {
			t.Fatal(err)
		}
		for _, tool := range l.Tools {
			if tool.Source == nil || tool.Source.Releases == nil || tool.Source.Name == "" {
				t.Errorf("tool %s: no complete source", tool.Name)
			}
			if tool.Version != want[tool.Name] {
				t.Errorf("tool %s: configuration %+v, want %+v", tool.Name, tool.Version, want[tool.Name])
			}
			if _, err := tool.Effective(tool.Version); err != nil {
				t.Errorf("tool %s: %v", tool.Name, err)
			}
			if tool.Version.Pin != "" && tool.Source.Support == nil {
				t.Errorf("tool %s: pinned without an end-of-life rule", tool.Name)
			}
		}
	}
	l, _ := layer.Get("go")
	for _, tool := range l.Tools {
		for _, g := range goTools {
			if tool.Name == g.name && (tool.Follows != "go" || tool.Works == nil) {
				t.Errorf("tool %s does not follow go", tool.Name)
			}
		}
	}
	// No upgrade messages of Deno and npm
	for name, variable := range map[string]string{"deno": "DENO_NO_UPDATE_CHECK", "node": "NPM_CONFIG_UPDATE_NOTIFIER"} {
		l, _ := layer.Get(name)
		if env, _ := l.Metadata["containerEnv"].(map[string]any); env[variable] == nil {
			t.Errorf("layer %s: containerEnv %s is missing", name, variable)
		}
	}
}
