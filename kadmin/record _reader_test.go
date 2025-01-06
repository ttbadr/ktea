package kadmin

import (
	"context"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slices"
	"strconv"
	"testing"
	"time"
)

func TestReadRecords(t *testing.T) {
	t.Run("Read from beginning with specific limit", func(t *testing.T) {
		t.Run("with one partition", func(t *testing.T) {
			topic := topicName()
			// given
			ka.CreateTopic(TopicCreationDetails{
				Name:          topic,
				NumPartitions: 1,
			})

			// when
			assert.EventuallyWithT(t, func(c *assert.CollectT) {
				for i := 0; i < 55; i++ {
					psm := ka.PublishRecord(&ProducerRecord{
						Topic: topic,
						Key:   strconv.Itoa(i),
						Value: "{\"id\":\"123\"}",
					})

					select {
					case err := <-psm.Err:
						t.Fatal(c, "Unable to publish", err)
					case p := <-psm.Published:
						assert.True(c, p)
					}
				}
			}, 10*time.Second, 10*time.Millisecond)

			// then
			rsm := ka.ReadRecords(context.Background(), ReadDetails{
				Topic:      Topic{topic, 1, 1, 1},
				Partitions: []int{},
				StartPoint: Beginning,
				Limit:      50,
			})

			var receivedRecords []int
			for {
				select {
				case r, ok := <-rsm.ConsumerRecord:
					if !ok {
						goto assertRecords
					}
					key, _ := strconv.Atoi(r.Key)
					receivedRecords = append(receivedRecords, key)
				}
			}

		assertRecords:
			{
				assert.Len(t, receivedRecords, 50)
				//assert.Equal(t, 49, slices.Max(receivedRecords))
				assert.Equal(t, 0, slices.Min(receivedRecords))
			}

			// clean up
			ka.DeleteTopic(topic)
		})

		t.Run("with multiple partitions", func(t *testing.T) {
			topic := topicName()
			// given
			ka.CreateTopic(TopicCreationDetails{
				Name:          topic,
				NumPartitions: 4,
			})

			// when
			assert.EventuallyWithT(t, func(c *assert.CollectT) {
				for i := 0; i < 55; i++ {
					psm := ka.PublishRecord(&ProducerRecord{
						Topic: topic,
						Key:   strconv.Itoa(i),
						Value: "{\"id\":\"123\"}",
					})

					select {
					case err := <-psm.Err:
						t.Fatal(c, "Unable to publish", err)
					case p := <-psm.Published:
						assert.True(c, p)
					}
				}
			}, 10*time.Second, 10*time.Millisecond)

			// then
			rsm := ka.ReadRecords(context.Background(), ReadDetails{
				Topic:      Topic{topic, 4, 1, 1},
				Partitions: []int{},
				StartPoint: Beginning,
				Limit:      50,
			})

			var receivedRecords []int
			for {
				select {
				case r, ok := <-rsm.ConsumerRecord:
					if !ok {
						goto assertRecords
					}
					key, _ := strconv.Atoi(r.Key)
					receivedRecords = append(receivedRecords, key)
				}
			}

		assertRecords:
			{
				assert.Len(t, receivedRecords, 50)
				assert.Equal(t, 0, slices.Min(receivedRecords))
			}

			// clean up
			ka.DeleteTopic(topic)
		})

	})

	t.Run("Read from MostRecent with specific limit", func(t *testing.T) {
		topic := topicName()
		// given
		ka.CreateTopic(TopicCreationDetails{
			Name:          topic,
			NumPartitions: 1,
		})

		// when
		assert.EventuallyWithT(t, func(c *assert.CollectT) {
			for i := 0; i < 55; i++ {
				psm := ka.PublishRecord(&ProducerRecord{
					Topic: topic,
					Key:   strconv.Itoa(i),
					Value: "{\"id\":\"123\"}",
				})

				select {
				case err := <-psm.Err:
					t.Fatal(c, "Unable to publish", err)
				case p := <-psm.Published:
					assert.True(c, p)
				}
			}
		}, 10*time.Second, 10*time.Millisecond)

		// then
		rsm := ka.ReadRecords(context.Background(), ReadDetails{
			Topic:      Topic{topic, 1, 1, 1},
			Partitions: []int{},
			StartPoint: MostRecent,
			Limit:      50,
		})

		var receivedRecords []int
		for {
			select {
			case r, ok := <-rsm.ConsumerRecord:
				if !ok {
					goto assertRecords
				}
				key, _ := strconv.Atoi(r.Key)
				receivedRecords = append(receivedRecords, key)
			}
		}

	assertRecords:
		{
			assert.Equal(t, 54, slices.Max(receivedRecords))
		}

		// clean up
		ka.DeleteTopic(topic)
	})
}
