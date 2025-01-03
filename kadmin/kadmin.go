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
