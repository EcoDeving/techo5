//go:build !dot && !spot

package display

import (
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/HuskerMinion/techo5/echod/internal/config"
	"github.com/HuskerMinion/techo5/echod/internal/feature/home"
	"github.com/HuskerMinion/techo5/echod/internal/lib/hass"
)

// The Dashboard style with a week of events, and with none, draws in every screen language on both
// panels; with SHOW_PREVIEW set, each is written there (dashboard-week-es-show8.png and so on) to look at.
func TestDashboardWeekDraws(t *testing.T) {
	config.Use(filepath.Join(t.TempDir(), "state.json"))
	if err := config.Set().Screen().ClockStyle(styleDashboard); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = config.Set().Screen().ClockStyle("") })
	clock24.Store(true)
	t.Cleanup(func() { clock24.Store(false) })
	now := time.Date(2026, 10, 8, 18, 42, 0, 0, time.Local)
	day := func(d, h, m int) time.Time { return time.Date(2026, 10, d, h, m, 0, 0, time.Local) }
	var week []hass.Day
	for i, c := range []string{"partlycloudy", "sunny", "rainy", "cloudy", "sunny"} {
		week = append(week, hass.Day{When: now.AddDate(0, 0, i), Condition: c, High: float64(31 - i), Low: float64(21 - i)})
	}
	s := scene{now: now, phase: "idle", weather: home.Weather{Condition: "partlycloudy", Temp: "27°"}, style: styleFacts{
		days: week,
		next: []hass.Event{
			{Summary: "Dinner", Start: day(8, 19, 30), End: day(8, 21, 0)},
			{Summary: "Team meeting", Start: day(9, 10, 0), End: day(9, 11, 0)},
			{Summary: "Family barbecue", Start: day(11, 0, 0), End: day(12, 0, 0), AllDay: true},
			{Summary: "Dentist", Start: day(13, 16, 0), End: day(13, 17, 0)},
			{Summary: "Birthday party", Start: day(14, 0, 0), End: day(15, 0, 0), AllDay: true},
			{Summary: "Piano lesson", Start: day(15, 17, 30), End: day(15, 18, 30)},
			{Summary: "Book club", Start: day(16, 20, 0), End: day(16, 22, 0)},
			{Summary: "Car service", Start: day(19, 9, 0), End: day(19, 10, 0)},
		},
	}}
	empty := s
	empty.style.next = nil
	t.Cleanup(func() { _ = config.Set().Screen().Language("") })
	for _, lang := range []string{"", "de", "es", "fr", "it", "nl", "ru"} {
		if err := config.Set().Screen().Language(lang); err != nil {
			t.Fatal(err)
		}
		for _, p := range []struct {
			name       string
			scene      scene
			wide, high int
		}{{"week", s, showWide, showHigh}, {"week", s, show8Wide, show8High}, {"empty", empty, showWide, showHigh}} {
			img := image.NewRGBA(image.Rect(0, 0, p.wide, p.high))
			newRenderer(img).draw(p.scene)
			dir := os.Getenv("SHOW_PREVIEW")
			if dir == "" {
				continue
			}
			name := "dashboard-" + p.name
			if lang != "" {
				name += "-" + lang
			}
			if p.wide == show8Wide {
				name += "-show8"
			}
			f, err := os.Create(filepath.Join(dir, name+".png"))
			if err != nil {
				t.Fatal(err)
			}
			if err := png.Encode(f, img); err != nil {
				t.Fatal(err)
			}
			f.Close()
		}
	}
}
