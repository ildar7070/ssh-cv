package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/ildar7070/i12k/internal/content"
)

// Minimum dimensions we render fully. Below this we fall back to a single
// line so very small windows don't show garbage.
const (
	minWidth  = 40
	minHeight = 12
)

func (m Model) View() string {
	if m.width < minWidth || m.height < minHeight {
		return "i12k — resize your terminal (min 40×12)"
	}

	if m.mode == modeSplash {
		return m.splashView()
	}
	return m.appView()
}

// ─── Splash ────────────────────────────────────────────────────────────

func (m Model) splashView() string {
	s := m.styles
	title := s.SplashTitle.Render("i12k")
	cta := "Press Enter to start"
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
	// and labels can be centred inside their cell. +2 leaves one space of
	// breathing room on each side of the longest label.
	maxLabel := 0
	labels := make([]string, len(allTabs))
	for i, t := range allTabs {
		labels[i] = fmt.Sprintf("%d·%s", i+1, t.String())
		if n := lipgloss.Width(labels[i]); n > maxLabel {
			maxLabel = n
		}
	}
	cellW := maxLabel + 2

	parts := make([]string, 0, len(allTabs)*2)
	for i, t := range allTabs {
		if i > 0 {
			parts = append(parts, s.TabsSep)
		}
		style := s.Tab
		if t == m.tab {
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
	switch m.tab {
	case tabStart:
		return m.startView(width, height)
	case tabCV:
		return m.listDetailView(width, height,
			cvListItems(m.profile.CV), m.cvIdx,
			renderCVDetail(m.styles, m.profile.CV, m.cvIdx))
	case tabProjects:
		return m.listDetailView(width, height,
			projectListItems(m.profile.Projects), m.projectsIdx,
			renderProjectDetail(m.styles, m.profile.Projects, m.projectsIdx))
	case tabContact:
		return m.contactView(width, height)
	}
	return ""
}

// ─── Footer ────────────────────────────────────────────────────────────

func (m Model) footerView(width int) string {
	s := m.styles
	var hint string
	switch m.tab {
	case tabCV, tabProjects:
		hint = "↑/↓ navigate · tab / 1-4 switch · q quit"
	default:
		hint = "tab / 1-4 switch · q quit"
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
	c := m.profile.Contact
	row := func(label, value string) string {
		if value == "" {
			return ""
		}
		return s.DetailSub.Render(fmt.Sprintf("%-10s", label)) +
			s.DetailLink.Render(value)
	}
	rows := []string{
		s.DetailHeading.Render("Get in touch"),
		"",
		row("email", c.Email),
		row("github", c.GitHub),
		row("linkedin", c.LinkedIn),
		row("instagram", c.Instagram),
	}
	block := lipgloss.JoinVertical(lipgloss.Left, rows...)
	return lipgloss.Place(width, height, lipgloss.Left, lipgloss.Top, block)
}

// ─── Generic list/detail (CV, Projects) ────────────────────────────────

func (m Model) listDetailView(width, height int, items []string, selected int, detail string) string {
	listW := width * 30 / 100
	if listW < 18 {
		listW = 18
	}
	gapW := 2
	detailW := width - listW - gapW
	if detailW < 10 {
		detailW = 10
	}

	list := m.renderList(items, selected, listW, height)
	det := lipgloss.NewStyle().
		Width(detailW).
		Height(height).
		Render(detail)

	return lipgloss.JoinHorizontal(lipgloss.Top, list, strings.Repeat(" ", gapW), det)
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
		line := truncate(marker+it, width-2)
		rows = append(rows, style.Width(width).Render(line))
	}
	out := lipgloss.JoinVertical(lipgloss.Left, rows...)
	return lipgloss.NewStyle().Width(width).Height(height).Render(out)
}

// ─── Detail renderers ──────────────────────────────────────────────────

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
	var b strings.Builder
	b.WriteString(s.DetailHeading.Render(e.Role))
	b.WriteByte('\n')
	b.WriteString(s.DetailSub.Render(e.Company))
	b.WriteString("   ")
	b.WriteString(s.DetailMuted.Render(e.Period))
	b.WriteString("\n\n")
	if e.Summary != "" {
		b.WriteString(s.DetailBody.Render(e.Summary))
		b.WriteString("\n\n")
	}
	for _, bul := range e.Bullets {
		b.WriteString(s.DetailMuted.Render("• "))
		b.WriteString(s.DetailBody.Render(bul))
		b.WriteByte('\n')
	}
	return b.String()
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
	var b strings.Builder
	b.WriteString(s.DetailHeading.Render(p.Name))
	b.WriteByte('\n')
	if p.Tagline != "" {
		b.WriteString(s.DetailSub.Render(p.Tagline))
		b.WriteByte('\n')
	}
	if p.URL != "" {
		b.WriteString(s.DetailMuted.Render(p.URL))
		b.WriteByte('\n')
	}
	b.WriteByte('\n')
	for _, l := range p.Details {
		b.WriteString(s.DetailBody.Render(l))
		b.WriteByte('\n')
	}
	return b.String()
}

// truncate cuts a string to max display width, adding an ellipsis if needed.
func truncate(s string, max int) string {
	if max <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	if max == 1 {
		return "…"
	}
	return string(r[:max-1]) + "…"
}
