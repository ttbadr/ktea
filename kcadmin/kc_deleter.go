package kcadmin

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"net/http"
)

func (k *DefaultKcAdmin) DeleteConnector(name string) tea.Msg {
	req, err := k.NewRequest(http.MethodDelete, "/connectors/"+name, nil)
	if err != nil {
		log.Error("Error Deleting Kafka Connector", err)
		return ConnectorDeletionErrMsg{err}
	}

	var (
		dc = make(chan bool)
		ec = make(chan error)
	)

	go execReq(
		req,
		k.client,
		func(_ any) { dc <- true },
		func(e error) { ec <- e },
	)

	return ConnectorDeletionStartedMsg{name, dc, ec}
}
