// Package tui is the Bubble Tea program rendered to each SSH visitor.
//
// Visitors land on a splash screen. After Enter, they navigate a tabbed app
// whose tabs and contents are driven entirely by content.toml — the Model
// itself only owns the mode (splash vs. app), the active tab index,
// per-list selection indices, and the terminal size.
//
// Sub-views are pure render functions; they do not hold state.
package tui

import (
	"math"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ildar7070/ssh-cv/internal/content"
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

// Model is the root Bubble Tea model.
type Model struct {
	profile *content.Profile
	styles  Styles

	tabs []content.TabSpec // resolved at construction; only non-empty tabs

	mode      mode
	activeTab int // index into tabs

	// Per-list selection. Keyed by TabID so we don't hardcode field names.
	selection map[content.TabID]int

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
		profile:   p,
		styles:    NewStyles(r, p.Theme),
		tabs:      p.VisibleTabs(),
		mode:      modeSplash,
		selection: make(map[content.TabID]int),
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
	switch key := msg.String(); key {
	case "esc":
		return m, tea.Quit

	// Tab switching: Tab / Shift+Tab. Bounded — no wrap-around.
	case "tab", "right", "l":
		if m.activeTab < len(m.tabs)-1 {
			m.activeTab++
		}
	case "shift+tab", "left", "h":
		if m.activeTab > 0 {
			m.activeTab--
		}

	// List navigation on list-based tabs.
	case "up", "k":
		m.moveSelection(-1)
	case "down", "j":
		m.moveSelection(+1)
	case "home", "g":
		m.setSelection(0)
	case "end", "G":
		m.setSelection(math.MaxInt)

	default:
		// Digit shortcuts 1..9 jump to the Nth visible tab.
		if len(key) == 1 && key[0] >= '1' && key[0] <= '9' {
			idx := int(key[0] - '1')
			if idx < len(m.tabs) {
				m.activeTab = idx
			}
		}
	}
	return m, nil
}

func (m *Model) currentTab() content.TabID {
	if m.activeTab < 0 || m.activeTab >= len(m.tabs) {
		return ""
	}
	return m.tabs[m.activeTab].ID
}

func (m *Model) listLen() int {
	switch m.currentTab() {
	case content.TabCV:
		return len(m.profile.CV)
	case content.TabProjects:
		return len(m.profile.Projects)
	}
	return 0
}

func (m *Model) moveSelection(delta int) {
	id := m.currentTab()
	n := m.listLen()
	if n == 0 {
		return
	}
	m.selection[id] = clamp(m.selection[id]+delta, 0, n-1)
}

func (m *Model) setSelection(v int) {
	id := m.currentTab()
	n := m.listLen()
	if n == 0 {
		return
	}
	m.selection[id] = clamp(v, 0, n-1)
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
