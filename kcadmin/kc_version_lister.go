package kcadmin

import (
	tea "github.com/charmbracelet/bubbletea"
	"net/http"
)

func (k *DefaultKcAdmin) ListVersion() tea.Msg {
	req, err := k.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		return ConnectorListingErrMsg{err}
	}

	versionChan := make(chan KafkaConnectVersion)
	errChan := make(chan error)

	go execReq(req, k.client, versionChan, errChan)

	return VersionListingStartedMsg{versionChan, errChan}
}
