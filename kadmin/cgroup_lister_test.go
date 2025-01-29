package kadmin

import (
	"context"
	"fmt"
	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestConsumerGroups(t *testing.T) {
	t.Run("List groups", func(t *testing.T) {
		topic := topicName()
		// given
		msg := ka.CreateTopic(TopicCreationDetails{
			Name:          topic,
			NumPartitions: 1,
			Properties:    nil,
		}).(TopicCreationStartedMsg)

		switch msg := msg.AwaitCompletion().(type) {
		case TopicCreatedMsg:
		case TopicCreationErrMsg:
			t.Fatal("Unable to create topic", msg.Err)
		}

		for i := 0; i < 10; i++ {
			ka.PublishRecord(&ProducerRecord{
				Key:       "key",
				Value:     "value",
				Topic:     topic,
				Partition: nil,
			})
		}

		expectedGroups := make(map[string]bool)
		for i := 0; i < 10; i++ {
			groupName := fmt.Sprintf("test-group-%d", i)
			expectedGroups[groupName] = false
			consumerGroup, err := sarama.NewConsumerGroupFromClient(groupName, kafkaClient())
			if err != nil {
				t.Fatal("Unable to create Consumer Group.", err)
			}

			handler := testConsumer{ExpectedMsgCount: 10}
			consumerGroup.Consume(context.WithoutCancel(context.Background()), []string{topic}, &handler)

			defer consumerGroup.Close()
		}

		cgroupListingStartedMsg := ka.ListCGroups().(ConsumerGroupListingStartedMsg)

		select {
		case groups := <-cgroupListingStartedMsg.ConsumerGroups:
			assert.Len(t, groups, 10, "Expected 10 consumer groups")

			// Verify that all expected groups are present
			for _, group := range groups {
				if _, exists := expectedGroups[group.Name]; exists {
					assert.NotEmpty(t, group.Members)
					assert.NotEmpty(t, group.Members[0].MemberId)
					assert.NotEmpty(t, group.Members[0].ClientId)
					assert.NotEmpty(t, group.Members[0].ClientHost)
					expectedGroups[group.Name] = true
				}
			}

			// Check that all groups in `expectedGroups` were found
			for groupName, found := range expectedGroups {
				assert.True(t, found, "Consumer group '%s' was not found", groupName)
			}
		case err := <-msg.Err:
			t.Fatal("Error while listing groups", err)
		case <-time.After(5 * time.Second):
			t.Fatal("Test timed out waiting for consumer groups")
		}
	})
}
