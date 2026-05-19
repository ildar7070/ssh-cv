package sections

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ildar7070/ssh-cv/internal/content"
)

// textRenderer is the "type = text" section. It shows a heading
// (Profile.Name + Tagline) followed by free-form paragraph lines.
type textRenderer struct{}

func (textRenderer) Type() string { return "text" }

func (textRenderer) Validate(s content.Section) error { return nil }

func (textRenderer) IsEmpty(s content.Section) bool {
	if len(s.Lines) > 0 {
		return false
	}
	return true
}

func (textRenderer) Render(s content.Section, ctx RenderContext) string {
	st := ctx.Styles
	p := ctx.Profile

	var blocks []string
	if p != nil && strings.TrimSpace(p.Name) != "" {
		blocks = append(blocks, st.DetailHeading.Render(p.Name))
		if strings.TrimSpace(p.Tagline) != "" {
			blocks = append(blocks, st.DetailSub.Render(p.Tagline))
		}
		blocks = append(blocks, "")
	}
	var body strings.Builder
	for i, l := range s.Lines {
		if i > 0 {
			body.WriteByte('\n')
		}
		body.WriteString(st.DetailBody.Render(l))
	}
	blocks = append(blocks, body.String())

	return place(ctx.Width, ctx.Height,
		lipgloss.JoinVertical(lipgloss.Left, blocks...))
}

func (textRenderer) FooterHint() string { return "" }

func (textRenderer) HandleKey(_ content.Section, sel int, _ tea.KeyMsg) (int, bool) {
	return sel, false
}

func init() { Register(textRenderer{}) }
