package sections

import (
	"fmt"
	"math"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/ildar7070/ssh-cv/internal/content"
)

// 30/70 split for list-based sections.
const (
	listPaneRatioNumerator   = 30
	listPaneRatioDenominator = 100
	minListWidth             = 18
	minDetailWidth           = 10
	listDetailGap            = 2
)

// listRenderer is the "type = list" section. Left pane: titles. Right
// pane: title, subtitle, meta, summary, bullets of the selected item.
type listRenderer struct{}

func (listRenderer) Type() string { return "list" }

func (listRenderer) Validate(s content.Section) error {
	for i, it := range s.Items {
		if strings.TrimSpace(it.Title) == "" {
			return fmt.Errorf("items[%d]: title is required for list sections", i)
		}
	}
	return nil
}

func (listRenderer) IsEmpty(s content.Section) bool {
	return len(s.Items) == 0
}

func (listRenderer) Render(s content.Section, ctx RenderContext) string {
	width, height := ctx.Width, ctx.Height
	listW := width * listPaneRatioNumerator / listPaneRatioDenominator
	if listW < minListWidth {
		listW = minListWidth
	}
	detailW := width - listW - listDetailGap
	if detailW < minDetailWidth {
		detailW = minDetailWidth
	}

	list := renderList(ctx.Styles, s.Items, ctx.Selected, listW, height)
	det := lipgloss.NewStyle().
		Width(detailW).
		Height(height).
		Render(renderListDetail(ctx.Styles, s.Items, ctx.Selected))

	return lipgloss.JoinHorizontal(lipgloss.Top, list, strings.Repeat(" ", listDetailGap), det)
}

func (listRenderer) FooterHint() string {
	return "↑/↓ select · ←/→ tab · t ◑ · q quit"
}

func (listRenderer) HandleKey(s content.Section, sel int, msg tea.KeyPressMsg) (int, bool) {
	n := len(s.Items)
	if n == 0 {
		return sel, false
	}
	switch msg.String() {
	case "up", "k":
		return clamp(sel-1, 0, n-1), true
	case "down", "j":
		return clamp(sel+1, 0, n-1), true
	case "home", "g":
		return 0, true
	case "end", "G":
		return clamp(math.MaxInt, 0, n-1), true
	}
	return sel, false
}

func renderList(st Styles, items []content.Item, selected, width, height int) string {
	rows := make([]string, 0, len(items))
	for i, it := range items {
		marker := "  "
		style := st.ListItem
		if i == selected {
			marker = "▸ "
			style = st.ListItemSelected
		}
		line := ansi.Truncate(marker+it.Title, width-2, "…")
		rows = append(rows, style.Width(width).Render(line))
	}
	out := lipgloss.JoinVertical(lipgloss.Left, rows...)
	return lipgloss.NewStyle().Width(width).Height(height).Render(out)
}

func renderListDetail(st Styles, items []content.Item, idx int) string {
	if idx < 0 || idx >= len(items) {
		return st.DetailMuted.Render("(no entry)")
	}
	it := items[idx]
	bullets := make([]string, 0, len(it.Bullets))
	for _, bul := range it.Bullets {
		bullets = append(bullets, st.DetailMuted.Render("• ")+st.DetailBody.Render(bul))
	}
	return renderDetailBlock(st, detailBlock{
		heading:  it.Title,
		subtitle: it.Subtitle,
		meta:     it.Meta,
		body:     it.Summary,
		bullets:  bullets,
	})
}

func init() { Register(listRenderer{}) }
