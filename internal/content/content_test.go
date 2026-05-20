package content

import (
	"errors"
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

// fakeValidator stands in for the TUI's section registry so this package
// can be tested without importing it (which would cause a cycle).
type fakeValidator struct {
	known map[string]bool
}

func (f fakeValidator) Known(typ string) bool { return f.known[typ] }
func (f fakeValidator) Validate(s Section) error {
	if !f.known[s.Type] {
		return nil
	}
	return nil
}
func (f fakeValidator) IsEmpty(s Section) bool {
	// Consider a section empty when it has neither lines nor items.
	return len(s.Lines) == 0 && len(s.Items) == 0
}

func withFakeValidator(t *testing.T, types ...string) {
	t.Helper()
	prev := sectionValidator
	known := make(map[string]bool, len(types))
	for _, ty := range types {
		known[ty] = true
	}
	SetSectionValidator(fakeValidator{known: known})
	t.Cleanup(func() { sectionValidator = prev })
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
	if len(p.Sections) != 0 {
		t.Errorf("minimal profile should have no sections, got %d", len(p.Sections))
	}
}

func TestLoad_MissingFile_GivesMountHint(t *testing.T) {
	_, err := Load(filepath.Join(t.TempDir(), "does-not-exist.toml"))
	if err == nil {
		t.Fatal("expected error for missing content file")
	}
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("error should wrap os.ErrNotExist, got %v", err)
	}
	if !strings.Contains(err.Error(), "mount your content.toml") {
		t.Errorf("error should hint at mounting, got %v", err)
	}
}

func TestLoad_MissingName_Errors(t *testing.T) {
	path := writeTOML(t, `tagline = "no name"`)
	if _, err := Load(path); err == nil || !strings.Contains(err.Error(), "name") {
		t.Fatalf("expected name-required error, got %v", err)
	}
}

func TestLoad_UnknownSectionType_Errors(t *testing.T) {
	withFakeValidator(t, "text", "list", "links")
	path := writeTOML(t, `
name = "Jane"
[[sections]]
id = "x"
type = "table"
`)
	_, err := Load(path)
	if err == nil || !strings.Contains(err.Error(), `"table"`) {
		t.Fatalf("expected unknown-type error mentioning table, got %v", err)
	}
}

func TestLoad_DuplicateSectionID_Errors(t *testing.T) {
	withFakeValidator(t, "list")
	path := writeTOML(t, `
name = "Jane"
[[sections]]
id = "experience"
type = "list"
[[sections]]
id = "experience"
type = "list"
`)
	_, err := Load(path)
	if err == nil || !strings.Contains(err.Error(), "more than once") {
		t.Fatalf("expected duplicate-id error, got %v", err)
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

func TestLoad_DefaultLabelFromID(t *testing.T) {
	withFakeValidator(t, "text")
	path := writeTOML(t, `
name = "Jane"
[[sections]]
id   = "start"
type = "text"
lines = ["hi"]
`)
	p, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got := p.Sections[0].Label; got != "Start" {
		t.Errorf("default label = %q, want %q", got, "Start")
	}
}

func TestLoad_NoValidator_Errors(t *testing.T) {
	// Hard fail if a profile declares sections but no SectionValidator has
	// been wired in. Prevents silent acceptance of unknown section types
	// when callers forget the side-effect import of internal/tui/sections.
	prev := sectionValidator
	sectionValidator = nil
	t.Cleanup(func() { sectionValidator = prev })

	path := writeTOML(t, `
name = "Jane"
[[sections]]
id   = "x"
type = "text"
lines = ["hi"]
`)
	_, err := Load(path)
	if err == nil || !strings.Contains(err.Error(), "validator") {
		t.Fatalf("expected no-validator error, got %v", err)
	}
}

func TestLoad_DefaultLabel_UnicodeFirstRune(t *testing.T) {
	withFakeValidator(t, "text")
	path := writeTOML(t, `
name = "Jane"
[[sections]]
id   = "über"
type = "text"
lines = ["hi"]
`)
	p, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got := p.Sections[0].Label; got != "Über" {
		t.Errorf("default label = %q, want %q", got, "Über")
	}
}

func TestVisibleSections_HidesEmpty(t *testing.T) {
	withFakeValidator(t, "text", "list")
	path := writeTOML(t, `
name = "Jane"

[[sections]]
id   = "start"
type = "text"
lines = ["hi"]

[[sections]]
id   = "experience"
type = "list"
`)
	p, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	vis := p.VisibleSections()
	if len(vis) != 1 || vis[0].ID != "start" {
		t.Fatalf("visible sections = %+v, want only start", vis)
	}
}
