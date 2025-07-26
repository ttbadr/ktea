package kcadmin

import (
	tea "github.com/charmbracelet/bubbletea"
	"ktea/config"
	"ktea/kadmin"
	"net/http"
)

type ConnCheckStartedMsg struct {
	ConnOk chan bool
	Err    chan error
}

type ConnCheckSucceededMsg struct{}

type ConnCheckErrMsg struct {
	Err error
}

func (c *ConnCheckStartedMsg) AwaitCompletion() tea.Msg {
	select {
	case <-c.ConnOk:
		return ConnCheckSucceededMsg{}
	case err := <-c.Err:
		return ConnCheckErrMsg{err}
	}
}

func (k *DefaultKcAdmin) CheckConnection() tea.Msg {
	req, err := k.NewRequest(http.MethodGet, "/connectors?expand=status", nil)
	if err != nil {
		return ConnectorListingErrMsg{err}
	}

	connOkChan := make(chan bool)
	errChan := make(chan error)

	go k.doCheckConnection(connOkChan, errChan, req)

	return ConnCheckStartedMsg{connOkChan, errChan}
}

func (k *DefaultKcAdmin) doCheckConnection(connOkChan chan bool, errChan chan error, req *http.Request) {
	kadmin.MaybeIntroduceLatency()

	conChan := make(chan Connectors)
	eChan := make(chan error)
	go k.doListActiveConnectors(conChan, eChan, req)

	select {
	case <-conChan:
		connOkChan <- true
	case err := <-eChan:
		errChan <- err
	}
}

// CheckKafkaConnectClustersConn implements ConnChecker.
// CheckKafkaConnectClustersConn checks if the Kafka Connect is reachable and returns a tea.Msg to report status.
func CheckKafkaConnectClustersConn(c *config.KafkaConnectConfig) tea.Msg {
	client := New(http.DefaultClient, c)
	return client.CheckConnection()
}
