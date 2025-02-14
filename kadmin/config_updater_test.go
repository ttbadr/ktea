package kadmin

import (
	kgo "github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUpdateTopicConfig(t *testing.T) {
	t.Run("Update Topic Config", func(t *testing.T) {
		topic := topicName()
		// given
		createTopic(t, []kgo.TopicConfig{
			{
				Topic:             topic,
				NumPartitions:     2,
				ReplicationFactor: 1,
			},
		})

		// when
		ka.UpdateConfig(TopicConfigToUpdate{
			topic,
			"delete.retention.ms",
			"172800000"},
		)

		// then

		// then
		msg := ka.ListConfigs(topic).(TopicConfigListingStartedMsg)
		var configs map[string]string
		select {
		case c := <-msg.Configs:
			configs = c
		case e := <-msg.Err:
			assert.Fail(t, "Failed to list configs", e)
			return
		}
		assert.Equal(t, "172800000", configs["delete.retention.ms"])

		// clean up
		ka.DeleteTopic(topic)
	})

	t.Run("Invalid value for update", func(t *testing.T) {
		topic := topicName()
		// given
		createTopic(t, []kgo.TopicConfig{
			{
				Topic:             topic,
				NumPartitions:     2,
				ReplicationFactor: 1,
			},
		})

		// when
		msg := ka.UpdateConfig(TopicConfigToUpdate{
			topic,
			"delete.retention.ms",
			"-172800000"},
		)

		// then
		assert.IsType(t, KAdminErrorMsg{}, msg)
		assert.Equal(t, "kafka server: Configuration is invalid - Invalid value -172800000 for configuration delete.retention.ms: Value must be at least 0", msg.(KAdminErrorMsg).Error.Error())

		// clean up
		ka.DeleteTopic(topic)
	})

	//t.Run("Broker not available", func(t *testing.T) {
	//	topic := topicName()
	//	// when
	//	ska := NewSaramaKAdmin(ConnectionDetails{
	//		BootstrapServers: []string{"localpost:123"},
	//		SASLConfig:       nil,
	//	})
	//	msg := ska.UpdateConfig(TopicConfigToUpdate{
	//		topic,
	//		"delete.retention.ms",
	//		"-172800000"},
	//	)
	//
	//	// then
	//	assert.IsType(t, KAdminErrorMsg{}, msg)
	//	assert.Equal(t, "Broker Not Available: not a client facing error and is used mostly by tools when a broker is not alive", msg.(KAdminErrorMsg).Error.Error())
	//})
}
