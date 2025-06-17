package kadmin

import (
	"context"
	"github.com/IBM/sarama"
	"testing"
	"time"
)

func TestCGroupDeleter(t *testing.T) {
	t.Run("successful deletion", func(t *testing.T) {
		// create a topic
		topic := topicName()
		msg := ka.CreateTopic(TopicCreationDetails{
			Name:              topic,
			NumPartitions:     1,
			Properties:        nil,
			ReplicationFactor: 1,
		}).(TopicCreationStartedMsg)

		switch msg := msg.AwaitCompletion().(type) {
		case TopicCreatedMsg:
		case TopicCreationErrMsg:
			t.Fatal("Unable to create topic", msg.Err)
		}

		// publish some data on the topic
		for i := 0; i < 10; i++ {
			ka.PublishRecord(&ProducerRecord{
				Key:       "key",
				Value:     []byte("value"),
				Topic:     topic,
				Partition: nil,
			})
		}

		// consume from topic using test-group consumer group
		consumerGroup, err := sarama.NewConsumerGroupFromClient("test-group", kafkaClient())
		if err != nil {
			t.Fatal("Unable to create Consumer Group.", err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // Adjust timeout as needed
		defer cancel()
		handler := testConsumer{ExpectedMsgCount: 10}
		consumerGroup.Consume(ctx, []string{topic}, &handler)
		err = consumerGroup.Close()
		if err != nil {
			t.Fatal("Unable to close group", err)
		}

		// delete the group
		cgroupDeletionStartedMsg := ka.DeleteCGroup("test-group").(CGroupDeletionStartedMsg)

		switch cgroupDeletionStartedMsg.AwaitCompletion().(type) {
		case CGroupDeletionErrMsg:
			t.Fatal("Failed to delete cgroup", msg.Err)
		case CGroupDeletedMsg:
		}
	})
}
