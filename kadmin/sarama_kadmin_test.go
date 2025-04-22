package kadmin

import (
	"github.com/stretchr/testify/assert"
	"ktea/config"
	"testing"
)

func TestToConnectionDetails(t *testing.T) {
	type args struct {
		cluster *config.Cluster
	}
	tests := []struct {
		name string
		args args
		want ConnectionDetails
	}{
		{
			name: "map properties",
			args: args{
				cluster: &config.Cluster{
					Name:             "DEV",
					Color:            "#red",
					Active:           false,
					BootstrapServers: []string{"localhost:9092"},
					SASLConfig: &config.SASLConfig{
						Username:         "Fred",
						Password:         "Wrong",
						SecurityProtocol: config.SASLPlaintextSecurityProtocol,
					},
					SSLEnabled:     true,
					SchemaRegistry: nil,
				},
			},
			want: ConnectionDetails{
				BootstrapServers: []string{"localhost:9092"},
				SASLConfig: &SASLConfig{
					Username: "Fred",
					Password: "Wrong",
					Protocol: PLAIN_TEXT,
				},
				SSLEnabled: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, ToConnectionDetails(tt.args.cluster), "ToConnectionDetails(%v)", tt.args.cluster)
		})
	}
}
