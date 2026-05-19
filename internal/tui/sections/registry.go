package sections

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/ildar7070/ssh-cv/internal/content"
)

// RenderContext carries the per-frame inputs a renderer needs. Profile
// is included so renderers that compose other top-level data (the start
// page mixes name/tagline with its own lines) can read it without
// stuffing it into Section.
type RenderContext struct {
	Profile  *content.Profile
	Styles   Styles
	Width    int
	Height   int
	Selected int
}

// Renderer is the contract every section type implements. One instance
// per type; renderers must be stateless and safe for concurrent use —
// the TUI calls them from per-session goroutines.
type Renderer interface {
	// Type is the value matched against Section.Type in content.toml.
	Type() string

	// Validate is called at load time. Returns a descriptive error if
	// the section's payload is missing required fields for this type.
	Validate(s content.Section) error

	// IsEmpty reports whether the section has nothing to show; empty
	// sections are hidden from the tab header.
	IsEmpty(s content.Section) bool

	// Render returns the body for this section.
	Render(s content.Section, ctx RenderContext) string

	// FooterHint is the help text shown in the footer while this section
	// is active. Return "" to fall back to the default hint.
	FooterHint() string

	// HandleKey lets the renderer consume list-style key events. Return
	// (newSelected, true) to indicate the key was handled. Renderers
	// that have no selection (text, links) should return (selected, false).
	HandleKey(s content.Section, selected int, msg tea.KeyMsg) (int, bool)
}

var registry = map[string]Renderer{}

// Register adds a renderer to the global registry. Call from init().
// Panics on duplicate registration so collisions surface at startup
// rather than as silent overrides.
func Register(r Renderer) {
	if _, exists := registry[r.Type()]; exists {
		panic(fmt.Sprintf("sections: type %q registered twice", r.Type()))
	}
	registry[r.Type()] = r
}

// Get returns the renderer for typ, if any.
func Get(typ string) (Renderer, bool) {
	r, ok := registry[typ]
	return r, ok
}

// validator is the content.SectionValidator wired into the content
// package on init. Centralises lookups so content stays decoupled from
// individual renderers.
type validator struct{}

func (validator) Known(typ string) bool {
	_, ok := registry[typ]
	return ok
}

func (validator) Validate(s content.Section) error {
	r, ok := registry[s.Type]
	if !ok {
		return fmt.Errorf("unknown type %q", s.Type)
	}
	return r.Validate(s)
}

func (validator) IsEmpty(s content.Section) bool {
	r, ok := registry[s.Type]
	if !ok {
		return true
	}
	return r.IsEmpty(s)
}

func init() {
	content.SetSectionValidator(validator{})
}
