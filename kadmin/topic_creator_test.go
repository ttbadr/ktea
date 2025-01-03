package kadmin

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateTopic(t *testing.T) {
	t.Run("Create new topic", func(t *testing.T) {
		topic := topicName()
		// when
		topicCreatedMsg := ka.CreateTopic(TopicCreationDetails{topic, 2, map[string]string{}}).(TopicCreationStartedMsg)

		select {
		case <-topicCreatedMsg.Created:
		case err := <-topicCreatedMsg.Err:
			t.Error("Unable to create Topic", err)
			return
		}

		// then
		listTopicsMsg := ka.ListTopics().(TopicListingStartedMsg)
		var topics []Topic
		select {
		case topics = <-listTopicsMsg.Topics:
		case <-listTopicsMsg.Err:
			assert.Fail(t, "Failed to list topics")
		}
		assert.Contains(t, topics, Topic{topic, 2, 1, 0})

		// clean up
		ka.DeleteTopic(topic)
	})
}
