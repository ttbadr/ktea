package kcadmin

import (
	"bytes"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"io"
	"ktea/kadmin"
	"net/http"
)

func (h *DefaultKcAdmin) DeleteConnector(name string) tea.Msg {
	req, err := http.NewRequest(http.MethodDelete, h.url("/connectors/"+name), nil)
	if err != nil {
		return ConnectorListingErrMsg{err}
	}

	var (
		dc = make(chan bool)
		ec = make(chan error)
	)

	go h.doDeleteConnector(req, dc, ec)

	return ConnectorDeletionStartedMsg{name, dc, ec}
}

func (h *DefaultKcAdmin) doDeleteConnector(
	req *http.Request,
	dc chan bool,
	ec chan error,
) {
	kadmin.MaybeIntroduceLatency()

	resp, err := h.client.Do(req)
	if err != nil {
		log.Error("Failed to delete connector", err)
		ec <- err
		return
	}
	defer func(b io.ReadCloser) {
		err := b.Close()
		if err != nil {
			ec <- err
		}
	}(resp.Body)

	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		log.Info("Connector deleted")
		dc <- true
	} else {
		var b bytes.Buffer
		_, err := b.ReadFrom(resp.Body)
		if err != nil {
			log.Info("Connector deletion error", err)
			ec <- fmt.Errorf("unable to delete connector received (%d)", resp.StatusCode)
			return
		}
		log.Info("Connector deletion error", resp.StatusCode)
		ec <- fmt.Errorf("unable to delete connector received (%d): %s", resp.StatusCode, b.String())
	}
}
