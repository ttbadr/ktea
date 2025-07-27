package kadmin

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"ktea/config"
	"ktea/sradmin"
	"os"
	"time"

	"github.com/IBM/sarama"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
)

type SaramaKafkaAdmin struct {
	client   sarama.Client
	admin    sarama.ClusterAdmin
	addrs    []string
	config   *sarama.Config
	producer sarama.SyncProducer
	sra      sradmin.SrAdmin
}

type ConnCheckStartedMsg struct {
	Cluster   *config.Cluster
	Connected chan bool
	Err       chan error
}

func (c *ConnCheckStartedMsg) AwaitCompletion() tea.Msg {
	select {
	case <-c.Connected:
		return ConnCheckSucceededMsg{}
	case err := <-c.Err:
		return ConnCheckErrMsg{Err: err}
	}
}

type ConnCheckSucceededMsg struct{}

type ConnCheckErrMsg struct {
	Err error
}

func ToConnectionDetails(cluster *config.Cluster) ConnectionDetails {
	var saslConfig *SASLConfig
	if cluster.SASLConfig != nil {
		var protocol SASLProtocol
		switch cluster.SASLConfig.SecurityProtocol {
		// SSL, to make wrongly configured PLAINTEXT protocols (as SSL) compatible. Should be removed in the future.
		case config.SASLPlaintextSecurityProtocol, "SSL":
			protocol = PlainText
		default:
			panic(fmt.Sprintf("Unknown SASL protocol: %s", cluster.SASLConfig.SecurityProtocol))
		}

		saslConfig = &SASLConfig{
			Username: cluster.SASLConfig.Username,
			Password: cluster.SASLConfig.Password,
			Protocol: protocol,
		}
	}

	connDetails := ConnectionDetails{
		BootstrapServers:      cluster.BootstrapServers,
		SASLConfig:            saslConfig,
		SSLEnabled:            cluster.SSLEnabled,
		TLSCertFile:           cluster.TLSCertFile,
		TLSKeyFile:            cluster.TLSKeyFile,
		TLSCAFile:             cluster.TLSCAFile,
		TLSInsecureSkipVerify: cluster.TLSInsecureSkipVerify,
	}
	return connDetails
}

func NewSaramaKadmin(cd ConnectionDetails) (Kadmin, error) {
	cfg := sarama.NewConfig()
	cfg.Producer.Return.Successes = true
	cfg.Producer.RequiredAcks = sarama.WaitForAll
	cfg.Producer.Partitioner = sarama.NewRoundRobinPartitioner
	cfg.Consumer.Offsets.Initial = sarama.OffsetOldest

	cfg.Net.TLS.Enable = cd.SSLEnabled

	if cd.SSLEnabled {
		cfg.Net.TLS.Enable = true
		cfg.Net.TLS.Config = &tls.Config{
			InsecureSkipVerify: cd.TLSInsecureSkipVerify,
			ClientAuth:         tls.NoClientCert,
		}

		if cd.TLSCAFile != "" {
			caCert, err := os.ReadFile(cd.TLSCAFile)
			if err != nil {
				return nil, fmt.Errorf("failed to read CA certificate: %w", err)
			}
			caCertPool := x509.NewCertPool()
			caCertPool.AppendCertsFromPEM(caCert)
			cfg.Net.TLS.Config.RootCAs = caCertPool
		}

		if cd.TLSCertFile != "" && cd.TLSKeyFile != "" {
			cert, err := tls.LoadX509KeyPair(cd.TLSCertFile, cd.TLSKeyFile)
			if err != nil {
				return nil, fmt.Errorf("failed to load client key pair: %w", err)
			}
			cfg.Net.TLS.Config.Certificates = []tls.Certificate{cert}
		}
	}

	if cd.SASLConfig != nil {
		cfg.Net.SASL.Enable = true
		cfg.Net.SASL.Mechanism = sarama.SASLTypePlaintext
		cfg.Net.SASL.User = cd.SASLConfig.Username
		cfg.Net.SASL.Password = cd.SASLConfig.Password
	}

	client, err := sarama.NewClient(cd.BootstrapServers, cfg)
	if err != nil {
		return nil, err
	}

	admin, err := sarama.NewClusterAdmin(cd.BootstrapServers, cfg)
	if err != nil {
		return nil, err
	}

	producer, err := sarama.NewSyncProducerFromClient(client)
	if err != nil {
		return nil, err
	}

	return &SaramaKafkaAdmin{
		client:   client,
		admin:    admin,
		addrs:    cd.BootstrapServers,
		producer: producer,
		config:   cfg,
	}, nil
}

func CheckKafkaConnectivity(cluster *config.Cluster) tea.Msg {
	connectedChan := make(chan bool)
	errChan := make(chan error)

	cd := ToConnectionDetails(cluster)
	cfg := sarama.NewConfig()

	cfg.Net.TLS.Enable = cd.SSLEnabled

	if cd.SSLEnabled {
		cfg.Net.TLS.Enable = true
		cfg.Net.TLS.Config = &tls.Config{
			InsecureSkipVerify: cd.TLSInsecureSkipVerify,
			ClientAuth:         tls.NoClientCert,
		}

		if cd.TLSCAFile != "" {
			caCert, err := os.ReadFile(cd.TLSCAFile)
			if err != nil {
				log.Error("failed to read CA certificate for connectivity check", "error", err)
				errChan <- fmt.Errorf("failed to read CA certificate: %w", err)
				return ConnCheckStartedMsg{
					Cluster:   cluster,
					Connected: connectedChan,
					Err:       errChan,
				}
			}
			caCertPool := x509.NewCertPool()
			caCertPool.AppendCertsFromPEM(caCert)
			cfg.Net.TLS.Config.RootCAs = caCertPool
		}

		if cd.TLSCertFile != "" && cd.TLSKeyFile != "" {
			cert, err := tls.LoadX509KeyPair(cd.TLSCertFile, cd.TLSKeyFile)
			if err != nil {
				log.Error("failed to load client key pair for connectivity check", "error", err)
				errChan <- fmt.Errorf("failed to load client key pair: %w", err)
				return ConnCheckStartedMsg{
					Cluster:   cluster,
					Connected: connectedChan,
					Err:       errChan,
				}
			}
			cfg.Net.TLS.Config.Certificates = []tls.Certificate{cert}
		}
	}

	if cd.SASLConfig != nil {
		cfg.Net.SASL.Enable = true
		cfg.Net.SASL.Mechanism = sarama.SASLTypePlaintext
		cfg.Net.SASL.User = cd.SASLConfig.Username
		cfg.Net.SASL.Password = cd.SASLConfig.Password
		cfg.Net.DialTimeout = 5 * time.Second
		cfg.Net.ReadTimeout = 5 * time.Second
		cfg.Net.WriteTimeout = 5 * time.Second
	}

	go doCheckConnectivity(cd, cfg, errChan, connectedChan)

	return ConnCheckStartedMsg{
		Cluster:   cluster,
		Connected: connectedChan,
		Err:       errChan,
	}
}

func doCheckConnectivity(cd ConnectionDetails, config *sarama.Config, errChan chan error, connectedChan chan bool) {
	MaybeIntroduceLatency()
	c, err := sarama.NewClient(cd.BootstrapServers, config)
	if err != nil {
		errChan <- err
		return
	}
	defer func(c sarama.Client) {
		err := c.Close()
		if err != nil {
			log.Error("Unable to close connectivity check connection", err)
		}
	}(c)
	connectedChan <- true
}
