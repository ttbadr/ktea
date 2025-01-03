package kadmin

type Publisher interface {
	PublishRecord(p *ProducerRecord) PublicationStartedMsg
}

type ProducerRecord struct {
	Key       string
	Value     string
	Topic     string
	Partition *int
}

type PublicationStartedMsg struct {
	Err       chan error
	Published chan bool
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
