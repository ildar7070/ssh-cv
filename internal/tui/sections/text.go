package sections

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/ildar7070/ssh-cv/internal/content"
)

// Gap between the text column and the ASCII-art column when they sit
// side by side, and the minimum width the text column needs before we
// give up on the side-by-side layout and stack instead.
const (
	asciiGap       = 3
	minTextColumnW = 24
)

// textRenderer is the "type = text" section. It shows a heading
// (Profile.Name + Tagline) followed by free-form paragraph lines, with an
// optional ASCII-art block beside or below the text.
type textRenderer struct{}

func (textRenderer) Type() string { return "text" }

func (textRenderer) Validate(s content.Section) error { return nil }

func (textRenderer) IsEmpty(s content.Section) bool {
	if strings.TrimSpace(s.ASCII) != "" {
		return false
	}
	for _, l := range s.Lines {
		if strings.TrimSpace(l) != "" {
			return false
		}
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
	text := lipgloss.JoinVertical(lipgloss.Left, blocks...)

	content := layoutWithASCII(st, text, s.ASCII, ctx.Width)
	return place(ctx.Width, ctx.Height, content)
}

// layoutWithASCII places the ASCII-art block to the right of the text when
// the terminal is wide enough for both columns, and below it otherwise.
// An empty ascii returns the text unchanged.
func layoutWithASCII(st Styles, text, ascii string, width int) string {
	ascii = strings.TrimRight(ascii, "\n")
	if strings.TrimSpace(ascii) == "" {
		return text
	}
	art := st.Ascii.Render(ascii)
	artW := lipgloss.Width(art)

	// Side by side only if the text column keeps a usable width.
	if width-artW-asciiGap >= minTextColumnW {
		textCol := lipgloss.NewStyle().Width(width - artW - asciiGap).Render(text)
		return lipgloss.JoinHorizontal(lipgloss.Top,
			textCol, strings.Repeat(" ", asciiGap), art)
	}
	return lipgloss.JoinVertical(lipgloss.Left, text, "", art)
}

func (textRenderer) FooterHint() string { return "" }

func (textRenderer) HandleKey(_ content.Section, sel int, _ tea.KeyPressMsg) (int, bool) {
	return sel, false
}

func init() { Register(textRenderer{}) }
