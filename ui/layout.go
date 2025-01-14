package ui

import "github.com/charmbracelet/lipgloss"

func JoinVerticalSkipEmptyViews(position lipgloss.Position, views ...string) string {
	return join(views, position)
}

func JoinHorizontalSkipEmptyViews(views ...string) string {
	return join(views, lipgloss.Center)
}

func join(views []string, position lipgloss.Position) string {
	var renderViews []string
	for _, view := range views {
		if view != "" {
			renderViews = append(renderViews, view)
		}
	}
	return lipgloss.JoinVertical(position, renderViews...)
}
