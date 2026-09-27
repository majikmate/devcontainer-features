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
)

// Node.js release lines: a line is a major release (for example 24). It ends
// on its end date in the release schedule of the Node.js project.
var nodePin = &layer.Pin{
	Arg:     "NODE_PIN",
	Example: "24",
	Policy:  "a line is a Node.js major release; it ends on its end date in the Node.js release schedule (" + nodeSchedulePage + ")",
	Newest: func(line string) (string, error) {
		if !majorLine.MatchString(line) {
			return "", fmt.Errorf("NODE_PIN=%s: a Node.js line is a major version, for example 24", line)
		}
		data, err := sys.Get("https://nodejs.org/dist/index.json")
		if err != nil {
			return "", err
		}
		return nodeNewestIn(data, line)
	},
	Support: func(line string) error {
		if !majorLine.MatchString(line) {
			return fmt.Errorf("NODE_PIN=%s: a Node.js line is a major version, for example 24", line)
		}
		data, err := sys.Get(nodeScheduleURL)
		if err != nil {
			return err
		}
		return nodeSupport(data, line, time.Now().UTC())
	},
}

const (
	nodeScheduleURL  = "https://raw.githubusercontent.com/nodejs/Release/main/schedule.json"
	nodeSchedulePage = "https://github.com/nodejs/Release#release-schedule"
)

var majorLine = regexp.MustCompile(`^[0-9]+$`)

// nodeNewestIn returns the newest release of a major line from the Node.js
// release list (newest first), for example "v24.9.0" for "24".
func nodeNewestIn(data []byte, line string) (string, error) {
	var releases []struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(data, &releases); err != nil {
		return "", err
	}
	for _, r := range releases {
		if layer.InLine(r.Version, line) {
			return r.Version, nil
		}
	}
	return "", fmt.Errorf("Node.js %s has no release", line)
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
