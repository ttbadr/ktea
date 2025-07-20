package kcadmin

import (
	tea "github.com/charmbracelet/bubbletea"
	"ktea/config"
	"net/http"
)

type KcAdmin interface {
	ConnectorLister
	ConnectorDeleter
}

// ConnectorLister defines the behavior of listing active Kafka connectors.
// returns a tea.Msg that can be either a ConnectorListingStartedMsg or a ConnectorListingErrMsg
type ConnectorLister interface {
	ListActiveConnectors() tea.Msg
}

type ConnectorDeleter interface {
	DeleteConnector(name string) tea.Msg
}

// ConnChecker is a function that checks a Kafka Connect Cluster connection and returns a tea.Msg.
type ConnChecker func(c *config.KafkaConnectConfig) tea.Msg

type ConnectorStatus struct {
	Name      string         `json:"name"`
	Connector ConnectorState `json:"connector"`
	Tasks     []TaskState    `json:"tasks"`
	Type      string         `json:"type"`
}

type ConnectorState struct {
	State    string `json:"state"`
	WorkerID string `json:"worker_id"`
}

type TaskState struct {
	ID       int    `json:"id"`
	State    string `json:"state"`
	WorkerID string `json:"worker_id"`
}

type Connectors map[string]struct {
	Status ConnectorStatus `json:"status"`
}

type Client interface {
	Do(req *http.Request) (*http.Response, error)
}

type DefaultKcAdmin struct {
	client  Client
	baseUrl string
}

type ConnectorListingStartedMsg struct {
	Connectors chan Connectors
	Err        chan error
}

type ConnectorsListedMsg struct {
	Connectors
}

func (c *ConnectorListingStartedMsg) AwaitCompletion() tea.Msg {
	select {
	case con := <-c.Connectors:
		return ConnectorsListedMsg{con}
	case err := <-c.Err:
		return ConnectorListingErrMsg{err}
	}
}

type ConnectorListingErrMsg struct {
	Err error
}

type ConnectorDeletionStartedMsg struct {
	Name    string
	Deleted chan bool
	Err     chan error
}

func (m *ConnectorDeletionStartedMsg) AwaitCompletion() tea.Msg {
	select {
	case <-m.Deleted:
		return ConnectorDeletedMsg{m.Name}
	case err := <-m.Err:
		return ConnectorDeletionErrMsg{err}
	}
}

type ConnectorDeletedMsg struct {
	Name string
}

type ConnectorDeletionErrMsg struct {
	Err error
}

func (k *DefaultKcAdmin) url(path string) string {
	return k.baseUrl + path
}

func New(c Client, baseUrl string) *DefaultKcAdmin {
	return &DefaultKcAdmin{client: c, baseUrl: baseUrl}
}
