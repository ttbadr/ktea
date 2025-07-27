package kcadmin

import (
	"github.com/stretchr/testify/assert"
	"io"
	"ktea/config"
	"net/http"
	"strings"
	"testing"
)

type mockClient struct {
}

func (m mockClient) Do(*http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: 200,
		Body: io.NopCloser(strings.NewReader(`
 {
        "FileStreamSinkConnectorConnector_0": {
            "status": {
            "name": "FileStreamSinkConnectorConnector_0",
            "connector": {
                "state": "RUNNING",
                "worker_id": "10.0.0.162:8083"
            },
            "tasks": [
                {
                "id": 0,
                "state": "RUNNING",
                "worker_id": "10.0.0.162:8083"
                }
            ],
            "type": "sink"
            }
        },
        "DatagenConnectorConnector_0": {
            "status": {
            "name": "DatagenConnectorConnector_0",
            "connector": {
                "state": "RUNNING",
                "worker_id": "10.0.0.162:8083"
            },
            "tasks": [
                {
                "id": 0,
                "state": "RUNNING",
                "worker_id": "10.0.0.162:8083"
                }
            ],
            "type": "source"
            }
        }
}`)),
	}, nil
}

func TestLister(t *testing.T) {
	t.Run("Returns list of connectors", func(t *testing.T) {
		kcA := New(mockClient{}, &config.KafkaConnectConfig{
			Name: "test",
			Url:  "http://localhost:8083",
		})

		startedMsg := kcA.ListActiveConnectors().(ConnectorListingStartedMsg)

		msg := startedMsg.AwaitCompletion()

		assert.IsType(t, ConnectorsListedMsg{}, msg)

		connectors := msg.(ConnectorsListedMsg).Connectors

		assert.Contains(t, connectors, "FileStreamSinkConnectorConnector_0")
		assert.Contains(t, connectors, "DatagenConnectorConnector_0")

		fs := connectors["FileStreamSinkConnectorConnector_0"]
		assert.Equal(t, "sink", fs.Status.Type)
		assert.Equal(t, "RUNNING", fs.Status.Connector.State)
		assert.Equal(t, "10.0.0.162:8083", fs.Status.Connector.WorkerID)
		assert.Len(t, fs.Status.Tasks, 1)
		assert.Equal(t, 0, fs.Status.Tasks[0].ID)
		assert.Equal(t, "RUNNING", fs.Status.Tasks[0].State)

		dg := connectors["DatagenConnectorConnector_0"]
		assert.Equal(t, "source", dg.Status.Type)
		assert.Equal(t, "RUNNING", dg.Status.Connector.State)
		assert.Equal(t, "10.0.0.162:8083", dg.Status.Connector.WorkerID)
		assert.Len(t, dg.Status.Tasks, 1)
		assert.Equal(t, 0, dg.Status.Tasks[0].ID)
		assert.Equal(t, "RUNNING", dg.Status.Tasks[0].State)
	})
}
