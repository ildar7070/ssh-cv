// Package sections owns the pluggable per-tab renderers used by the TUI.
//
// Each renderer is a Renderer implementation registered in its own
// init(); the TUI's bodyView looks the right one up by Section.Type and
// delegates rendering, key handling, and emptiness checks to it. New
// section types are added by dropping a new file in this package — the
// core TUI does not need to change.
package sections

import (
	"charm.land/lipgloss/v2"

	"github.com/ildar7070/ssh-cv/internal/content"
)

// Built-in dark palette. Fills in empty fields of the dark Palette.
const (
	darkAccent     = "#7ee787"
	darkAccent2    = "#d2a8ff"
	darkForeground = "#e6edf3"
	darkMuted      = "#7d8590"
	darkBackground = "#0d1117"
	darkSelection  = "#1f2a1f"
)

// Built-in light palette. Fills in empty fields of the light Palette, i.e.
// when content.toml has no [theme.light] (or omits individual keys).
const (
	lightAccent     = "#1a7f37"
	lightAccent2    = "#8250df"
	lightForeground = "#1f2328"
	lightMuted      = "#59636e"
	lightBackground = "#ffffff"
	lightSelection  = "#ddf4e0"
)

// DarkDefaults and LightDefaults are the built-in palettes used to fill in
// empty fields of a user-supplied Palette.
var (
	DarkDefaults = content.Palette{
		Accent:     darkAccent,
		Accent2:    darkAccent2,
		Foreground: darkForeground,
		Muted:      darkMuted,
		Background: darkBackground,
		Selection:  darkSelection,
	}
	LightDefaults = content.Palette{
		Accent:     lightAccent,
		Accent2:    lightAccent2,
		Foreground: lightForeground,
		Muted:      lightMuted,
		Background: lightBackground,
		Selection:  lightSelection,
	}
)

// Styles holds every Lipgloss style used by the TUI. Styles carry no color
// profile of their own; under Bubble Tea v2 the program downsamples colors to
// the client terminal's capabilities at render time.
type Styles struct {
	Doc lipgloss.Style

	Tab       lipgloss.Style
	TabActive lipgloss.Style
	TabsSep   string

	ListItem         lipgloss.Style
	ListItemSelected lipgloss.Style

	DetailHeading lipgloss.Style
	DetailSub     lipgloss.Style
	DetailMuted   lipgloss.Style
	DetailBody    lipgloss.Style
	DetailLink    lipgloss.Style

	// Ascii styles the optional ASCII-art block in text sections.
	Ascii lipgloss.Style

	Footer lipgloss.Style

	SplashTitle         lipgloss.Style
	SplashHint          lipgloss.Style
	SplashHintHighlight lipgloss.Style
}

// NewStyles builds a style set from a palette. Empty palette fields fall
// back to the matching field in defaults (DarkDefaults or LightDefaults).
func NewStyles(pal, defaults content.Palette) Styles {
	accent := lipgloss.Color(orDefault(pal.Accent, defaults.Accent))
	accent2 := lipgloss.Color(orDefault(pal.Accent2, defaults.Accent2))
	fg := lipgloss.Color(orDefault(pal.Foreground, defaults.Foreground))
	muted := lipgloss.Color(orDefault(pal.Muted, defaults.Muted))
	bgDark := lipgloss.Color(orDefault(pal.Background, defaults.Background))
	bgSelect := lipgloss.Color(orDefault(pal.Selection, defaults.Selection))

	ns := lipgloss.NewStyle

	return Styles{
		Doc: ns().Padding(1, 2),

		Tab:       ns().Foreground(muted),
		TabActive: ns().Foreground(bgDark).Background(accent).Bold(true),
		TabsSep:   ns().Foreground(muted).Render(" │ "),

		ListItem:         ns().Padding(0, 1).Foreground(fg),
		ListItemSelected: ns().Padding(0, 1).Foreground(accent).Background(bgSelect).Bold(true),

		DetailHeading: ns().Foreground(accent).Bold(true),
		DetailSub:     ns().Foreground(accent2),
		DetailMuted:   ns().Foreground(muted),
		DetailBody:    ns().Foreground(fg),
		// DetailLink is rendered without padding/background so Apple Terminal's
		// URL detector can match the whole word in one go — Cmd-click then
		// opens it on a single click instead of requiring a double-click.
		DetailLink: ns().Foreground(fg).Underline(true),

		Ascii: ns().Foreground(accent2),

		Footer: ns().Foreground(muted).Padding(1, 0, 0, 0),

		SplashTitle:         ns().Foreground(accent).Bold(true),
		SplashHint:          ns().Foreground(muted),
		SplashHintHighlight: ns().Foreground(bgDark).Background(accent).Bold(true),
	}
}

func orDefault(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
