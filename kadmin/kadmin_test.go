package kadmin

import (
	"context"
	"fmt"
	"github.com/IBM/sarama"
	kgo "github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go/modules/kafka"
	"log"
	"net"
	"strconv"
	"testing"
	"time"
)

var ka *SaramaKafkaAdmin
var ctx context.Context
var kc *kafka.KafkaContainer
var brokers []string

var topicSuffix = 0

func topicName() string {
	topicName := fmt.Sprintf("topic-%d", topicSuffix)
	topicSuffix++
	return topicName
}

func init() {
	ctx = context.Background()
	id := kafka.WithClusterID("")
	k, err := kafka.Run(ctx,
		"confluentinc/confluent-local:7.5.0",
		id,
	)
	if err != nil {
		log.Fatal(fmt.Sprintf("failed to start container: %s", err))
	}
	kc = k
	brokers, _ = kc.Brokers(ctx)
	ka, err = New(ConnectionDetails{
		BootstrapServers: brokers,
		SASLConfig:       nil,
	})
	if err != nil {
		log.Fatal(fmt.Sprintf("failed to create connection: %s", err))
	}
}

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
		assert.Equal(t, []Topic{{topic, 2, 1, 0}}, topics)

		// clean up
		ka.DeleteTopic(topic)
	})
}

func TestListTopics(t *testing.T) {
	t.Run("List all topics", func(t *testing.T) {
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

		assert.EventuallyWithT(t, func(c *assert.CollectT) {
			listTopicsMsg := ka.ListTopics().(TopicListingStartedMsg)
			var topics []Topic
			select {
			case topics = <-listTopicsMsg.Topics:
			case err := <-listTopicsMsg.Err:
				t.Error(t, "Failed to list topics", err)
				return
			}
			assert.ElementsMatch(t, []Topic{{topic1, 2, 1, 0}, {topic2, 1, 1, 0}}, topics)
		}, 2*time.Second, 10*time.Millisecond)

		// clean up
		ka.DeleteTopic(topic1)
		ka.DeleteTopic(topic2)
	})
}

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
				assert.Equal(c, []Topic{{topic1, 2, 1, 0}}, topics)
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
			Topic: Topic{topic, 1, 1, 1},
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
			Topic:      Topic{topic, 2, 1, 1},
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
		assert.Equal(t, "Invalid value -172800000 for configuration delete.retention.ms: Value must be at least 0", msg.(KAdminErrorMsg).Error.Error())

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

type testConsumer struct {
	ExpectedMsgCount int
	msgCount         int
}

func (t *testConsumer) Setup(session sarama.ConsumerGroupSession) error { return nil }

func (t *testConsumer) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (t *testConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message, ok := <-claim.Messages():
			if !ok {
				return nil
			}
			t.msgCount++
			session.MarkMessage(message, "")
			if t.msgCount == t.ExpectedMsgCount-1 {
				return nil
			}
		case <-session.Context().Done():
			return nil
		}
	}
}

func TestConsumerGroups(t *testing.T) {
	t.Run("List groups", func(t *testing.T) {
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

		expectedGroups := make(map[string]bool)
		for i := 0; i < 10; i++ {
			groupName := fmt.Sprintf("test-group-%d", i)
			expectedGroups[groupName] = false
			consumerGroup, err := sarama.NewConsumerGroupFromClient(groupName, ka.client)
			if err != nil {
				t.Fatal("Unable to create Consumer Group.", err)
			}

			handler := testConsumer{ExpectedMsgCount: 10}
			consumerGroup.Consume(context.WithoutCancel(context.Background()), []string{topic}, &handler)

			defer consumerGroup.Close()
		}

		msg := ka.ListConsumerGroups().(ConsumerGroupListingStartedMsg)

		select {
		case groups := <-msg.ConsumerGroups:
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
		consumerGroup, err := sarama.NewConsumerGroupFromClient(groupName, ka.client)
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

func createTopic(t *testing.T, topicConfigs []kgo.TopicConfig) {
	conn, err := kgo.Dial("tcp", brokers[0])
	controller, err := conn.Controller()
	kgo.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	var controllerConn *kgo.Conn
	controllerConn, err = kgo.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	err = controllerConn.CreateTopics(topicConfigs...)
	if err != nil {
		t.Fatal(err.Error())
	}
}
