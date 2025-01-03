package kadmin

import tea "github.com/charmbracelet/bubbletea"

type CGroupLister interface {
	ListConsumerGroups() tea.Msg
}

type ConsumerGroup struct {
	Name    string
	Members []GroupMember
}

type ConsumerGroupListingStartedMsg struct {
	Err            chan error
	ConsumerGroups chan []*ConsumerGroup
}

func (msg *ConsumerGroupListingStartedMsg) AwaitCompletion() tea.Msg {
	select {
	case groups := <-msg.ConsumerGroups:
		return ConsumerGroupsListedMsg{groups}
	case err := <-msg.Err:
		return ConsumerGroupListingErrorMsg{err}
	}
}

type ConsumerGroupsListedMsg struct {
	ConsumerGroups []*ConsumerGroup
}

type ConsumerGroupListingErrorMsg struct {
	Err error
}

func (ka *SaramaKafkaAdmin) ListConsumerGroups() tea.Msg {
	errChan := make(chan error)
	groupsChan := make(chan []*ConsumerGroup)

	go ka.doListConsumerGroups(groupsChan, errChan)

	return ConsumerGroupListingStartedMsg{errChan, groupsChan}
}

func (ka *SaramaKafkaAdmin) doListConsumerGroups(groupsChan chan []*ConsumerGroup, errorChan chan error) {
	maybeIntroduceLatency()
	if listGroupResponse, err := ka.admin.ListConsumerGroups(); err != nil {
		errorChan <- err
	} else {
		var consumerGroups []*ConsumerGroup
		var groupNames []string
		var groupByName = make(map[string]*ConsumerGroup)

		for name, _ := range listGroupResponse {
			consumerGroup := ConsumerGroup{Name: name}
			consumerGroups = append(consumerGroups, &consumerGroup)
			groupByName[name] = &consumerGroup
			groupNames = append(groupNames, name)
		}

		describeConsumerGroupResponse, err := ka.admin.DescribeConsumerGroups(groupNames)
		if err != nil {
			errorChan <- err
			return
		}

		for _, groupDescription := range describeConsumerGroupResponse {
			group := groupByName[groupDescription.GroupId]
			var groupMembers []GroupMember
			for _, m := range groupDescription.Members {
				member := GroupMember{}
				member.MemberId = m.MemberId
				member.ClientId = m.ClientId
				member.ClientHost = m.ClientHost
				groupMembers = append(groupMembers, member)
			}
			group.Members = groupMembers
		}
		groupsChan <- consumerGroups
	}
}
