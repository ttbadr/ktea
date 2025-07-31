package kcadmin

import (
	tea "github.com/charmbracelet/bubbletea"
	"net/http"
)

func (k *DefaultKcAdmin) ListVersion() tea.Msg {
	req, err := k.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		return VersionListingErrMsg{err}
	}

	versionChan := make(chan KafkaConnectVersion)
	errChan := make(chan error)

	go execReq(
		req,
		k.client,
		func(r KafkaConnectVersion) { versionChan <- r },
		func(e error) { errChan <- e },
	)

	return VersionListingStartedMsg{versionChan, errChan}
}
