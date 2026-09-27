package features

import (
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/majikmate/devcontainer-core/pkg/layer"
)

// Excerpt of https://go.dev/dl/?mode=json&include=all
const goReleases = `[
 {"version": "go1.29rc1", "stable": false},
 {"version": "go1.28.2", "stable": true},
 {"version": "go1.28.0", "stable": true},
 {"version": "go1.27.10", "stable": true},
 {"version": "go1.27.9", "stable": true},
 {"version": "go1.27.0", "stable": true},
 {"version": "go1.26.12", "stable": true}
]`

func TestGoPin(t *testing.T) {
	releases, err := parseGoReleases([]byte(goReleases))
	if err != nil {
		t.Fatal(err)
	}
	if v, err := goNewestIn(releases, "1.27"); err != nil || v != "1.27.10" {
		t.Errorf("newest 1.27 = %q, %v; want 1.27.10", v, err)
	}
	if _, err := goNewestIn(releases, "1.29"); err == nil {
		t.Error("1.29 (only a release candidate): no error")
	}
	if _, err := goNewestIn(releases, "1"); err == nil {
		t.Error("line 1: no error")
	}
	for _, line := range []string{"1.27", "1.28"} {
		if err := goSupport(releases, line); err != nil {
			t.Errorf("support of %s: %v", line, err)
		}
	}
	if err := goSupport(releases, "1.30"); err == nil || errors.As(err, new(*layer.EndOfLifeError)) {
		t.Errorf("1.30 (not released): err = %v", err)
	}

	err = goSupport(releases, "1.26")
	var eol *layer.EndOfLifeError
	if !errors.As(err, &eol) {
		t.Fatalf("1.26: err = %v, want an EndOfLifeError", err)
	}
	if !strings.Contains(eol.Since, "Go 1.28.0 was released") || !reflect.DeepEqual(eol.Supported, []string{"1.27", "1.28"}) {
		t.Errorf("1.26: %+v", eol)
	}
	// The message of the framework
	tool := layer.Tool{Name: "go", Pin: &layer.Pin{Arg: "GO_PIN", Support: func(line string) error { return goSupport(releases, line) }}}
	want := "go 1.26 (GO_PIN=1.26) has reached its end of life (Go 1.28.0 was released; Go supports the two newest major releases). Source: https://go.dev/doc/devel/release#policy. Change ARG GO_PIN in the Dockerfile to a supported version (supported: 1.27, 1.28)."
	if err := tool.CheckPin("1.26"); err == nil || err.Error() != want {
		t.Errorf("message:\n got %v\nwant %s", err, want)
	}
}

func TestGoModuleVersions(t *testing.T) {
	list := "v0.19.1\nv0.20.0-pre.1\nv0.20.0\nv0.9.10\nv0.19.0\n"
	if got, want := releaseVersions(list), []string{"v0.20.0", "v0.19.1", "v0.19.0", "v0.9.10"}; !reflect.DeepEqual(got, want) {
		t.Errorf("versions = %v, want %v", got, want)
	}
	mod := "module golang.org/x/tools/gopls\n\ngo 1.28.0\n\ntoolchain go1.28.2\n"
	if got := goDirective(mod); got != "1.28.0" {
		t.Errorf("go directive = %q", got)
	}
	if got := goDirective("module x\n"); got != "" {
		t.Errorf("no go directive = %q", got)
	}
	cases := []struct {
		a, b string
		want int
	}{
		{"1.28.0", "1.27.10", 1},
		{"1.27", "1.27.10", -1},
		{"1.27.0", "1.27", 0},
		{"", "1.27.3", -1},
		{"go1.27.3", "v1.27.3", 0},
	}
	for _, c := range cases {
		if got := compareVersions(c.a, c.b); got != c.want {
			t.Errorf("compareVersions(%q, %q) = %d, want %d", c.a, c.b, got, c.want)
		}
	}
}

// Excerpts of https://nodejs.org/dist/index.json and of the Node.js release schedule
const (
	nodeReleases = `[{"version":"v26.1.0"},{"version":"v24.11.2"},{"version":"v24.11.1"},{"version":"v22.21.0"}]`
	nodeSchedule = `{
 "v20": {"start": "2023-04-18", "lts": "2023-10-24", "maintenance": "2024-10-22", "end": "2026-04-30", "codename": "Iron"},
 "v22": {"start": "2024-04-24", "lts": "2024-10-29", "maintenance": "2025-10-21", "end": "2027-04-30", "codename": "Jod"},
 "v24": {"start": "2025-05-06", "lts": "2025-10-28", "maintenance": "2026-10-20", "end": "2028-04-30", "codename": "Krypton"},
 "v25": {"start": "2025-10-15", "maintenance": "2026-04-01", "end": "2026-06-01"},
 "v26": {"start": "2026-04-22", "lts": "2026-10-28", "maintenance": "2027-10-20", "end": "2029-04-30", "codename": ""}
}`
)

func TestNodePin(t *testing.T) {
	if v, err := nodeNewestIn([]byte(nodeReleases), "24"); err != nil || v != "v24.11.2" {
		t.Errorf("newest 24 = %q, %v; want v24.11.2", v, err)
	}
	if v, err := nodeNewestIn([]byte(nodeReleases), "2"); err == nil {
		t.Errorf("line 2 = %q, want an error", v)
	}
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

func TestDenoPin(t *testing.T) {
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

	tags := "abc\trefs/tags/v2.9.1\nabc\trefs/tags/v2.10.0\nabc\trefs/tags/v2.10.0-rc.1\nabc\trefs/tags/v20.0.0\n"
	if v, err := highestTag(tags, "2"); err != nil || v != "v2.10.0" {
		t.Errorf("highest 2 = %q, %v; want v2.10.0", v, err)
	}
}

func TestPinsDeclared(t *testing.T) {
	for name, arg := range map[string]string{"go": "GO_PIN", "node": "NODE_PIN", "deno": "DENO_PIN"} {
		l, err := layer.Get(name)
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for _, tool := range l.Tools {
			if tool.Pin != nil {
				found = tool.Pin.Arg == arg && tool.Pin.Newest != nil && tool.Pin.Support != nil && tool.Pin.Example != "" && tool.Pin.Policy != ""
			}
		}
		if !found {
			t.Errorf("layer %s: no complete pin %s", name, arg)
		}
	}
	// The Go tools built with "go install" follow Go
	l, _ := layer.Get("go")
	for _, tool := range l.Tools {
		for _, g := range goTools {
			if tool.Name == g.name && (tool.Follows != "go" || tool.NewestFor == nil) {
				t.Errorf("tool %s does not follow go", tool.Name)
			}
		}
	}
}
