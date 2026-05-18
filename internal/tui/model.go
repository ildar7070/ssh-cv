// Package tui is the Bubble Tea program rendered to each SSH visitor.
//
// Layout:
//
//	┌─ splash ─────────────────────┐      ┌─ app ────────────────────────┐
//	│                              │      │ Start │ CV │ Projects │ ...  │
//	│           i12k               │ ───► ├──────────────────────────────┤
//	│   Press Enter to start       │      │ List       │ Detail          │
//	│                              │      │            │                 │
//	└──────────────────────────────┘      └──────────────────────────────┘
package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ildar7070/i12k/internal/content"
)

// blinkInterval controls how often the splash CTA toggles between its
// normal and highlighted style.
const blinkInterval = 1 * time.Second

// blinkMsg is emitted by the blink timer to drive the splash CTA animation.
type blinkMsg struct{}

func blinkCmd() tea.Cmd {
	return tea.Tick(blinkInterval, func(time.Time) tea.Msg { return blinkMsg{} })
}

type mode int

const (
	modeSplash mode = iota
	modeApp
)

// Tab identifiers. Order matches the header rendering and 1–4 shortcuts.
type tab int

const (
	tabStart tab = iota
	tabCV
	tabProjects
	tabContact
)

func (t tab) String() string {
	switch t {
	case tabStart:
		return "Start"
	case tabCV:
		return "CV"
	case tabProjects:
		return "Projects"
	case tabContact:
		return "Contact"
	}
	return ""
}

var allTabs = []tab{tabStart, tabCV, tabProjects, tabContact}

// Model is the root Bubble Tea model. It owns the current mode (splash vs.
// app), the active tab, the per-list selection index, and the terminal size.
// Sub-views are pure render functions — they don't hold their own state.
type Model struct {
	profile *content.Profile
	styles  Styles

	mode mode
	tab  tab

	cvIdx       int
	projectsIdx int

	// splashBlink toggles every blinkInterval while in modeSplash, driving
	// the "Press Enter to start" highlight pulse.
	splashBlink bool

	width  int
	height int
}

// New builds a Model. The renderer must come from
// bubbletea.MakeRenderer(sess) so color styling respects the SSH client's
// terminal — passing nil falls back to lipgloss.DefaultRenderer which in a
// distroless container reports no color support.
func New(p *content.Profile, r *lipgloss.Renderer) Model {
	if r == nil {
		r = lipgloss.DefaultRenderer()
	}
	return Model{
		profile: p,
		styles:  NewStyles(r),
		mode:    modeSplash,
		tab:     tabStart,
	}
}

func (m Model) Init() tea.Cmd { return blinkCmd() }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil

	case blinkMsg:
		// Keep the timer running so it can resume if the user ever
		// returns to splash (currently one-way, but cheap to keep alive).
		m.splashBlink = !m.splashBlink
		return m, blinkCmd()

	case tea.KeyMsg:
		// Quit works everywhere.
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}

		if m.mode == modeSplash {
			switch msg.String() {
			case "enter", " ":
				m.mode = modeApp
			}
			return m, nil
		}

		return m.updateApp(msg)
	}
	return m, nil
}

func (m Model) updateApp(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		return m, tea.Quit

	// Tab switching: Tab / Shift+Tab and digits 1–4.
	// Bounded — no wrap-around. Holding the key stops at the first/last tab.
	case "tab", "right", "l":
		if m.tab < tab(len(allTabs)-1) {
			m.tab++
		}
	case "shift+tab", "left", "h":
		if m.tab > 0 {
			m.tab--
		}
	case "1":
		m.tab = tabStart
	case "2":
		m.tab = tabCV
	case "3":
		m.tab = tabProjects
	case "4":
		m.tab = tabContact

	// List navigation on list-based tabs.
	case "up", "k":
		m.moveSelection(-1)
	case "down", "j":
		m.moveSelection(+1)
	case "home", "g":
		m.setSelection(0)
	case "end", "G":
		m.setSelection(1 << 30)
	}
	return m, nil
}

func (m *Model) listLen() int {
	switch m.tab {
	case tabCV:
		return len(m.profile.CV)
	case tabProjects:
		return len(m.profile.Projects)
	}
	return 0
}

func (m *Model) selectionPtr() *int {
	switch m.tab {
	case tabCV:
		return &m.cvIdx
	case tabProjects:
		return &m.projectsIdx
	}
	return nil
}

func (m *Model) moveSelection(delta int) {
	idx := m.selectionPtr()
	n := m.listLen()
	if idx == nil || n == 0 {
		return
	}
	*idx = clamp(*idx+delta, 0, n-1)
}

func (m *Model) setSelection(v int) {
	idx := m.selectionPtr()
	n := m.listLen()
	if idx == nil || n == 0 {
		return
	}
	*idx = clamp(v, 0, n-1)
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
