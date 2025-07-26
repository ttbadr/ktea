package kcadmin

import (
	"encoding/json"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"io"
	"ktea/kadmin"
	"net/http"
)

func (k *DefaultKcAdmin) ListActiveConnectors() tea.Msg {
	req, err := k.NewRequest(http.MethodGet, "/connectors?expand=status", nil)

	if err != nil {
		log.Error("Error Listing Kafka Connectors", err)
		return ConnectorListingErrMsg{err}
	}

	var (
		cChan = make(chan Connectors)
		eChan = make(chan error)
	)
	go k.doListActiveConnectors(cChan, eChan, req)

	return ConnectorListingStartedMsg{cChan, eChan}
}

func (k *DefaultKcAdmin) doListActiveConnectors(
	cChan chan Connectors,
	eChan chan error,
	req *http.Request,
) {
	kadmin.MaybeIntroduceLatency()
	resp, err := k.client.Do(req)
	if err != nil {
		log.Error("Error Listing Kafka Connectors", err)
		eChan <- err
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			eChan <- err
		}
	}(resp.Body)

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("Error Listing Kafka Connectors", err)
		eChan <- err
		return
	}

	var connectors Connectors
	if err := json.Unmarshal(b, &connectors); err != nil {
		log.Error("Error Listing Kafka Connectors", err)
		eChan <- err
		return
	}

	log.Debug("Listed Kafka Connectors", connectors)

	cChan <- connectors
}
