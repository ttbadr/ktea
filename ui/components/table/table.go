package table

import (
	"github.com/charmbracelet/bubbles/table"
	"ktea/styles"
)

func NewDefaultTable() table.Model {
	return table.New(
		table.WithFocused(true),
		table.WithStyles(styles.Table.Styles),
	)
}
