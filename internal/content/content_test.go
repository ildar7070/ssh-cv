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

func TestLoad_ExampleTOML(t *testing.T) {
	// Sanity: the shipped example must always parse and validate against
	// the real renderer registry. Importing sections here would be a
	// cycle, so this test runs from the tui/sections side instead — see
	// internal/tui/sections/example_test.go.
	t.Skip("moved to internal/tui/sections to avoid an import cycle")
}
