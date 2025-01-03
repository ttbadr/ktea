package kadmin

import (
	"context"
	"github.com/IBM/sarama"
)

type StartPoint int

const (
	Beginning  StartPoint = 0
	MostRecent StartPoint = 1
	Today      StartPoint = 2
)

type RecordReader interface {
	ReadRecords(ctx context.Context, rd ReadDetails) ReadingStartedMsg
}

type ReadingStartedMsg struct {
	ConsumerRecord chan ConsumerRecord
	Err            chan error
}

type ReadDetails struct {
	Topic      Topic
	Partitions []int
	StartPoint StartPoint
}

type ConsumerRecord struct {
	Key       string
	Value     string
	Partition int64
	Offset    int64
	Headers   map[string]string
}

func (ka *SaramaKafkaAdmin) ReadRecords(ctx context.Context, rd ReadDetails) ReadingStartedMsg {
	rsm := ReadingStartedMsg{
		ConsumerRecord: make(chan ConsumerRecord),
		Err:            make(chan error),
	}
	var partitions []int
	if len(rd.Partitions) == 0 {
		partitions = make([]int, rd.Topic.Partitions)

		for i := range partitions {
			partitions[i] = i
		}
	} else {
		partitions = rd.Partitions
	}
	client, err := sarama.NewConsumerFromClient(ka.client)
	if err != nil {
		return ReadingStartedMsg{}
	}
	for _, partition := range partitions {
		go func(partition int) {
			consumer, err := client.ConsumePartition(rd.Topic.Name, int32(partition), 0)

			if err != nil {
				rsm.Err <- err
				return
			}

			defer consumer.Close()

			messages := consumer.Messages()

			for {
				select {
				case err = <-consumer.Errors():
					rsm.Err <- err
					return
				case <-ctx.Done():
					return
				case msg := <-messages:
					headers := make(map[string]string)
					for _, h := range msg.Headers {
						headers[string(h.Key)] = string(h.Value)
					}

					rsm.ConsumerRecord <- ConsumerRecord{
						Key:       string(msg.Key),
						Value:     string(msg.Value),
						Partition: int64(msg.Partition),
						Offset:    msg.Offset,
						Headers:   headers,
					}
				}
			}
		}(partition)

	}
	return rsm
}
