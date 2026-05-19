// Package content loads the personal profile shown to SSH visitors.
//
// The profile is parsed from a TOML file. Top-level keys describe the
// person (name, tagline, splash, theme); the body is an ordered list of
// [[sections]] — one section per visible tab. Each section names a
// renderer type ("text", "list", "links", …) registered in
// internal/tui/sections.
//
// See content.example.toml for a fully documented template.
package content

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	toml "github.com/pelletier/go-toml/v2"
)

// SectionValidator is implemented by renderer registries that want to
// validate section payloads at load time. The TUI's section registry
// registers itself via SetSectionValidator; content stays decoupled from
// the renderer package to avoid an import cycle.
type SectionValidator interface {
	Validate(s Section) error
	Known(typ string) bool
	IsEmpty(s Section) bool
}

var sectionValidator SectionValidator

// SetSectionValidator wires in a renderer registry. Called from the TUI
// package's init().
func SetSectionValidator(v SectionValidator) { sectionValidator = v }

// Profile is the parsed content.toml file.
type Profile struct {
	Name    string `toml:"name"`
	Tagline string `toml:"tagline"`

	Splash   Splash    `toml:"splash"`
	Theme    Theme     `toml:"theme"`
	Sections []Section `toml:"sections"`
}

// Splash configures the entry screen shown before a visitor presses Enter.
// Empty fields fall back to sensible defaults (title=Name, cta="Press Enter to start").
type Splash struct {
	Title string `toml:"title"`
	CTA   string `toml:"cta"`
}

// Theme overrides the default Lipgloss palette. All fields are optional;
// empty values keep the built-in defaults.
type Theme struct {
	Accent     string `toml:"accent"`
	Accent2    string `toml:"accent2"`
	Foreground string `toml:"foreground"`
	Muted      string `toml:"muted"`
	Background string `toml:"background"`
	Selection  string `toml:"selection"`
}

// Section is one tab in the TUI. The renderer addressed by Type decides
// which of the optional fields it consumes; unused fields are simply
// ignored. New section types are added by registering a renderer, not by
// extending this struct unless the new shape genuinely doesn't fit.
type Section struct {
	ID    string `toml:"id"`
	Type  string `toml:"type"`
	Label string `toml:"label"`

	// Text-style sections: free-form paragraph lines.
	Lines []string `toml:"lines"`

	// List-style and links-style sections: ordered rows.
	Items []Item `toml:"items"`
}

// Item is the row shape used by list-style and links-style sections.
// Renderers pick the fields that apply to their layout.
type Item struct {
	// list-style fields
	Title    string   `toml:"title"`
	Subtitle string   `toml:"subtitle"`
	Meta     string   `toml:"meta"`
	Summary  string   `toml:"summary"`
	Bullets  []string `toml:"bullets"`

	// links-style fields
	Label string `toml:"label"`
	Value string `toml:"value"`
}

// Load reads, parses, and validates the profile at path.
func Load(path string) (*Profile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var p Profile
	if err := toml.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	p.applyDefaults()
	if err := p.Validate(); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return &p, nil
}

func (p *Profile) applyDefaults() {
	if p.Splash.Title == "" {
		p.Splash.Title = p.Name
	}
	if p.Splash.CTA == "" {
		p.Splash.CTA = "Press Enter to start"
	}
	for i := range p.Sections {
		if p.Sections[i].Label == "" {
			p.Sections[i].Label = defaultLabel(p.Sections[i].ID)
		}
	}
}

func defaultLabel(id string) string {
	if id == "" {
		return ""
	}
	r, size := utf8.DecodeRuneInString(id)
	if r == utf8.RuneError {
		return id
	}
	return string(unicode.ToUpper(r)) + id[size:]
}

// Validate returns a descriptive error if the profile is missing required
// data or references unknown section types / invalid colors.
func (p *Profile) Validate() error {
	if strings.TrimSpace(p.Name) == "" {
		return fmt.Errorf("name is required")
	}
	if len(p.Sections) > 0 && sectionValidator == nil {
		return fmt.Errorf("no section validator registered — import internal/tui/sections for its init() to wire one in")
	}
	seen := make(map[string]bool, len(p.Sections))
	for i, s := range p.Sections {
		if strings.TrimSpace(s.ID) == "" {
			return fmt.Errorf("sections[%d]: id is required", i)
		}
		if strings.TrimSpace(s.Type) == "" {
			return fmt.Errorf("sections[%d] (%q): type is required", i, s.ID)
		}
		if seen[s.ID] {
			return fmt.Errorf("section id %q listed more than once", s.ID)
		}
		seen[s.ID] = true
		if !sectionValidator.Known(s.Type) {
			return fmt.Errorf("section %q: unknown type %q", s.ID, s.Type)
		}
		if err := sectionValidator.Validate(s); err != nil {
			return fmt.Errorf("section %q: %w", s.ID, err)
		}
	}
	for field, color := range map[string]string{
		"accent":     p.Theme.Accent,
		"accent2":    p.Theme.Accent2,
		"foreground": p.Theme.Foreground,
		"muted":      p.Theme.Muted,
		"background": p.Theme.Background,
		"selection":  p.Theme.Selection,
	} {
		if color != "" && !hexColor.MatchString(color) {
			return fmt.Errorf("theme.%s = %q: expected hex color like #7ee787", field, color)
		}
	}
	return nil
}

var hexColor = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

// VisibleSections returns sections whose backing data is non-empty, as
// judged by the registered SectionValidator. If no validator is wired in
// (tests, library use), every section is considered visible.
func (p *Profile) VisibleSections() []Section {
	out := make([]Section, 0, len(p.Sections))
	for _, s := range p.Sections {
		if sectionValidator != nil && sectionValidator.IsEmpty(s) {
			continue
		}
		out = append(out, s)
	}
	return out
}
