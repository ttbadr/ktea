package tabs

import (
	"ktea/ui"
)

type TabName int

var TopicsTab TabName = 0
var ClustersTab TabName = 2

type TabController interface {
	ui.View
}
