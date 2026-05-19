package sections

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ildar7070/ssh-cv/internal/content"
)

// Width of the "label" column on the links page.
const linkLabelWidth = 10

// linksRenderer is the "type = links" section. Label/value rows where
// the value is rendered underlined so terminals' URL detectors pick it up.
type linksRenderer struct{}

func (linksRenderer) Type() string { return "links" }

func (linksRenderer) Validate(s content.Section) error {
	for i, it := range s.Items {
		if strings.TrimSpace(it.Label) == "" {
			return fmt.Errorf("items[%d]: label is required for links sections", i)
		}
		if strings.TrimSpace(it.Value) == "" {
			return fmt.Errorf("items[%d] (%q): value is required for links sections", i, it.Label)
		}
	}
	return nil
}

func (linksRenderer) IsEmpty(s content.Section) bool {
	return len(s.Items) == 0
}

func (linksRenderer) Render(s content.Section, ctx RenderContext) string {
	st := ctx.Styles
	rows := []string{
		st.DetailHeading.Render(s.Label),
		"",
	}
	for _, it := range s.Items {
		if it.Value == "" {
			continue
		}
		rows = append(rows,
			st.DetailSub.Render(fmt.Sprintf("%-*s", linkLabelWidth, it.Label))+
				st.DetailLink.Render(it.Value),
		)
	}
	return place(ctx.Width, ctx.Height, lipgloss.JoinVertical(lipgloss.Left, rows...))
}

func (linksRenderer) FooterHint() string { return "" }

func (linksRenderer) HandleKey(_ content.Section, sel int, _ tea.KeyMsg) (int, bool) {
	return sel, false
}

func init() { Register(linksRenderer{}) }
