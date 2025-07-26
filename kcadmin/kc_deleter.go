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

func (k *DefaultKcAdmin) DeleteConnector(name string) tea.Msg {
	req, err := k.NewRequest(http.MethodDelete, k.url("/connectors/"+name), nil)
	if err != nil {
		log.Error("Error Deleting Kafka Connector", err)
		return ConnectorListingErrMsg{err}
	}

	var (
		dc = make(chan bool)
		ec = make(chan error)
	)

	go k.doDeleteConnector(req, dc, ec)

	return ConnectorDeletionStartedMsg{name, dc, ec}
}

func (k *DefaultKcAdmin) doDeleteConnector(
	req *http.Request,
	dc chan bool,
	ec chan error,
) {
	kadmin.MaybeIntroduceLatency()

	resp, err := k.client.Do(req)
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
		log.Info("Kafka Connector deleted")
		dc <- true
	} else {
		var b bytes.Buffer
		_, err := b.ReadFrom(resp.Body)
		if err != nil {
			log.Error("Error Deleting Kafka Connector", err)
			ec <- fmt.Errorf("unable to delete connector received (%d)", resp.StatusCode)
			return
		}
		log.Error("Error Deleting Kafka Connector", resp.StatusCode)
		ec <- fmt.Errorf("unable to delete connector received (%d): %s", resp.StatusCode, b.String())
	}
}
