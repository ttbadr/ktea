package kcadmin

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
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

	go execReq(req, k.client, cChan, eChan)

	return ConnectorListingStartedMsg{cChan, eChan}
}
