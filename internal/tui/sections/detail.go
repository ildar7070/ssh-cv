package sections

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// detailBlock is the shared layout used by list-style detail panes.
type detailBlock struct {
	heading  string
	subtitle string
	meta     string
	body     string
	bullets  []string
}

func renderDetailBlock(s Styles, d detailBlock) string {
	var b strings.Builder
	b.WriteString(s.DetailHeading.Render(d.heading))
	b.WriteByte('\n')
	if d.subtitle != "" {
		b.WriteString(s.DetailSub.Render(d.subtitle))
		if d.meta != "" {
			b.WriteString("   ")
			b.WriteString(s.DetailMuted.Render(d.meta))
		}
		b.WriteByte('\n')
	} else if d.meta != "" {
		b.WriteString(s.DetailMuted.Render(d.meta))
		b.WriteByte('\n')
	}
	b.WriteByte('\n')
	if d.body != "" {
		b.WriteString(s.DetailBody.Render(d.body))
		b.WriteString("\n\n")
	}
	for _, line := range d.bullets {
		b.WriteString(line)
		b.WriteByte('\n')
	}
	return b.String()
}

func place(width, height int, content string) string {
	return lipgloss.Place(width, height, lipgloss.Left, lipgloss.Top, content)
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
