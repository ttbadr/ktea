package kadmin

import (
	kgo "github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestListTopicConfigs(t *testing.T) {
	t.Run("List Topic Configs", func(t *testing.T) {
		topic := topicName()
		// given
		createTopic(t, []kgo.TopicConfig{
			{
				Topic:             topic,
				NumPartitions:     2,
				ReplicationFactor: 1,
			},
		})

		//when
		msg := ka.ListConfigs(topic).(TopicConfigListingStartedMsg)

		// then
		var configs map[string]string
		select {
		case c := <-msg.Configs:
			configs = c
		case e := <-msg.Err:
			assert.Fail(t, "Failed to list configs", e)
			return
		}
		assert.Equal(t, "delete", configs["cleanup.policy"])

		// clean up
		ka.DeleteTopic(topic)
	})
}
