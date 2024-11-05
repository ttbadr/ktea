package ui

import "github.com/charmbracelet/lipgloss"

func JoinVerticalSkipEmptyViews(views ...string) string {
	var renderViews []string
	for _, view := range views {
		if view != "" {
			renderViews = append(renderViews, view)
		}
	}
	return lipgloss.JoinVertical(lipgloss.Top, renderViews...)
}
