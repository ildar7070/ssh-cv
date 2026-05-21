package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/ildar7070/ssh-cv/internal/tui/sections"
)

// Layout constants.
const (
	// Minimum dimensions we render fully. Below this we fall back to a single
	// line so very small windows don't show garbage.
	minWidth  = 40
	minHeight = 12

	// Padding on each side of the longest tab label.
	tabCellPadding = 2

	defaultFooterHint = "←/→ switch tab · t ◑ · q quit"
)

func (m Model) View() tea.View {
	return tea.View{Content: m.content(), AltScreen: true}
}

func (m Model) content() string {
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
	s := m.styles()
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
	s := m.styles()
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
	s := m.styles()
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
	cur, ok := m.currentSection()
	if !ok {
		return ""
	}
	r, ok := sections.Get(cur.Type)
	if !ok {
		return m.styles().DetailMuted.Render(
			fmt.Sprintf("(no renderer for type %q)", cur.Type))
	}
	return r.Render(cur, sections.RenderContext{
		Profile:  m.profile,
		Styles:   m.styles(),
		Width:    width,
		Height:   height,
		Selected: m.selection[cur.ID],
	})
}

// ─── Footer ────────────────────────────────────────────────────────────

func (m Model) footerView(width int) string {
	s := m.styles()
	hint := defaultFooterHint
	if cur, ok := m.currentSection(); ok {
		if r, ok := sections.Get(cur.Type); ok {
			if h := r.FooterHint(); h != "" {
				hint = h
			}
		}
	}
	rule := strings.Repeat("─", width)
	return lipgloss.JoinVertical(lipgloss.Left,
		s.DetailMuted.Render(rule),
		s.Footer.Render(hint),
	)
}
