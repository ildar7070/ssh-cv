package tui

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/ildar7070/ssh-cv/internal/content"
	_ "github.com/ildar7070/ssh-cv/internal/tui/sections"
)

func newTestProfile() *content.Profile {
	return &content.Profile{
		Name:    "Test",
		Tagline: "test tagline",
		Splash:  content.Splash{Title: "Test", CTA: "Press Enter to start"},
		Sections: []content.Section{
			{ID: "start", Type: "text", Label: "Start", Lines: []string{"hello"}},
			{ID: "experience", Type: "list", Label: "Experience", Items: []content.Item{
				{Title: "A"}, {Title: "B"}, {Title: "C"},
			}},
			{ID: "projects", Type: "list", Label: "Projects", Items: []content.Item{
				{Title: "P1"}, {Title: "P2"},
			}},
			{ID: "contact", Type: "links", Label: "Contact", Items: []content.Item{
				{Label: "email", Value: "x@y"},
			}},
		},
	}
}

func keyMsg(k string) tea.KeyPressMsg {
	switch k {
	case "tab":
		return tea.KeyPressMsg{Code: tea.KeyTab}
	case "shift+tab":
		return tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift}
	case "enter":
		return tea.KeyPressMsg{Code: tea.KeyEnter}
	case "up":
		return tea.KeyPressMsg{Code: tea.KeyUp}
	case "down":
		return tea.KeyPressMsg{Code: tea.KeyDown}
	case "esc":
		return tea.KeyPressMsg{Code: tea.KeyEsc}
	}
	r := []rune(k)[0]
	return tea.KeyPressMsg{Code: r, Text: k}
}

func step(m Model, k string) Model {
	out, _ := m.Update(keyMsg(k))
	return out.(Model)
}

func TestSplashToApp_Enter(t *testing.T) {
	m := New(newTestProfile())
	if m.mode != modeSplash {
		t.Fatalf("initial mode = %v, want splash", m.mode)
	}
	m = step(m, "enter")
	if m.mode != modeApp {
		t.Errorf("after enter, mode = %v, want app", m.mode)
	}
}

func TestTabSwitching_TabKeyAndDigits(t *testing.T) {
	m := New(newTestProfile())
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
	m := New(newTestProfile())
	m = step(m, "enter")

	m = step(m, "shift+tab")
	if m.activeTab != 0 {
		t.Errorf("shift+tab at tab 0: activeTab = %d, want 0 (no wrap)", m.activeTab)
	}

	for i := 0; i < 10; i++ {
		m = step(m, "tab")
	}
	if m.activeTab != 3 {
		t.Errorf("after many tabs, activeTab = %d, want 3 (no wrap)", m.activeTab)
	}
}

func TestSelectionBounds(t *testing.T) {
	m := New(newTestProfile())
	m = step(m, "enter")
	m = step(m, "2") // experience tab (3 items)

	m = step(m, "up")
	if got := m.selection["experience"]; got != 0 {
		t.Errorf("up at top: selection = %d, want 0", got)
	}
	for i := 0; i < 5; i++ {
		m = step(m, "down")
	}
	if got := m.selection["experience"]; got != 2 {
		t.Errorf("after 5 downs in 3-item list: selection = %d, want 2", got)
	}
	m = step(m, "G")
	if got := m.selection["experience"]; got != 2 {
		t.Errorf("after G: selection = %d, want 2", got)
	}
	m = step(m, "g")
	if got := m.selection["experience"]; got != 0 {
		t.Errorf("after g: selection = %d, want 0", got)
	}
}

func TestEmptyTabsHidden(t *testing.T) {
	p := &content.Profile{
		Name: "Test",
		Sections: []content.Section{
			{ID: "start", Type: "text", Label: "Start", Lines: []string{"hi"}},
			{ID: "experience", Type: "list", Label: "Experience"},
			{ID: "projects", Type: "list", Label: "Projects"},
			{ID: "contact", Type: "links", Label: "Contact"},
		},
	}

	m := New(p)
	if len(m.tabs) != 1 {
		t.Fatalf("visible tabs = %d (%+v), want 1", len(m.tabs), m.tabs)
	}
	if m.tabs[0].ID != "start" {
		t.Errorf("only visible tab = %q, want start", m.tabs[0].ID)
	}
}

func TestQuit_Q(t *testing.T) {
	m := New(newTestProfile())
	_, cmd := m.Update(keyMsg("q"))
	if cmd == nil {
		t.Fatalf("expected tea.Quit command, got nil")
	}
	if msg := cmd(); msg == nil {
		t.Errorf("quit cmd returned nil msg")
	}
}
