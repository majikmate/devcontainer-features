// Tests of the Deno release source (versions.go) and of the declaration of
// the layer deno, with example release tags.

package deno

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/majikmate/devcontainer-core/pkg/layer"
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
		{layer.Config{Pin: "2", Channel: "lts"}, "v2.9.3"}, // line 2, channel lts: exactly the LTS release
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

// TestLayer checks that the layer uses the release choice of the constants
// and switches off the Deno upgrade message.
func TestLayer(t *testing.T) {
	l, err := layer.Get("deno")
	if err != nil {
		t.Fatal(err)
	}
	if want := (layer.Config{Pin: denoPin, Channel: denoChannel}); len(l.Tools) != 1 || l.Tools[0].Version != want {
		t.Errorf("tools %+v, want deno with %+v", l.Tools, want)
	}
	if env, _ := l.Metadata["containerEnv"].(map[string]any); env["DENO_NO_UPDATE_CHECK"] == nil {
		t.Error("containerEnv DENO_NO_UPDATE_CHECK is missing")
	}
}
