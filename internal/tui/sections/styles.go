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

// Default palette. Used when content.Theme leaves a field empty.
const (
	defaultAccent     = "#7ee787"
	defaultAccent2    = "#d2a8ff"
	defaultForeground = "#e6edf3"
	defaultMuted      = "#7d8590"
	defaultBackground = "#0d1117"
	defaultSelection  = "#1f2a1f"
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

	Footer lipgloss.Style

	SplashTitle         lipgloss.Style
	SplashHint          lipgloss.Style
	SplashHintHighlight lipgloss.Style
}

// NewStyles builds the style set for the given theme. Empty theme fields
// fall back to the built-in palette.
func NewStyles(theme content.Theme) Styles {
	accent := lipgloss.Color(orDefault(theme.Accent, defaultAccent))
	accent2 := lipgloss.Color(orDefault(theme.Accent2, defaultAccent2))
	fg := lipgloss.Color(orDefault(theme.Foreground, defaultForeground))
	muted := lipgloss.Color(orDefault(theme.Muted, defaultMuted))
	bgDark := lipgloss.Color(orDefault(theme.Background, defaultBackground))
	bgSelect := lipgloss.Color(orDefault(theme.Selection, defaultSelection))

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
