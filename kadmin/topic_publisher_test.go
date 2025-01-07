package kadmin

import (
	"context"
	kgo "github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestPublish(t *testing.T) {
	t.Run("Publish a text record", func(t *testing.T) {
		topic := topicName()
		// given
		createTopic(t, []kgo.TopicConfig{
			{
				Topic:             topic,
				NumPartitions:     1,
				ReplicationFactor: 1,
			},
		})

		// when
		assert.EventuallyWithT(t, func(c *assert.CollectT) {
			psm := ka.PublishRecord(&ProducerRecord{
				Topic: topic,
				Key:   "123",
				Value: "{\"id\":\"123\"}",
			})

			select {
			case err := <-psm.Err:
				t.Fatal(c, "Unable to publish", err)
			case p := <-psm.Published:
				assert.True(c, p)
			}
		}, 10*time.Second, 10*time.Millisecond)

		// then
		ctx, cancel := context.WithCancel(context.Background())
		rsm := ka.ReadRecords(ctx, ReadDetails{
			Topic: &Topic{topic, 1, 1, 1},
		})
		assert.EventuallyWithT(t, func(c *assert.CollectT) {
			assert.Equal(c, "{\"id\":\"123\"}", (<-rsm.ConsumerRecord).Value)
		}, 2*time.Second, 10*time.Millisecond)

		// clean up
		cancel()
		ka.DeleteTopic(topic)
	})

	t.Run("Publish to specific partition", func(t *testing.T) {
		topic := topicName()
		// given
		ka.CreateTopic(TopicCreationDetails{
			Name:          topic,
			NumPartitions: 3,
		})

		// when
		assert.EventuallyWithT(t, func(c *assert.CollectT) {
			var partition = 2
			psm := ka.PublishRecord(&ProducerRecord{
				Topic:     topic,
				Key:       "123",
				Value:     "{\"id\":\"123\"}",
				Partition: &partition,
			})

			select {
			case err := <-psm.Err:
				t.Fatal(c, "Unable to publish", err)
			case p := <-psm.Published:
				assert.True(c, p)
			}
		}, 10*time.Second, 10*time.Millisecond)

		// then
		rsm := ka.ReadRecords(context.Background(), ReadDetails{
			Topic:      &Topic{topic, 2, 1, 1},
			Partitions: []int{2},
		})
		assert.EventuallyWithT(t, func(c *assert.CollectT) {
			record := <-rsm.ConsumerRecord
			assert.Equal(c, "{\"id\":\"123\"}", record.Value)
			assert.Equal(c, int64(2), record.Partition)
		}, 5*time.Second, 10*time.Millisecond)

		// clean up
		ka.DeleteTopic(topic)
	})

}
