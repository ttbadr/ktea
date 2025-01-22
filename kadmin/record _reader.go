package kadmin

import (
	"context"
	"github.com/IBM/sarama"
	"github.com/charmbracelet/log"
	"ktea/serdes"
	"sync"
	"time"
)

type StartPoint int

const (
	Beginning  StartPoint = 0
	MostRecent StartPoint = 1
)

type RecordReader interface {
	ReadRecords(ctx context.Context, rd ReadDetails) ReadingStartedMsg
}

type ReadingStartedMsg struct {
	ConsumerRecord chan ConsumerRecord
	Err            chan error
}

type ReadDetails struct {
	Topic      *Topic
	Partitions []int
	StartPoint StartPoint
	Limit      int
}

type Header struct {
	Key   string
	Value string
}

type ConsumerRecord struct {
	Key       string
	Value     string
	Partition int64
	Offset    int64
	Headers   []Header
	Timestamp time.Time
}

func (ka *SaramaKafkaAdmin) ReadRecords(ctx context.Context, rd ReadDetails) ReadingStartedMsg {
	ctx, cancelFunc := context.WithCancel(ctx)
	startedMsg := ReadingStartedMsg{
		ConsumerRecord: make(chan ConsumerRecord, len(rd.Partitions)),
		Err:            make(chan error),
	}

	client, err := sarama.NewConsumerFromClient(ka.client)
	if err != nil {
		close(startedMsg.ConsumerRecord)
		close(startedMsg.Err)
		return startedMsg
	}

	var (
		msgCount   int
		mu         sync.Mutex
		closeOnce  sync.Once
		wg         sync.WaitGroup
		offsets    map[int]int64
		ok         bool
		partitions []int
	)

	partitions = ka.determineReadPartitions(rd)

	if rd.StartPoint != Beginning {
		offsets, ok = ka.fetchFirstAvailableOffsets(partitions, rd, startedMsg)
		if !ok {
			close(startedMsg.Err)
			cancelFunc()
			return startedMsg
		}
	}

	wg.Add(len(partitions))

	for _, partition := range partitions {
		go func(partition int) {
			defer wg.Done()

			startingOffset := ka.determineStartingOffset(partition, rd, offsets)
			consumer, err := client.ConsumePartition(rd.Topic.Name, int32(partition), startingOffset)
			if err != nil {
				log.Error(err)
				startedMsg.Err <- err
				cancelFunc()
				return
			}
			defer consumer.Close()

			msgChan := consumer.Messages()

			for {
				select {
				case err := <-consumer.Errors():
					startedMsg.Err <- err
					return
				case <-ctx.Done():
					return
				case msg := <-msgChan:
					var headers []Header
					for _, h := range msg.Headers {
						headers = append(headers, Header{
							string(h.Key),
							string(h.Value),
						})
					}

					deserializer := serdes.NewAvroDeserializer(ka.sra)
					var payload string
					payload, err = deserializer.Deserialize(msg.Value)
					if err != nil {
						payload = err.Error()
					}

					consumerRecord := ConsumerRecord{
						Key:       string(msg.Key),
						Value:     payload,
						Partition: int64(msg.Partition),
						Offset:    msg.Offset,
						Headers:   headers,
						Timestamp: msg.Timestamp,
					}

					var shouldClose bool

					mu.Lock()
					msgCount++
					if msgCount >= rd.Limit {
						shouldClose = true
					}
					mu.Unlock()

					select {
					case startedMsg.ConsumerRecord <- consumerRecord:
					case <-ctx.Done():
						return
					}

					if shouldClose {
						cancelFunc() // Cancel the context to stop other goroutines
						return
					}
				}
			}
		}(partition)
	}

	go func() {
		wg.Wait()
		closeOnce.Do(func() {
			close(startedMsg.ConsumerRecord)
		})
	}()

	return startedMsg
}

func (ka *SaramaKafkaAdmin) determineReadPartitions(rd ReadDetails) []int {
	var partitions []int
	if len(rd.Partitions) == 0 {
		partitions = make([]int, rd.Topic.Partitions)
		for i := range partitions {
			partitions[i] = i
		}
	} else {
		partitions = rd.Partitions
	}
	return partitions
}

func (ka *SaramaKafkaAdmin) determineStartingOffset(partition int, rd ReadDetails, partByOffset map[int]int64) int64 {
	var startingOffset int64
	if rd.StartPoint == Beginning {
		startingOffset = sarama.OffsetOldest
	} else {
		latestOffset := partByOffset[partition]
		startingOffset = latestOffset - int64(float64(int64(rd.Limit))/float64(len(partByOffset)))
		if startingOffset < 0 {
			startingOffset = 0
		}
	}
	return startingOffset
}

func (ka *SaramaKafkaAdmin) fetchFirstAvailableOffsets(partitions []int, rd ReadDetails, rsm ReadingStartedMsg) (map[int]int64, bool) {
	offsets := make(map[int]int64)
	var wg sync.WaitGroup
	var mu sync.Mutex
	errorsChan := make(chan error, len(partitions))

	for _, partition := range partitions {
		wg.Add(1)
		go func(partition int) {
			defer wg.Done()

			offset, err := ka.client.GetOffset(rd.Topic.Name, int32(partition), sarama.OffsetNewest)
			if err != nil {
				errorsChan <- err
				return
			}

			mu.Lock()
			offsets[partition] = offset
			mu.Unlock()
		}(partition)
	}

	wg.Wait()

	select {
	case err := <-errorsChan:
		rsm.Err <- err
		close(rsm.ConsumerRecord)
		close(rsm.Err)
		return nil, false
	default:
		return offsets, true
	}
}
