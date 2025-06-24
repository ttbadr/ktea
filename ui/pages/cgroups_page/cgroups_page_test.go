package cgroups_page

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"ktea/kadmin"
	"ktea/tests"
	"strings"
	"testing"
)

type MockCGroupLister struct {
}

func (m MockCGroupLister) ListCGroups() tea.Msg {
	return nil
}

type MockCGroupDeleter struct {
}

type MockCGroupDeletionStartedMsg struct{}

func (m MockCGroupDeleter) DeleteCGroup(name string) tea.Msg {
	return MockCGroupDeletionStartedMsg{}
}

func TestCgroupsPage(t *testing.T) {
	t.Run("Default sort by Consumer Group Asc", func(t *testing.T) {
		page, _ := New(&MockCGroupLister{}, &MockCGroupDeleter{})

		_ = page.Update(kadmin.ConsumerGroupsListedMsg{
			ConsumerGroups: []*kadmin.ConsumerGroup{
				{
					Name:    "group3",
					Members: []kadmin.GroupMember{},
				},
				{
					Name:    "group1",
					Members: []kadmin.GroupMember{},
				},
				{
					Name:    "group2",
					Members: []kadmin.GroupMember{},
				},
			},
		})

		render := page.View(tests.NewKontext(), tests.TestRenderer)

		assert.Contains(t, render, "▲ Consumer Group")

		g1Idx := strings.Index(render, "group1")
		g2Idx := strings.Index(render, "group2")
		g3Idx := strings.Index(render, "group3")

		assert.Less(t, g1Idx, g2Idx)
		assert.Less(t, g1Idx, g3Idx)
		assert.Less(t, g2Idx, g3Idx)
	})

	t.Run("Toggle sort by Consumer Group Desc", func(t *testing.T) {
		page, _ := New(&MockCGroupLister{}, &MockCGroupDeleter{})

		_ = page.Update(kadmin.ConsumerGroupsListedMsg{
			ConsumerGroups: []*kadmin.ConsumerGroup{
				{
					Name:    "group3",
					Members: []kadmin.GroupMember{},
				},
				{
					Name:    "group1",
					Members: []kadmin.GroupMember{},
				},
				{
					Name:    "group2",
					Members: []kadmin.GroupMember{},
				},
			},
		})

		page.Update(tests.Key(tea.KeyF3))
		page.Update(tests.Key(tea.KeyEnter))
		render := page.View(tests.NewKontext(), tests.TestRenderer)

		assert.Contains(t, render, "▼ Consumer Group")

		g1Idx := strings.Index(render, "group1")
		g2Idx := strings.Index(render, "group2")
		g3Idx := strings.Index(render, "group3")

		assert.Less(t, g3Idx, g1Idx)
		assert.Less(t, g3Idx, g2Idx)
		assert.Less(t, g2Idx, g1Idx)
	})

	t.Run("Toggle sort by Members", func(t *testing.T) {
		page, _ := New(&MockCGroupLister{}, &MockCGroupDeleter{})

		_ = page.Update(kadmin.ConsumerGroupsListedMsg{
			ConsumerGroups: []*kadmin.ConsumerGroup{
				{
					Name: "group3",
					Members: []kadmin.GroupMember{
						{
							MemberId:   "g3Id1",
							ClientId:   "",
							ClientHost: "",
						},
						{
							MemberId:   "g3Id2",
							ClientId:   "",
							ClientHost: "",
						},
						{
							MemberId:   "g3Id3",
							ClientId:   "",
							ClientHost: "",
						},
					},
				},
				{
					Name: "group1",
					Members: []kadmin.GroupMember{
						{
							MemberId:   "g1Id1",
							ClientId:   "",
							ClientHost: "",
						},
					},
				},
				{
					Name: "group2",
					Members: []kadmin.GroupMember{
						{
							MemberId:   "g2Id1",
							ClientId:   "",
							ClientHost: "",
						},
						{
							MemberId:   "g2Id2",
							ClientId:   "",
							ClientHost: "",
						},
					},
				},
			},
		})

		page.Update(tests.Key(tea.KeyF3))
		page.Update(tests.Key(tea.KeyRight))
		page.Update(tests.Key(tea.KeyEnter))
		render := page.View(tests.NewKontext(), tests.TestRenderer)

		assert.Contains(t, render, "▼ Members")

		g1Idx := strings.Index(render, "group1")
		g2Idx := strings.Index(render, "group2")
		g3Idx := strings.Index(render, "group3")

		assert.Less(t, g3Idx, g1Idx)
		assert.Less(t, g3Idx, g2Idx)
		assert.Less(t, g2Idx, g1Idx)

		page.Update(tests.Key(tea.KeyEnter))
		render = page.View(tests.NewKontext(), tests.TestRenderer)

		assert.Contains(t, render, "▲ Members")

		g1Idx = strings.Index(render, "group1")
		g2Idx = strings.Index(render, "group2")
		g3Idx = strings.Index(render, "group3")

		assert.Less(t, g1Idx, g2Idx)
		assert.Less(t, g1Idx, g3Idx)
		assert.Less(t, g2Idx, g3Idx)
	})

	t.Run("Delete consumer group", func(t *testing.T) {
		page, _ := New(&MockCGroupLister{}, &MockCGroupDeleter{})

		_ = page.Update(kadmin.ConsumerGroupsListedMsg{
			ConsumerGroups: []*kadmin.ConsumerGroup{
				{
					Name:    "group3",
					Members: []kadmin.GroupMember{},
				},
				{
					Name:    "group1",
					Members: []kadmin.GroupMember{},
				},
				{
					Name:    "group2",
					Members: []kadmin.GroupMember{},
				},
			},
		})
		page.View(tests.NewKontext(), tests.TestRenderer)

		t.Run("F2 triggers version delete", func(t *testing.T) {
			page.Update(tests.Key(tea.KeyDown))
			page.Update(tests.Key(tea.KeyF2))

			render := page.View(tests.NewKontext(), tests.TestRenderer)

			assert.Regexp(t, "┃ 🗑️  group2 will be deleted permanently\\W+Delete!\\W+Cancel.", render)
		})

		t.Run("esc cancels deletion", func(t *testing.T) {
			page.Update(tests.Key(tea.KeyEsc))

			render := page.View(tests.NewKontext(), tests.TestRenderer)

			assert.NotRegexp(t, "┃ 🗑️  group2 will be deleted permanently\\W+Delete!\\W+Cancel.", render)
		})

		t.Run("selecting cancel cancels deletion", func(t *testing.T) {
			page.Update(tests.Key(tea.KeyF2))
			page.Update(tests.Key(tea.KeyEnter))

			render := page.View(tests.NewKontext(), tests.TestRenderer)

			assert.NotRegexp(t, "┃ 🗑️  group2 will be deleted permanently\\W+Delete!\\W+Cancel.", render)
		})

		t.Run("effectively delete schema", func(t *testing.T) {
			render := page.View(tests.NewKontext(), tests.TestRenderer)
			assert.Contains(t, render, "│ group2")

			page.Update(tests.Key(tea.KeyF2))
			page.Update(tests.Key('d'))
			cmd := page.Update(tests.Key(tea.KeyEnter))

			assert.IsType(t, MockCGroupDeletionStartedMsg{}, cmd())

			page.Update(kadmin.CGroupDeletedMsg{GroupName: "group2"})

			render = page.View(tests.NewKontext(), tests.TestRenderer)

			assert.NotContains(t, render, "│ group2")
		})
	})
}
