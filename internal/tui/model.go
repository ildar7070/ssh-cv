// Package tui is the Bubble Tea program rendered to each SSH visitor.
//
// Visitors land on a splash screen. After Enter, they navigate a tabbed app
// whose tabs and contents are driven entirely by content.toml — the Model
// itself only owns the mode (splash vs. app), the active tab index,
// per-section selection indices, and the terminal size. Per-section
// rendering, key handling, and emptiness checks live in
// internal/tui/sections.
//
// Sub-views are pure render functions; they do not hold state.
package tui

import (
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/ildar7070/ssh-cv/internal/content"
	"github.com/ildar7070/ssh-cv/internal/tui/sections"
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
	styles  sections.Styles

	tabs []content.Section // resolved at construction; only non-empty sections

	mode      mode
	activeTab int // index into tabs

	// Per-section selection. Keyed by Section.ID so we don't hardcode field names.
	selection map[string]int

	// splashBlink toggles every blinkInterval while in modeSplash, driving
	// the "Press Enter to start" highlight pulse.
	splashBlink bool

	width  int
	height int
}

// New builds a Model. Under Bubble Tea v2 the program downsamples colors to
// the SSH client's terminal at render time, so styles need no per-session
// renderer.
func New(p *content.Profile) Model {
	return Model{
		profile:   p,
		styles:    sections.NewStyles(p.Theme),
		tabs:      p.VisibleSections(),
		mode:      modeSplash,
		selection: make(map[string]int),
	}
}

func (m Model) Init() tea.Cmd { return blinkCmd() }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil

	case blinkMsg:
		if m.mode != modeSplash {
			return m, nil
		}
		m.splashBlink = !m.splashBlink
		return m, blinkCmd()

	case tea.KeyPressMsg:
		// Quit works everywhere.
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}

		if m.mode == modeSplash {
			switch msg.String() {
			case "enter", "space":
				m.mode = modeApp
			}
			return m, nil
		}

		return m.updateApp(msg)
	}
	return m, nil
}

func (m Model) updateApp(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	switch key {
	case "esc":
		return m, tea.Quit

	// Tab switching: Tab / Shift+Tab. Bounded — no wrap-around.
	case "tab", "right", "l":
		if m.activeTab < len(m.tabs)-1 {
			m.activeTab++
		}
		return m, nil
	case "shift+tab", "left", "h":
		if m.activeTab > 0 {
			m.activeTab--
		}
		return m, nil
	}

	// Digit shortcuts 1..9 jump to the Nth visible tab.
	if len(key) == 1 && key[0] >= '1' && key[0] <= '9' {
		idx := int(key[0] - '1')
		if idx < len(m.tabs) {
			m.activeTab = idx
		}
		return m, nil
	}

	// Anything else: ask the active section's renderer if it wants the key.
	cur, ok := m.currentSection()
	if !ok {
		return m, nil
	}
	r, ok := sections.Get(cur.Type)
	if !ok {
		return m, nil
	}
	if newSel, handled := r.HandleKey(cur, m.selection[cur.ID], msg); handled {
		m.selection[cur.ID] = newSel
	}
	return m, nil
}

func (m *Model) currentSection() (content.Section, bool) {
	if m.activeTab < 0 || m.activeTab >= len(m.tabs) {
		return content.Section{}, false
	}
	return m.tabs[m.activeTab], true
}
