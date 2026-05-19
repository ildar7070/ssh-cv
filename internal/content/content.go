// Package content loads the personal profile shown to SSH visitors.
//
// The profile is parsed from a TOML file. Everything except `name` is
// optional — leaving a section out hides the corresponding tab in the
// TUI. See content.example.toml for a fully documented template.
package content

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	toml "github.com/pelletier/go-toml/v2"
)

// TabID is a stable identifier for a built-in tab renderer.
type TabID string

const (
	TabStart    TabID = "start"
	TabCV       TabID = "cv"
	TabProjects TabID = "projects"
	TabContact  TabID = "contact"
)

// BuiltinTabs lists the tab IDs the TUI knows how to render. Used both for
// validation and as the implicit default when no [[tabs]] block is given.
var BuiltinTabs = []TabID{TabStart, TabCV, TabProjects, TabContact}

// Profile is the parsed content.toml file.
type Profile struct {
	Name    string `toml:"name"`
	Tagline string `toml:"tagline"`

	Splash  Splash    `toml:"splash"`
	About   About     `toml:"about"`
	Theme   Theme     `toml:"theme"`
	Tabs    []TabSpec `toml:"tabs"`
	CV      []CVEntry `toml:"cv"`
	Projects []Project `toml:"projects"`
	Contact []ContactLink `toml:"contact"`
}

// Splash configures the entry screen shown before a visitor presses Enter.
// Empty fields fall back to sensible defaults (title=Name, cta="Press Enter to start").
type Splash struct {
	Title string `toml:"title"`
	CTA   string `toml:"cta"`
}

type About struct {
	Lines []string `toml:"lines"`
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

// TabSpec controls which built-in tabs appear in the header and in what order.
// Label is optional — when empty, the TUI uses a sensible default per ID.
type TabSpec struct {
	ID    TabID  `toml:"id"`
	Label string `toml:"label"`
}

type CVEntry struct {
	Role    string   `toml:"role"`
	Company string   `toml:"company"`
	Period  string   `toml:"period"`
	Summary string   `toml:"summary"`
	Bullets []string `toml:"bullets"`
}

type Project struct {
	Name    string   `toml:"name"`
	Tagline string   `toml:"tagline"`
	URL     string   `toml:"url"`
	Details []string `toml:"details"`
}

// ContactLink is one row on the contact page. Value is the URL or address;
// Label is the short tag shown to the left ("email", "github", ...).
type ContactLink struct {
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
	if len(p.Tabs) == 0 {
		p.Tabs = make([]TabSpec, 0, len(BuiltinTabs))
		for _, id := range BuiltinTabs {
			p.Tabs = append(p.Tabs, TabSpec{ID: id})
		}
	}
	for i, t := range p.Tabs {
		if t.Label == "" {
			p.Tabs[i].Label = defaultTabLabel(t.ID)
		}
	}
}

func defaultTabLabel(id TabID) string {
	switch id {
	case TabStart:
		return "Start"
	case TabCV:
		return "CV"
	case TabProjects:
		return "Projects"
	case TabContact:
		return "Contact"
	}
	return string(id)
}

// Validate returns a descriptive error if the profile is missing required
// data or references unknown tab IDs / invalid colors.
func (p *Profile) Validate() error {
	if strings.TrimSpace(p.Name) == "" {
		return fmt.Errorf("name is required")
	}
	seen := make(map[TabID]bool, len(p.Tabs))
	for _, t := range p.Tabs {
		if !isBuiltinTab(t.ID) {
			return fmt.Errorf("tab %q is not a built-in tab (allowed: %v)", t.ID, BuiltinTabs)
		}
		if seen[t.ID] {
			return fmt.Errorf("tab %q listed more than once", t.ID)
		}
		seen[t.ID] = true
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

func isBuiltinTab(id TabID) bool {
	for _, b := range BuiltinTabs {
		if id == b {
			return true
		}
	}
	return false
}

var hexColor = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

// HasContent reports whether the given tab has anything meaningful to render.
// Used by the TUI to hide empty tabs from the header.
func (p *Profile) HasContent(id TabID) bool {
	switch id {
	case TabStart:
		return strings.TrimSpace(p.Name) != "" || strings.TrimSpace(p.Tagline) != "" || len(p.About.Lines) > 0
	case TabCV:
		return len(p.CV) > 0
	case TabProjects:
		return len(p.Projects) > 0
	case TabContact:
		return len(p.Contact) > 0
	}
	return false
}

// VisibleTabs returns the tabs the TUI should render — i.e. each declared
// tab whose underlying section has content.
func (p *Profile) VisibleTabs() []TabSpec {
	out := make([]TabSpec, 0, len(p.Tabs))
	for _, t := range p.Tabs {
		if p.HasContent(t.ID) {
			out = append(out, t)
		}
	}
	return out
}
