package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/ildar7070/ssh-cv/internal/content"
)

// Layout constants.
const (
	// Minimum dimensions we render fully. Below this we fall back to a single
	// line so very small windows don't show garbage.
	minWidth  = 40
	minHeight = 12

	// 30/70 split for list-based tabs (CV, Projects).
	listPaneRatioNumerator   = 30
	listPaneRatioDenominator = 100
	minListWidth             = 18
	minDetailWidth           = 10
	listDetailGap            = 2

	// Padding on each side of the longest tab label.
	tabCellPadding = 2
	// Width of the "label" column on the contact page.
	contactLabelWidth = 10
)

func (m Model) View() string {
	if m.width < minWidth || m.height < minHeight {
		return "ssh-cv — resize your terminal (min 40×12)"
	}

	if m.mode == modeSplash {
		return m.splashView()
	}
	return m.appView()
}

// ─── Splash ────────────────────────────────────────────────────────────

func (m Model) splashView() string {
	s := m.styles
	title := s.SplashTitle.Render(m.profile.Splash.Title)
	cta := m.profile.Splash.CTA
	var hint string
	if m.splashBlink {
		hint = s.SplashHintHighlight.Render(cta)
	} else {
		hint = s.SplashHint.Render(cta)
	}
	block := lipgloss.JoinVertical(lipgloss.Center, title, "", hint)
	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center, block)
}

// ─── App shell ─────────────────────────────────────────────────────────

func (m Model) appView() string {
	s := m.styles
	inner := s.Doc.GetHorizontalFrameSize()
	innerW := m.width - inner
	if innerW < 1 {
		innerW = 1
	}

	header := m.headerView(innerW)
	footer := m.footerView(innerW)

	bodyH := m.height -
		s.Doc.GetVerticalFrameSize() -
		lipgloss.Height(header) -
		lipgloss.Height(footer)
	if bodyH < 3 {
		bodyH = 3
	}

	body := m.bodyView(innerW, bodyH)
	return s.Doc.Render(lipgloss.JoinVertical(lipgloss.Left, header, body, footer))
}

// ─── Header ────────────────────────────────────────────────────────────

func (m Model) headerView(width int) string {
	s := m.styles
	// Compute the widest tab label so every tab gets the same rendered width
	// and labels can be centred inside their cell.
	maxLabel := 0
	labels := make([]string, len(m.tabs))
	for i, t := range m.tabs {
		labels[i] = fmt.Sprintf("%d·%s", i+1, t.Label)
		if n := lipgloss.Width(labels[i]); n > maxLabel {
			maxLabel = n
		}
	}
	cellW := maxLabel + tabCellPadding

	parts := make([]string, 0, len(m.tabs)*2)
	for i := range m.tabs {
		if i > 0 {
			parts = append(parts, s.TabsSep)
		}
		style := s.Tab
		if i == m.activeTab {
			style = s.TabActive
		}
		parts = append(parts, style.Width(cellW).Align(lipgloss.Center).Render(labels[i]))
	}
	tabs := lipgloss.JoinHorizontal(lipgloss.Top, parts...)
	rule := strings.Repeat("─", width)
	return lipgloss.JoinVertical(lipgloss.Left,
		tabs,
		s.DetailMuted.Render(rule),
	)
}

// ─── Body dispatch ─────────────────────────────────────────────────────

func (m Model) bodyView(width, height int) string {
	switch m.currentTab() {
	case content.TabStart:
		return m.startView(width, height)
	case content.TabCV:
		return m.listDetailView(width, height,
			cvListItems(m.profile.CV), m.selection[content.TabCV],
			renderCVDetail(m.styles, m.profile.CV, m.selection[content.TabCV]))
	case content.TabProjects:
		return m.listDetailView(width, height,
			projectListItems(m.profile.Projects), m.selection[content.TabProjects],
			renderProjectDetail(m.styles, m.profile.Projects, m.selection[content.TabProjects]))
	case content.TabContact:
		return m.contactView(width, height)
	}
	return ""
}

// ─── Footer ────────────────────────────────────────────────────────────

func (m Model) footerView(width int) string {
	s := m.styles
	var hint string
	switch m.currentTab() {
	case content.TabCV, content.TabProjects:
		hint = "↑/↓ navigate · tab / 1-9 switch · q quit"
	default:
		hint = "tab / 1-9 switch · q quit"
	}
	rule := strings.Repeat("─", width)
	return lipgloss.JoinVertical(lipgloss.Left,
		s.DetailMuted.Render(rule),
		s.Footer.Render(hint),
	)
}

// ─── Start tab ─────────────────────────────────────────────────────────

func (m Model) startView(width, height int) string {
	s := m.styles
	p := m.profile
	heading := s.DetailHeading.Render(p.Name)
	tagline := s.DetailSub.Render(p.Tagline)
	var about strings.Builder
	for i, l := range p.About.Lines {
		if i > 0 {
			about.WriteByte('\n')
		}
		about.WriteString(s.DetailBody.Render(l))
	}
	block := lipgloss.JoinVertical(lipgloss.Left,
		heading,
		tagline,
		"",
		about.String(),
	)
	return lipgloss.Place(width, height, lipgloss.Left, lipgloss.Top, block)
}

// ─── Contact tab ───────────────────────────────────────────────────────

func (m Model) contactView(width, height int) string {
	s := m.styles
	rows := []string{
		s.DetailHeading.Render("Get in touch"),
		"",
	}
	for _, link := range m.profile.Contact {
		if link.Value == "" {
			continue
		}
		rows = append(rows,
			s.DetailSub.Render(fmt.Sprintf("%-*s", contactLabelWidth, link.Label))+
				s.DetailLink.Render(link.Value),
		)
	}
	block := lipgloss.JoinVertical(lipgloss.Left, rows...)
	return lipgloss.Place(width, height, lipgloss.Left, lipgloss.Top, block)
}

// ─── Generic list/detail (CV, Projects) ────────────────────────────────

func (m Model) listDetailView(width, height int, items []string, selected int, detail string) string {
	listW := width * listPaneRatioNumerator / listPaneRatioDenominator
	if listW < minListWidth {
		listW = minListWidth
	}
	detailW := width - listW - listDetailGap
	if detailW < minDetailWidth {
		detailW = minDetailWidth
	}

	list := m.renderList(items, selected, listW, height)
	det := lipgloss.NewStyle().
		Width(detailW).
		Height(height).
		Render(detail)

	return lipgloss.JoinHorizontal(lipgloss.Top, list, strings.Repeat(" ", listDetailGap), det)
}

func (m Model) renderList(items []string, selected, width, height int) string {
	s := m.styles
	rows := make([]string, 0, len(items))
	for i, it := range items {
		marker := "  "
		style := s.ListItem
		if i == selected {
			marker = "▸ "
			style = s.ListItemSelected
		}
		line := ansi.Truncate(marker+it, width-2, "…")
		rows = append(rows, style.Width(width).Render(line))
	}
	out := lipgloss.JoinVertical(lipgloss.Left, rows...)
	return lipgloss.NewStyle().Width(width).Height(height).Render(out)
}

// ─── Detail renderers ──────────────────────────────────────────────────

type detailBlock struct {
	heading  string
	subtitle string // optional
	meta     string // optional, rendered muted
	body     string // optional paragraph between header and bullets
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

func cvListItems(cv []content.CVEntry) []string {
	out := make([]string, len(cv))
	for i, e := range cv {
		out[i] = e.Role
	}
	return out
}

func renderCVDetail(s Styles, cv []content.CVEntry, idx int) string {
	if idx < 0 || idx >= len(cv) {
		return s.DetailMuted.Render("(no entry)")
	}
	e := cv[idx]
	bullets := make([]string, 0, len(e.Bullets))
	for _, bul := range e.Bullets {
		bullets = append(bullets, s.DetailMuted.Render("• ")+s.DetailBody.Render(bul))
	}
	return renderDetailBlock(s, detailBlock{
		heading:  e.Role,
		subtitle: e.Company,
		meta:     e.Period,
		body:     e.Summary,
		bullets:  bullets,
	})
}

func projectListItems(ps []content.Project) []string {
	out := make([]string, len(ps))
	for i, p := range ps {
		out[i] = p.Name
	}
	return out
}

func renderProjectDetail(s Styles, ps []content.Project, idx int) string {
	if idx < 0 || idx >= len(ps) {
		return s.DetailMuted.Render("(no entry)")
	}
	p := ps[idx]
	bullets := make([]string, 0, len(p.Details))
	for _, line := range p.Details {
		bullets = append(bullets, s.DetailBody.Render(line))
	}
	return renderDetailBlock(s, detailBlock{
		heading:  p.Name,
		subtitle: p.Tagline,
		meta:     p.URL,
		bullets:  bullets,
	})
}
