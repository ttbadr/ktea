package kadmin

import (
	tea "github.com/charmbracelet/bubbletea"
	"ktea/config"
)

const (
	PlainText SASLProtocol = 0
)

const (
	TopicResourceType = 2
)

type Kadmin interface {
	TopicCreator
	TopicDeleter
	TopicLister
	Publisher
	RecordReader
	OffsetLister
	CGroupLister
	CGroupDeleter
	ConfigUpdater
	TopicConfigLister
	SraSetter
}

type ConnectionDetails struct {
	BootstrapServers []string
	SASLConfig       *SASLConfig
	SSLEnabled       bool
	TLSCertFile      string
	TLSKeyFile       string
	TLSCAFile        string
	TLSInsecureSkipVerify bool
}

type SASLProtocol int

type SASLConfig struct {
	Username string
	Password string
	Protocol SASLProtocol
}

type GroupMember struct {
	MemberId   string
	ClientId   string
	ClientHost string
}

type KAdminErrorMsg struct {
	Error error
}

type ConnErrMsg struct {
	Err error
}

type Instantiator func(cd ConnectionDetails) (Kadmin, error)

type ConnChecker func(cluster *config.Cluster) tea.Msg

func SaramaInstantiator() Instantiator {
	return func(cd ConnectionDetails) (Kadmin, error) {
		return NewSaramaKadmin(cd)
	}
}
