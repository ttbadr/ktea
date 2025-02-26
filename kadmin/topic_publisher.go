package kadmin

import (
	"github.com/IBM/sarama"
	tea "github.com/charmbracelet/bubbletea"
)

type Publisher interface {
	PublishRecord(p *ProducerRecord) PublicationStartedMsg
}

type ProducerRecord struct {
	Key       string
	Value     string
	Topic     string
	Partition *int
	Headers   map[string]string
}

type PublicationStartedMsg struct {
	Err       chan error
	Published chan bool
}

type PublicationFailed struct {
}

type PublicationSucceeded struct {
}

func (p *PublicationStartedMsg) AwaitCompletion() tea.Msg {
	select {
	case <-p.Err:
		return PublicationFailed{}
	case <-p.Published:
		return PublicationSucceeded{}
	}
}

func (ka *SaramaKafkaAdmin) PublishRecord(p *ProducerRecord) PublicationStartedMsg {
	errChan := make(chan error)
	published := make(chan bool)

	go ka.doPublishRecord(p, errChan, published)

	return PublicationStartedMsg{
		Err:       errChan,
		Published: published,
	}
}

func (ka *SaramaKafkaAdmin) doPublishRecord(
	p *ProducerRecord,
	errChan chan error,
	published chan bool,
) {
	MaybeIntroduceLatency()
	var partition int32
	if p.Partition == nil {
		ka.config.Producer.Partitioner = sarama.NewHashPartitioner
	} else {
		partition = int32(*p.Partition)
		ka.config.Producer.Partitioner = sarama.NewManualPartitioner
	}

	var headers []sarama.RecordHeader
	for key, value := range p.Headers {
		headers = append(headers, sarama.RecordHeader{
			Key:   []byte(key),
			Value: []byte(value),
		})
	}

	_, _, err := ka.producer.SendMessage(&sarama.ProducerMessage{
		Topic:     p.Topic,
		Key:       sarama.StringEncoder(p.Key),
		Value:     sarama.StringEncoder(p.Value),
		Partition: partition,
		Headers:   headers,
	})
	if err != nil {
		errChan <- err
	}
	published <- true
}
