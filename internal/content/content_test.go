package content

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeTOML(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "content.toml")
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write tmp toml: %v", err)
	}
	return path
}

func TestLoad_MinimalProfile_AppliesDefaults(t *testing.T) {
	path := writeTOML(t, `name = "Jane"`)
	p, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if p.Splash.Title != "Jane" {
		t.Errorf("splash.title default = %q, want %q", p.Splash.Title, "Jane")
	}
	if p.Splash.CTA != "Press Enter to start" {
		t.Errorf("splash.cta default = %q, want %q", p.Splash.CTA, "Press Enter to start")
	}
	if len(p.Tabs) != len(BuiltinTabs) {
		t.Fatalf("tabs default count = %d, want %d", len(p.Tabs), len(BuiltinTabs))
	}
	for i, want := range BuiltinTabs {
		if p.Tabs[i].ID != want {
			t.Errorf("tabs[%d].id = %q, want %q", i, p.Tabs[i].ID, want)
		}
		if p.Tabs[i].Label == "" {
			t.Errorf("tabs[%d].label is empty — should have a default", i)
		}
	}
}

func TestLoad_MissingName_Errors(t *testing.T) {
	path := writeTOML(t, `tagline = "no name"`)
	if _, err := Load(path); err == nil || !strings.Contains(err.Error(), "name") {
		t.Fatalf("expected name-required error, got %v", err)
	}
}

func TestLoad_UnknownTabID_Errors(t *testing.T) {
	path := writeTOML(t, `
name = "Jane"
[[tabs]]
id = "cvv"
`)
	_, err := Load(path)
	if err == nil || !strings.Contains(err.Error(), `"cvv"`) {
		t.Fatalf("expected unknown-tab error mentioning cvv, got %v", err)
	}
}

func TestLoad_DuplicateTab_Errors(t *testing.T) {
	path := writeTOML(t, `
name = "Jane"
[[tabs]]
id = "cv"
[[tabs]]
id = "cv"
`)
	_, err := Load(path)
	if err == nil || !strings.Contains(err.Error(), "more than once") {
		t.Fatalf("expected duplicate-tab error, got %v", err)
	}
}

func TestLoad_BadHexColor_Errors(t *testing.T) {
	path := writeTOML(t, `
name = "Jane"
[theme]
accent = "red"
`)
	_, err := Load(path)
	if err == nil || !strings.Contains(err.Error(), "theme.accent") {
		t.Fatalf("expected theme color error, got %v", err)
	}
}

func TestLoad_CustomSplashAndTheme(t *testing.T) {
	path := writeTOML(t, `
name = "Jane"
[splash]
title = "JD"
cta = "Hit Enter"
[theme]
accent = "#abcdef"
`)
	p, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if p.Splash.Title != "JD" {
		t.Errorf("splash.title = %q, want JD", p.Splash.Title)
	}
	if p.Splash.CTA != "Hit Enter" {
		t.Errorf("splash.cta = %q, want Hit Enter", p.Splash.CTA)
	}
	if p.Theme.Accent != "#abcdef" {
		t.Errorf("theme.accent = %q, want #abcdef", p.Theme.Accent)
	}
}

func TestVisibleTabs_HidesEmptySections(t *testing.T) {
	path := writeTOML(t, `
name = "Jane"
[[cv]]
role = "Dev"
`)
	p, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	visible := p.VisibleTabs()
	wantIDs := map[TabID]bool{TabStart: true, TabCV: true}
	if len(visible) != len(wantIDs) {
		t.Fatalf("visible tabs = %d (%+v), want %d", len(visible), visible, len(wantIDs))
	}
	for _, v := range visible {
		if !wantIDs[v.ID] {
			t.Errorf("unexpected visible tab %q", v.ID)
		}
	}
}

func TestLoad_ExampleTOML(t *testing.T) {
	// Sanity: the shipped example must always parse and validate.
	p, err := Load("../../content.example.toml")
	if err != nil {
		t.Fatalf("content.example.toml does not load: %v", err)
	}
	if p.Name == "" {
		t.Error("example profile has empty name")
	}
	if len(p.VisibleTabs()) == 0 {
		t.Error("example profile has no visible tabs")
	}
}
