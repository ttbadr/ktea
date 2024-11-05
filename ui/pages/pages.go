package pages

import (
	"ktea/kadmin"
	"ktea/ui"
	"ktea/ui/components/statusbar"
)

type Page interface {
	ui.View
	statusbar.Provider
}

type LoadTopicsPageMsg struct {
	Reload bool
}

type LoadCreateTopicPageMsg struct{}

type LoadTopicConfigPageMsg struct{}

type LoadPublishPageMsg struct {
	Topic kadmin.Topic
}

type LoadConsumptionPageMsg struct {
	Topic kadmin.Topic
}

type LoadConsumptionFormPageMsg struct {
}

type LoadCGroupsPageMsg struct {
}

type LoadCGroupTopicsPageMsg struct {
	GroupName string
}

type LoadCreateSubjectPageMsg struct{}

type LoadSubjectsPageMsg struct{}

type LoadSchemaDetailsPageMsg struct {
	Subject string
}
