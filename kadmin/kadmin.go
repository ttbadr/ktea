package kadmin

import (
	"time"
)

const (
	PLAIN_TEXT SASLProtocol = 0
	SSL        SASLProtocol = 1
)

const (
	TOPIC_RESOURCE_TYPE = 2
	DEFAULT_TIMEOUT     = 10 * time.Second
)

type ConnectionDetails struct {
	BootstrapServers []string
	SASLConfig       *SASLConfig
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

type Instantiator func(cd ConnectionDetails) (Kadmin, error)

func SaramaInstantiator() Instantiator {
	return func(cd ConnectionDetails) (Kadmin, error) {
		return NewSaramaKadmin(cd)
	}
}
