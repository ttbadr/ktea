package kcadmin

import (
	"encoding/json"
	tea "github.com/charmbracelet/bubbletea"
	"io"
	"ktea/kadmin"
	"net/http"
)

func (k *DefaultKcAdmin) ListActiveConnectors() tea.Msg {
	req, err := http.NewRequest(http.MethodGet, k.url("/connectors?expand=status"), nil)
	if err != nil {
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
		eChan <- err
		return
	}

	var connectors Connectors
	if err := json.Unmarshal(b, &connectors); err != nil {
		eChan <- err
		return
	}

	cChan <- connectors
}
