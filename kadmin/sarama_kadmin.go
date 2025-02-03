package kadmin

import (
	"github.com/IBM/sarama"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"ktea/config"
	"ktea/sradmin"
)

type SaramaKafkaAdmin struct {
	client   sarama.Client
	admin    sarama.ClusterAdmin
	addrs    []string
	config   *sarama.Config
	producer sarama.SyncProducer
	sra      sradmin.SrAdmin
}

type ConnectivityCheckStartedMsg struct {
	Connected chan bool
	Err       chan error
}

func (c *ConnectivityCheckStartedMsg) AwaitCompletion() tea.Msg {
	select {
	case <-c.Connected:
		return ConnectionCheckSucceeded{}
	case err := <-c.Err:
		return ConnectivityCheckErrMsg{Err: err}
	}
}

type ConnectionCheckSucceeded struct{}

type ConnectivityCheckErrMsg struct {
	Err error
}

func ToConnectionDetails(cluster *config.Cluster) ConnectionDetails {
	var saslConfig *SASLConfig
	if cluster.SASLConfig != nil {
		saslConfig = &SASLConfig{
			Username: cluster.SASLConfig.Username,
			Password: cluster.SASLConfig.Password,
			Protocol: SSL,
		}
	}

	connDetails := ConnectionDetails{
		BootstrapServers: cluster.BootstrapServers,
		SASLConfig:       saslConfig,
	}
	return connDetails
}

func NewSaramaKadmin(cd ConnectionDetails) (Kadmin, error) {
	cfg := sarama.NewConfig()
	cfg.Producer.Return.Successes = true
	cfg.Producer.RequiredAcks = sarama.WaitForAll
	cfg.Producer.Partitioner = sarama.NewRoundRobinPartitioner
	cfg.Consumer.Offsets.Initial = sarama.OffsetOldest

	if cd.SASLConfig != nil {
		cfg.Net.TLS.Enable = true
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

func SaramaConnectivityChecker(cluster *config.Cluster) tea.Msg {
	connectedChan := make(chan bool)
	errChan := make(chan error)

	cd := ToConnectionDetails(cluster)
	cfg := sarama.NewConfig()

	if cd.SASLConfig != nil {
		cfg.Net.TLS.Enable = true
		cfg.Net.SASL.Enable = true
		cfg.Net.SASL.Mechanism = sarama.SASLTypePlaintext
		cfg.Net.SASL.User = cd.SASLConfig.Username
		cfg.Net.SASL.Password = cd.SASLConfig.Password
	}

	go doCheckConnectivity(cd, cfg, errChan, connectedChan)

	return ConnectivityCheckStartedMsg{
		Connected: connectedChan,
		Err:       errChan,
	}
}

func doCheckConnectivity(cd ConnectionDetails, config *sarama.Config, errChan chan error, connectedChan chan bool) {
	maybeIntroduceLatency()
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
