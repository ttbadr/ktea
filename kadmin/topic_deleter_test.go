package kadmin

import (
	kgo "github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestDeleteTopic(t *testing.T) {
	t.Run("Delete", func(t *testing.T) {
		t.Run("existing topic", func(t *testing.T) {
			topic1 := topicName()
			topic2 := topicName()
			// given
			createTopic(t, []kgo.TopicConfig{
				{
					Topic:             topic1,
					NumPartitions:     2,
					ReplicationFactor: 1,
				},
				{
					Topic:             topic2,
					NumPartitions:     1,
					ReplicationFactor: 1,
				},
			})

			// when
			msg := ka.DeleteTopic(topic2)

			// then
			assert.IsType(t, TopicDeletedMsg{}, msg)
			assert.EventuallyWithT(t, func(c *assert.CollectT) {
				listTopicsMsg := ka.ListTopics().(TopicListingStartedMsg)
				var topics []Topic
				select {
				case topics = <-listTopicsMsg.Topics:
				case err := <-listTopicsMsg.Err:
					t.Error(t, "Failed to list topics", err)
				}
				assert.Contains(c, topics, Topic{topic1, 2, 1, 0})
				assert.NotContains(c, topics, Topic{topic2, 2, 1, 0})
			}, 2*time.Second, 10*time.Millisecond)
			// clean up
			ka.DeleteTopic(topic1)
		})

		t.Run("none existing topic", func(t *testing.T) {
			topic1 := topicName()
			topic2 := topicName()
			// given
			createTopic(t, []kgo.TopicConfig{
				{
					Topic:             topic1,
					NumPartitions:     2,
					ReplicationFactor: 1,
				},
				{
					Topic:             topic2,
					NumPartitions:     1,
					ReplicationFactor: 1,
				},
			})

			// when
			noneExistingTopic := topicName()
			msg := ka.DeleteTopic(noneExistingTopic)

			// then
			assert.IsType(t, KAdminErrorMsg{}, msg)

			// clean up
			ka.DeleteTopic(topic1)
			ka.DeleteTopic(topic2)
			ka.DeleteTopic(noneExistingTopic)
		})

	})
}
