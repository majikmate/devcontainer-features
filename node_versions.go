package features

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/majikmate/devcontainer-core/pkg/layer"
	"github.com/majikmate/devcontainer-core/pkg/sys"
	"github.com/majikmate/devcontainer-core/pkg/versions"
)

const (
	nodeIndexURL     = "https://nodejs.org/dist/index.json"
	nodeScheduleURL  = "https://raw.githubusercontent.com/nodejs/Release/main/schedule.json"
	nodeSchedulePage = "https://github.com/nodejs/Release#release-schedule"
)

// nodeSource lists the Node.js releases. Channels
// (https://nodejs.org/en/about/previous-releases): "lts" are the releases of
// an LTS line (Active or Maintenance LTS), "current" are all releases. A
// release line is a major release (for example 24); it ends on its end date
// in the Node.js release schedule.
var nodeSource = &layer.Source{
	Name: "Node.js releases (" + nodeIndexURL + ")",
	Channels: []layer.Channel{
		{Name: "lts", Label: "long-term support, the releases of an LTS line"},
		{Name: "current", Label: "all releases"},
	},
	Policy: "a line is a Node.js major release; it ends on its end date in the Node.js release schedule (" + nodeSchedulePage + ")",
	Releases: func() ([]layer.Release, error) {
		data, err := sys.Get(nodeIndexURL)
		if err != nil {
			return nil, err
		}
		return parseNodeReleases(data)
	},
	Support: func(line string) error {
		if !majorLine.MatchString(line) {
			return fmt.Errorf("Node.js line %s: a Node.js line is a major version, for example 24", line)
		}
		data, err := sys.Get(nodeScheduleURL)
		if err != nil {
			return err
		}
		return nodeSupport(data, line, time.Now().UTC())
	},
}

var majorLine = regexp.MustCompile(`^[0-9]+$`)

// parseNodeReleases reads the Node.js release list: every release is in the
// channel "current", a release with an LTS code name also in "lts". Highest
// version first.
func parseNodeReleases(data []byte) ([]layer.Release, error) {
	var list []struct {
		Version string `json:"version"`
		LTS     any    `json:"lts"`
	}
	if err := json.Unmarshal(data, &list); err != nil {
		return nil, err
	}
	var releases []layer.Release
	for _, r := range list {
		channels := []string{"current"}
		if lts, ok := r.LTS.(string); ok && lts != "" {
			channels = append(channels, "lts")
		}
		releases = append(releases, layer.Release{Version: r.Version, Channels: channels})
	}
	if len(releases) == 0 {
		return nil, fmt.Errorf("the Node.js release list is empty")
	}
	sort.SliceStable(releases, func(i, j int) bool { return versions.Compare(releases[i].Version, releases[j].Version) > 0 })
	return releases, nil
}

// nodeChannelName names the channel of an installed release from
// process.release.lts: the LTS code name, for example "LTS Krypton", or
// "Current".
func nodeChannelName(lts string) string {
	lts = strings.TrimSpace(lts)
	if lts == "" || lts == "undefined" {
		return "Current"
	}
	return "LTS " + lts
}

// nodeSupport returns an *layer.EndOfLifeError when the end date of the line
// in the release schedule has passed.
func nodeSupport(data []byte, line string, today time.Time) error {
	var schedule map[string]struct {
		Start string `json:"start"`
		End   string `json:"end"`
	}
	if err := json.Unmarshal(data, &schedule); err != nil {
		return err
	}
	date := func(s string) time.Time {
		t, _ := time.Parse("2006-01-02", s)
		return t
	}
	entry, ok := schedule["v"+line]
	if !ok || entry.End == "" {
		return fmt.Errorf("Node.js %s is not in the release schedule", line)
	}
	end := date(entry.End)
	if today.Before(end) {
		return nil
	}
	var supported []int
	for name, e := range schedule {
		major, err := strconv.Atoi(strings.TrimPrefix(name, "v"))
		if err != nil || e.Start == "" || e.End == "" {
			continue
		}
		if !date(e.Start).After(today) && today.Before(date(e.End)) {
			supported = append(supported, major)
		}
	}
	sort.Ints(supported)
	var names []string
	for _, major := range supported {
		names = append(names, strconv.Itoa(major))
	}
	return &layer.EndOfLifeError{
		Since:     "support ended on " + entry.End,
		Source:    nodeSchedulePage,
		Supported: names,
	}
}
