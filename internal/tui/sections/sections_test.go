package sections

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ildar7070/ssh-cv/internal/content"
)

func mustRenderer(t *testing.T, typ string) Renderer {
	t.Helper()
	r, ok := Get(typ)
	if !ok {
		t.Fatalf("renderer %q not registered", typ)
	}
	return r
}

func TestRegistry_BuiltinTypes(t *testing.T) {
	for _, typ := range []string{"text", "list", "links"} {
		if _, ok := Get(typ); !ok {
			t.Errorf("renderer %q not registered", typ)
		}
	}
}

func TestList_ValidateRequiresTitle(t *testing.T) {
	r := mustRenderer(t, "list")
	err := r.Validate(content.Section{
		Type:  "list",
		Items: []content.Item{{Title: ""}},
	})
	if err == nil {
		t.Fatal("expected error when list item has no title")
	}
}

func TestLinks_ValidateRequiresLabelAndValue(t *testing.T) {
	r := mustRenderer(t, "links")
	if err := r.Validate(content.Section{Type: "links", Items: []content.Item{{Label: ""}}}); err == nil {
		t.Error("expected error when links item has no label")
	}
	if err := r.Validate(content.Section{Type: "links", Items: []content.Item{{Label: "x", Value: ""}}}); err == nil {
		t.Error("expected error when links item has no value")
	}
}

func TestList_HandleKey(t *testing.T) {
	r := mustRenderer(t, "list")
	s := content.Section{Type: "list", Items: []content.Item{{Title: "a"}, {Title: "b"}, {Title: "c"}}}

	if got, handled := r.HandleKey(s, 0, tea.KeyMsg{Type: tea.KeyDown}); !handled || got != 1 {
		t.Errorf("down from 0: got %d handled=%v, want 1 true", got, handled)
	}
	if got, handled := r.HandleKey(s, 0, tea.KeyMsg{Type: tea.KeyUp}); !handled || got != 0 {
		t.Errorf("up at 0: got %d handled=%v, want 0 true", got, handled)
	}
	if got, handled := r.HandleKey(s, 0, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")}); !handled || got != 2 {
		t.Errorf("end: got %d handled=%v, want 2 true", got, handled)
	}
	if _, handled := r.HandleKey(s, 1, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")}); handled {
		t.Error("unrelated key should not be handled")
	}
}

func TestIsEmpty(t *testing.T) {
	cases := []struct {
		typ  string
		sec  content.Section
		want bool
	}{
		{"text", content.Section{Type: "text"}, true},
		{"text", content.Section{Type: "text", Lines: []string{"hi"}}, false},
		{"list", content.Section{Type: "list"}, true},
		{"list", content.Section{Type: "list", Items: []content.Item{{Title: "x"}}}, false},
		{"links", content.Section{Type: "links"}, true},
		{"links", content.Section{Type: "links", Items: []content.Item{{Label: "e", Value: "v"}}}, false},
	}
	for _, c := range cases {
		r := mustRenderer(t, c.typ)
		if got := r.IsEmpty(c.sec); got != c.want {
			t.Errorf("%s IsEmpty(%+v) = %v, want %v", c.typ, c.sec, got, c.want)
		}
	}
}

func TestExampleTOML_ParsesAndValidates(t *testing.T) {
	// With this package imported, content.SetSectionValidator has been
	// wired, so Load enforces real type/validation rules.
	p, err := content.Load("../../../content.example.toml")
	if err != nil {
		t.Fatalf("content.example.toml does not load: %v", err)
	}
	if p.Name == "" {
		t.Error("example profile has empty name")
	}
	if len(p.VisibleSections()) == 0 {
		t.Error("example profile has no visible sections")
	}
}

func TestRender_DoesNotPanic(t *testing.T) {
	// Smoke test: each renderer should produce something for a populated section.
	styles := NewStyles(lipgloss.DefaultRenderer(), content.Theme{})
	profile := &content.Profile{Name: "Test", Tagline: "tag"}
	ctx := RenderContext{Profile: profile, Styles: styles, Width: 80, Height: 20}

	sections := []content.Section{
		{ID: "s", Type: "text", Label: "S", Lines: []string{"hi"}},
		{ID: "l", Type: "list", Label: "L", Items: []content.Item{{Title: "a", Bullets: []string{"b"}}}},
		{ID: "c", Type: "links", Label: "C", Items: []content.Item{{Label: "e", Value: "v"}}},
	}
	for _, s := range sections {
		r := mustRenderer(t, s.Type)
		out := r.Render(s, ctx)
		if strings.TrimSpace(out) == "" {
			t.Errorf("renderer %q produced empty output", s.Type)
		}
	}
}
