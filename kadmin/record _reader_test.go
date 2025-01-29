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
			msg := ka.CreateTopic(TopicCreationDetails{
				Name:          topic,
				NumPartitions: 1,
			}).(TopicCreationStartedMsg)

			switch msg.AwaitCompletion().(type) {
			case TopicCreatedMsg:
			case TopicCreationErrMsg:
				t.Fatal("Unable to create topic", msg.Err)
			}
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
				Topic:      &Topic{topic, 1, 1, 1},
				Partitions: []int{},
				StartPoint: Beginning,
				Limit:      50,
			}).(ReadingStartedMsg)

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
			msg := ka.CreateTopic(TopicCreationDetails{
				Name:          topic,
				NumPartitions: 4,
			}).(TopicCreationStartedMsg)

			switch msg.AwaitCompletion().(type) {
			case TopicCreatedMsg:
			case TopicCreationErrMsg:
				t.Fatal("Unable to create topic", msg.Err)
			}

			// when
			assert.EventuallyWithT(t, func(c *assert.CollectT) {
				for i := 0; i < 52; i++ {
					partition := i % 4
					psm := ka.PublishRecord(&ProducerRecord{
						Topic:     topic,
						Key:       strconv.Itoa(i),
						Partition: &partition,
						Value:     "{\"id\":\"123\"}",
					})

					select {
					case err := <-psm.Err:
						t.Fatal("Unable to publish", err)
					case p := <-psm.Published:
						assert.True(c, p)
					}
				}
			}, 10*time.Second, 10*time.Millisecond)

			// then
			rsm := ka.ReadRecords(context.Background(), ReadDetails{
				Topic:      &Topic{topic, 4, 1, 1},
				Partitions: []int{},
				StartPoint: Beginning,
				Limit:      40,
			}).(ReadingStartedMsg)

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
				assert.Len(t, receivedRecords, 40)
				assert.Equal(t, 0, slices.Min(receivedRecords))
			}

			// clean up
			ka.DeleteTopic(topic)
		})
	})

	t.Run("Read from MostRecent", func(t *testing.T) {
		t.Run("with specific in range limit", func(t *testing.T) {
			topic := topicName()
			// given
			msg := ka.CreateTopic(TopicCreationDetails{
				Name:          topic,
				NumPartitions: 1,
			}).(TopicCreationStartedMsg)

			switch msg.AwaitCompletion().(type) {
			case TopicCreatedMsg:
			case TopicCreationErrMsg:
				t.Fatal("Unable to create topic", msg.Err)
			}

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
				Topic:      &Topic{topic, 1, 1, 1},
				Partitions: []int{},
				StartPoint: MostRecent,
				Limit:      50,
			}).(ReadingStartedMsg)

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
				assert.Equal(t, 53, slices.Max(receivedRecords))
			}

			// clean up
			ka.DeleteTopic(topic)
		})

		t.Run("with specific in out of range limit", func(t *testing.T) {
			topic := topicName()
			// given
			msg := ka.CreateTopic(TopicCreationDetails{
				Name:          topic,
				NumPartitions: 1,
			}).(TopicCreationStartedMsg)

			switch msg.AwaitCompletion().(type) {
			case TopicCreatedMsg:
			case TopicCreationErrMsg:
				t.Fatal("Unable to create topic", msg.Err)
			}

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
				Topic:      &Topic{topic, 1, 1, 1},
				Partitions: []int{},
				StartPoint: MostRecent,
				Limit:      500,
			}).(ReadingStartedMsg)

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

	})
}

type want struct {
	start int64
	end   int64
}

type determineStartingOffsetTest struct {
	name        string
	want        want
	readDetails ReadDetails
	offsets     offsets
}

func TestDetermineStartingOffset(t *testing.T) {
	var tests = []determineStartingOffsetTest{
		{
			name: "beginning one partition enough records available",
			readDetails: ReadDetails{
				Topic: &Topic{
					Name:       "test-topic",
					Partitions: 1,
					Replicas:   1,
					Isr:        1,
				},
				Partitions: []int{0},
				StartPoint: Beginning,
				Limit:      50,
			},
			offsets: offsets{
				oldest:         1,
				firstAvailable: 291,
			},

			want: want{
				start: 1,
				end:   50,
			},
		},
		{
			name: "beginning multiple partition enough records available",
			readDetails: ReadDetails{
				Topic: &Topic{
					Name:       "test-topic",
					Partitions: 5,
					Replicas:   1,
					Isr:        1,
				},
				Partitions: []int{0, 1, 2, 3, 4},
				StartPoint: Beginning,
				Limit:      50,
			},
			offsets: offsets{
				oldest:         1,
				firstAvailable: 291,
			},

			want: want{
				start: 1,
				end:   10,
			},
		},
		{
			name: "beginning one partition not enough records available",
			readDetails: ReadDetails{
				Topic: &Topic{
					Name:       "test-topic",
					Partitions: 1,
					Replicas:   1,
					Isr:        1,
				},
				Partitions: []int{0},
				StartPoint: Beginning,
				Limit:      50,
			},
			offsets: offsets{
				oldest:         55,
				firstAvailable: 76,
			},
			want: want{
				start: 55,
				end:   75,
			},
		},
		{
			name: "beginning multiple partition not enough records available",
			readDetails: ReadDetails{
				Topic: &Topic{
					Name:       "test-topic",
					Partitions: 4,
					Replicas:   1,
					Isr:        1,
				},
				Partitions: []int{0, 2, 3, 4},
				StartPoint: Beginning,
				Limit:      50,
			},
			offsets: offsets{
				oldest:         10,
				firstAvailable: 12,
			},

			want: want{
				start: 10,
				end:   11,
			},
		},
		{
			name: "most recent one partition enough records available",
			readDetails: ReadDetails{
				Topic: &Topic{
					Name:       "test-topic",
					Partitions: 1,
					Replicas:   1,
					Isr:        1,
				},
				Partitions: []int{0},
				StartPoint: MostRecent,
				Limit:      50,
			},
			offsets: offsets{
				oldest:         1,
				firstAvailable: 291,
			},

			want: want{
				start: 240,
				end:   290,
			},
		},
		{
			name: "most recent one partition not enough records available",
			readDetails: ReadDetails{
				Topic: &Topic{
					Name:       "test-topic",
					Partitions: 1,
					Replicas:   1,
					Isr:        1,
				},
				Partitions: []int{0},
				StartPoint: MostRecent,
				Limit:      50,
			},
			offsets: offsets{
				oldest:         278,
				firstAvailable: 291,
			},

			want: want{
				start: 278,
				end:   290,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			offset := ka.(*SaramaKafkaAdmin).determineReadingOffsets(
				test.readDetails,
				test.offsets,
			)
			assert.Equal(t, test.want.start, offset.start, "unexpected start")
			assert.Equal(t, test.want.end, offset.end, "unexpected end")
		})
	}
}
