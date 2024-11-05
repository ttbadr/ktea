package cgroups_tab

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"ktea/config"
	"ktea/kadmin"
	"ktea/kontext"
	"ktea/ui"
	"testing"
)

type MockConsumerGroupOffsetLister struct{}

func (m *MockConsumerGroupOffsetLister) ListOffsets(group string) tea.Msg {
	return nil
}

type MockConsumerGroupLister struct{}

func (m *MockConsumerGroupLister) ListConsumerGroups() tea.Msg {
	return nil
}

func TestGroupsTab(t *testing.T) {
	t.Run("List consumer groups", func(t *testing.T) {
		groupsTab, _ := New(&MockConsumerGroupLister{}, &MockConsumerGroupOffsetLister{})

		groupsTab.Update(kadmin.ConsumerGroupsListedMsg{
			ConsumerGroups: []*kadmin.ConsumerGroup{
				{
					Name: "Group1",
					Members: []kadmin.GroupMember{
						{
							MemberId:   "Group1Id1",
							ClientId:   "Group1ClientId1",
							ClientHost: "127.0.0.1",
						},
					},
				},
				{
					Name:    "Group2",
					Members: nil,
				},
			},
		})

		render := ansi.Strip(groupsTab.View(&kontext.ProgramKtx{
			WindowWidth:     100,
			WindowHeight:    100,
			AvailableHeight: 100,
			Config: &config.Config{
				Clusters: []config.Cluster{
					{
						Name:             "PRD",
						BootstrapServers: []string{"localhost:9092"},
						SASLConfig:       nil,
					},
				},
			},
		}, ui.TestRenderer))

		assert.Contains(t, render, "Group1")
		assert.Contains(t, render, "Group2")
	})
}
