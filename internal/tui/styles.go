package tui

import "github.com/charmbracelet/lipgloss"

// Styles holds every Lipgloss style used by the TUI. It must be built from a
// per-session *lipgloss.Renderer (see NewStyles) so the color profile reflects
// the SSH client's terminal capabilities. Using the global default renderer
// would inherit the server process's environment — in our case a distroless
// container with no TERM set, which silently strips all color.
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

// NewStyles builds the style set bound to the given renderer. Pass the
// renderer returned by bubbletea.MakeRenderer(sess) on each connect.
func NewStyles(r *lipgloss.Renderer) Styles {
	var (
		accent   = lipgloss.Color("#7ee787")
		accent2  = lipgloss.Color("#d2a8ff")
		fg       = lipgloss.Color("#e6edf3")
		muted    = lipgloss.Color("#7d8590")
		bgSelect = lipgloss.Color("#1f2a1f")
		bgDark   = lipgloss.Color("#0d1117")
	)

	ns := r.NewStyle

	return Styles{
		Doc: ns().Padding(1, 2),

		// Tabs are width-equalised and centred at render time (view.go computes
		// the longest label and applies .Width().Align(Center) on both styles).
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
