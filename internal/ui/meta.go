package ui

// meta.go holds the Go-side data for the small self-describing readouts
// (clock, coordinates, uptime, build date) shown in meta.templ. Kept in its
// own file because these are process-wide, unlike the rest of page.go which
// is per-request.

import (
	"fmt"
	"time"

	"github.com/mattbx/mattbx-go/internal/build"
)

// SiteTimeZone is the operator's own clock — the one the site shows,
// regardless of the visitor's location.
const SiteTimeZone = "Australia/Sydney"

// SiteCoordinates is a rounded, city-level lon/lat, not a precise location.
const SiteCoordinates = "-33.87, 151.21"

// processStart is set once, at binary startup, for the uptime readout.
var processStart = time.Now()

// siteLocation loads SiteTimeZone once. A bad IANA name is a coding error
// that should fail immediately in development, not something to guard
// against on every request.
var siteLocation = func() *time.Location {
	loc, err := time.LoadLocation(SiteTimeZone)
	if err != nil {
		panic("ui: invalid SiteTimeZone: " + err.Error())
	}
	return loc
}()

// localTime is the clock's server-rendered starting value; terminal.js ticks
// it from there using the same time zone (data-tz).
func localTime(t time.Time) string { return t.In(siteLocation).Format("15:04:05") }

// localDateTime is the machine-readable value for the <time> element.
func localDateTime(t time.Time) string { return t.In(siteLocation).Format(time.RFC3339) }

// uptimeLabel formats how long this process has been running, matching the
// "Nd HH:MM:SS" shape terminal.js's formatDuration reproduces client-side.
func uptimeLabel() string {
	d := time.Since(processStart)
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60
	return fmt.Sprintf("%dd %02d:%02d:%02d", days, hours, minutes, seconds)
}

// processStartUnix feeds terminal.js a start time so it can tick the uptime
// readout between page loads without polling the server.
func processStartUnix() string { return fmt.Sprintf("%d", processStart.Unix()) }

// buildLabel reports when this image was built. See internal/build and the
// Dockerfile's ldflags.
func buildLabel() string { return build.Date }

// offsetZone is one entry in the global-offsets readout.
type offsetZone struct {
	abbr string
	tz   string
}

// offsetZones is the fixed set shown in the colophon. Add or remove entries
// here rather than in the template.
var offsetZones = []offsetZone{
	{"UTC", "UTC"},
	{"EST", "America/New_York"},
	{"CST", "America/Chicago"},
	{"PST", "America/Los_Angeles"},
}

// offsetLabels computes each zone's current UTC offset live, so it stays
// correct across daylight saving changes rather than going stale like a
// hardcoded string would.
func offsetLabels() []string {
	now := time.Now()
	labels := make([]string, 0, len(offsetZones))
	for _, z := range offsetZones {
		loc, err := time.LoadLocation(z.tz)
		if err != nil {
			continue // misconfigured entry — skip rather than break the page
		}
		_, offset := now.In(loc).Zone()
		hours := offset / 3600
		sign := "+"
		if hours < 0 {
			sign = "-"
			hours = -hours
		}
		labels = append(labels, fmt.Sprintf("%s%s%d", z.abbr, sign, hours))
	}
	return labels
}
