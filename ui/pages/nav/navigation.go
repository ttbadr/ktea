package nav

import (
	"ktea/kadmin"
	"ktea/sradmin"
	"ktea/ui"
	"ktea/ui/components/statusbar"
)

type Page interface {
	ui.View
	statusbar.Provider
}

type LoadTopicsPageMsg struct {
	Refresh bool
}

type LoadCreateTopicPageMsg struct{}

type LoadTopicConfigPageMsg struct{}

type LoadPublishPageMsg struct {
	Topic *kadmin.Topic
}

type LoadConsumptionPageMsg struct {
	ReadDetails kadmin.ReadDetails
}

type LoadCachedConsumptionPageMsg struct {
}

type LoadConsumptionFormPageMsg struct {
	Topic       *kadmin.Topic
	ReadDetails *kadmin.ReadDetails
}

type LoadRecordDetailPageMsg struct {
	Record *kadmin.ConsumerRecord
	Topic  *kadmin.Topic
}

type LoadCGroupsPageMsg struct {
}

type LoadCGroupTopicsPageMsg struct {
	GroupName string
}

type LoadCreateSubjectPageMsg struct{}

type LoadSubjectsPageMsg struct {
	Refresh bool
}

type LoadSchemaDetailsPageMsg struct {
	Subject sradmin.Subject
}
