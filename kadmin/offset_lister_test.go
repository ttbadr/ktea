package kadmin

import (
	"context"
	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestConsumerGroupOffsets(t *testing.T) {
	t.Run("List Offsets", func(t *testing.T) {
		topic := topicName()
		// given
		ka.CreateTopic(TopicCreationDetails{
			Name:          topic,
			NumPartitions: 1,
			Properties:    nil,
		})

		for i := 0; i < 10; i++ {
			ka.PublishRecord(&ProducerRecord{
				Key:       "key",
				Value:     "value",
				Topic:     topic,
				Partition: nil,
			})
		}

		groupName := "offset-test-group"
		consumerGroup, err := sarama.NewConsumerGroupFromClient(groupName, kafkaClient())
		if err != nil {
			t.Fatal("Unable to create Consumer Group.", err)
		}

		handler := testConsumer{ExpectedMsgCount: 10}
		consumerGroup.Consume(context.WithoutCancel(context.Background()), []string{topic}, &handler)

		defer consumerGroup.Close()

		offsetListingStartedMsg := ka.ListOffsets(groupName).(OffsetListingStartedMsg)

		select {
		case offsets := <-offsetListingStartedMsg.Offsets:
			assert.NotNil(t, offsets)
			assert.Len(t, offsets, 1)
			assert.Equal(t, offsets[0], TopicPartitionOffset{
				Topic:     topic,
				Partition: 0,
				Offset:    9,
			})
		case err := <-offsetListingStartedMsg.Err:
			t.Fatal("Error while listing offsets", err)
		case <-time.After(5 * time.Second):
			t.Fatal("Test timed out waiting for offsets")
		}
	})
}
