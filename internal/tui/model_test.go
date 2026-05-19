package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/ildar7070/ssh-cv/internal/content"
)

func newTestProfile() *content.Profile {
	return &content.Profile{
		Name:    "Test",
		Tagline: "test tagline",
		Splash:  content.Splash{Title: "Test", CTA: "Press Enter to start"},
		About:   content.About{Lines: []string{"hello"}},
		Tabs: []content.TabSpec{
			{ID: content.TabStart, Label: "Start"},
			{ID: content.TabCV, Label: "CV"},
			{ID: content.TabProjects, Label: "Projects"},
			{ID: content.TabContact, Label: "Contact"},
		},
		CV: []content.CVEntry{
			{Role: "A"},
			{Role: "B"},
			{Role: "C"},
		},
		Projects: []content.Project{
			{Name: "P1"},
			{Name: "P2"},
		},
		Contact: []content.ContactLink{{Label: "email", Value: "x@y"}},
	}
}

func keyMsg(k string) tea.KeyMsg {
	switch k {
	case "tab":
		return tea.KeyMsg{Type: tea.KeyTab}
	case "shift+tab":
		return tea.KeyMsg{Type: tea.KeyShiftTab}
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "up":
		return tea.KeyMsg{Type: tea.KeyUp}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEscape}
	}
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)}
}

func step(m Model, k string) Model {
	out, _ := m.Update(keyMsg(k))
	return out.(Model)
}

func TestSplashToApp_Enter(t *testing.T) {
	m := New(newTestProfile(), nil)
	if m.mode != modeSplash {
		t.Fatalf("initial mode = %v, want splash", m.mode)
	}
	m = step(m, "enter")
	if m.mode != modeApp {
		t.Errorf("after enter, mode = %v, want app", m.mode)
	}
}

func TestTabSwitching_TabKeyAndDigits(t *testing.T) {
	m := New(newTestProfile(), nil)
	m = step(m, "enter")

	m = step(m, "tab")
	if m.activeTab != 1 {
		t.Errorf("after tab, activeTab = %d, want 1", m.activeTab)
	}
	m = step(m, "3")
	if m.activeTab != 2 {
		t.Errorf("after '3', activeTab = %d, want 2 (zero-indexed)", m.activeTab)
	}
	m = step(m, "shift+tab")
	if m.activeTab != 1 {
		t.Errorf("after shift+tab, activeTab = %d, want 1", m.activeTab)
	}
}

func TestTabSwitching_NoWrap(t *testing.T) {
	m := New(newTestProfile(), nil)
	m = step(m, "enter")

	// Step left at first tab.
	m = step(m, "shift+tab")
	if m.activeTab != 0 {
		t.Errorf("shift+tab at tab 0: activeTab = %d, want 0 (no wrap)", m.activeTab)
	}

	// Step right past last tab (4 tabs → max index 3).
	for i := 0; i < 10; i++ {
		m = step(m, "tab")
	}
	if m.activeTab != 3 {
		t.Errorf("after many tabs, activeTab = %d, want 3 (no wrap)", m.activeTab)
	}
}

func TestSelectionBounds(t *testing.T) {
	m := New(newTestProfile(), nil)
	m = step(m, "enter")
	m = step(m, "2") // CV tab (3 items)

	// Up at top stays at 0.
	m = step(m, "up")
	if got := m.selection[content.TabCV]; got != 0 {
		t.Errorf("up at top: selection = %d, want 0", got)
	}
	// Down moves; bound at 2.
	for i := 0; i < 5; i++ {
		m = step(m, "down")
	}
	if got := m.selection[content.TabCV]; got != 2 {
		t.Errorf("after 5 downs in 3-item list: selection = %d, want 2", got)
	}
	// End jumps to last (already there).
	m = step(m, "G")
	if got := m.selection[content.TabCV]; got != 2 {
		t.Errorf("after G: selection = %d, want 2", got)
	}
	// Home jumps to first.
	m = step(m, "g")
	if got := m.selection[content.TabCV]; got != 0 {
		t.Errorf("after g: selection = %d, want 0", got)
	}
}

func TestEmptyTabsHidden(t *testing.T) {
	p := &content.Profile{Name: "Test"}
	// Run through Load-style defaults.
	p.Tabs = []content.TabSpec{
		{ID: content.TabStart, Label: "Start"},
		{ID: content.TabCV, Label: "CV"},
		{ID: content.TabProjects, Label: "Projects"},
		{ID: content.TabContact, Label: "Contact"},
	}

	m := New(p, nil)
	// Only Start is "non-empty" because we have a Name; no CV/Projects/Contact.
	if len(m.tabs) != 1 {
		t.Fatalf("visible tabs = %d (%+v), want 1", len(m.tabs), m.tabs)
	}
	if m.tabs[0].ID != content.TabStart {
		t.Errorf("only visible tab = %q, want start", m.tabs[0].ID)
	}
}

func TestQuit_Q(t *testing.T) {
	m := New(newTestProfile(), nil)
	_, cmd := m.Update(keyMsg("q"))
	if cmd == nil {
		t.Fatalf("expected tea.Quit command, got nil")
	}
	// Smoke: cmd should return a tea.QuitMsg.
	if msg := cmd(); msg == nil {
		t.Errorf("quit cmd returned nil msg")
	}
}
