package kadmin

import tea "github.com/charmbracelet/bubbletea"

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
