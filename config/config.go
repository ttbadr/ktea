package config

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
)

type AuthMethod int

type SecurityProtocol string

const (
	NoneAuthMethod                AuthMethod       = 0
	SASLAuthMethod                AuthMethod       = 1
	SASLPlaintextSecurityProtocol SecurityProtocol = "PLAIN_TEXT"
)

type SASLConfig struct {
	Username         string           `yaml:"username"`
	Password         string           `yaml:"password"`
	SecurityProtocol SecurityProtocol `yaml:"securityProtocol"`
}

type SchemaRegistryConfig struct {
	Url      string `yaml:"url"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type KafkaConnectConfig struct {
	Name     string  `yaml:"name"`
	Url      string  `yaml:"url"`
	Username *string `yaml:"username"`
	Password *string `yaml:"password"`
}

type Cluster struct {
	Name                  string                `yaml:"name"`
	Color                 string                `yaml:"color"`
	Active                bool                  `yaml:"active"`
	BootstrapServers      []string              `yaml:"servers"`
	SASLConfig            *SASLConfig           `yaml:"sasl"`
	SchemaRegistry        *SchemaRegistryConfig `yaml:"schema-registry"`
	SSLEnabled            bool                  `yaml:"ssl-enabled"`
	TLSCertFile           string                `yaml:"tls-cert-file,omitempty"`
	TLSKeyFile            string                `yaml:"tls-key-file,omitempty"`
	TLSCAFile             string                `yaml:"tls-ca-file,omitempty"`
	TLSInsecureSkipVerify bool                  `yaml:"tls-insecure-skip-verify,omitempty"`
	KafkaConnectClusters  []KafkaConnectConfig  `yaml:"kafka-connect-clusters"`
}

func (c *Cluster) HasSchemaRegistry() bool {
	return c.SchemaRegistry != nil
}

func (c *Cluster) HasKafkaConnect() bool {
	return len(c.KafkaConnectClusters) > 0
}

type Config struct {
	Clusters []Cluster `yaml:"clusters"`
	ConfigIO IO        `yaml:"-"`
}

func (c *Config) HasClusters() bool {
	return len(c.Clusters) > 0
}

type SchemaRegistryDetails struct {
	Url      string
	Username string
	Password string
}

type KafkaConnectClusterDetails struct {
	Name     string
	Url      string
	Username *string
	Password *string
}

type RegistrationDetails struct {
	Name                  string
	Color                 string
	Host                  string
	AuthMethod            AuthMethod
	SecurityProtocol      SecurityProtocol
	SSLEnabled            bool
	TLSCertFile           string
	TLSKeyFile            string
	TLSCAFile             string
	TLSInsecureSkipVerify bool
	NewName               *string
	Username              string
	Password              string
	SchemaRegistry        *SchemaRegistryDetails
	KafkaConnectClusters  []KafkaConnectClusterDetails
}

type ClusterDeletedMsg struct {
	Name string
}

type ClusterRegisteredMsg struct {
	Cluster *Cluster
}

type ConnectClusterDeleted struct {
	Name string
}

type ClusterRegisterer interface {
	RegisterCluster(d RegistrationDetails) tea.Msg
}

type ConnectClusterDeleter interface {
	DeleteKafkaConnectCluster(clusterName string, connectName string) tea.Msg
}

func (c *Config) DeleteKafkaConnectCluster(clusterName string, connectName string) tea.Msg {
	for i, cluster := range c.Clusters {
		if clusterName == cluster.Name {
			for _, connectCluster := range cluster.KafkaConnectClusters {
				if connectName == connectCluster.Name {
					c.Clusters[i].KafkaConnectClusters = deleteKafkaConnectCluster(c.Clusters[i].KafkaConnectClusters, connectName)
					c.flush()
					return ConnectClusterDeleted{connectName}
				}
			}
		}
	}
	return nil
}

func deleteKafkaConnectCluster(clusters []KafkaConnectConfig, name string) []KafkaConnectConfig {
	out := make([]KafkaConnectConfig, 0, len(clusters)-1)
	for _, c := range clusters {
		if c.Name != name {
			out = append(out, c)
		}
	}
	return out
}

// RegisterCluster registers a new cluster or updates an existing one in the Config.
//
// If a cluster with the same name exists, it updates its details while retaining the "Active" status (the active param
// in that case is ignored) and optionally renaming it. Otherwise, it adds the cluster to the Config.
//
// It returns a ClusterRegisteredMsg with the registered cluster.
func (c *Config) RegisterCluster(details RegistrationDetails) tea.Msg {
	cluster := ToCluster(details)

	// When no clusters exist yet, the first one created becomes the active one by default.
	if len(c.Clusters) == 0 {
		cluster.Active = true
	}

	// did the newly registered cluster update an existing one
	var isUpdated bool

	for i := range c.Clusters {
		if c.Clusters[i].Name == details.Name {
			isActive := c.Clusters[i].Active
			cluster.Active = isActive
			c.Clusters[i] = cluster
			if details.NewName != nil {
				c.Clusters[i].Name = *details.NewName
			}
			isUpdated = true
			break
		}
	}

	// not updated, add new cluster
	if !isUpdated {
		c.Clusters = append(c.Clusters, cluster)
	}

	c.flush()

	return ClusterRegisteredMsg{&cluster}
}

func ToCluster(details RegistrationDetails) Cluster {
	cluster := Cluster{
		Name:                  details.Name,
		Color:                 details.Color,
		BootstrapServers:      []string{details.Host},
		SSLEnabled:            details.SSLEnabled,
		TLSCertFile:           details.TLSCertFile,
		TLSKeyFile:            details.TLSKeyFile,
		TLSCAFile:             details.TLSCAFile,
		TLSInsecureSkipVerify: details.TLSInsecureSkipVerify,
	}

	if details.AuthMethod == SASLAuthMethod {
		cluster.SASLConfig = &SASLConfig{
			Username:         details.Username,
			Password:         details.Password,
			SecurityProtocol: details.SecurityProtocol,
		}
	}

	if details.SchemaRegistry != nil {
		cluster.SchemaRegistry = &SchemaRegistryConfig{
			Url:      details.SchemaRegistry.Url,
			Username: details.SchemaRegistry.Username,
			Password: details.SchemaRegistry.Password,
		}
	}

	if details.KafkaConnectClusters != nil {
		for _, connectCluster := range details.KafkaConnectClusters {
			cluster.KafkaConnectClusters = append(cluster.KafkaConnectClusters, KafkaConnectConfig{
				Name:     connectCluster.Name,
				Url:      connectCluster.Url,
				Username: connectCluster.Username,
				Password: connectCluster.Password,
			})
		}
	}

	return cluster
}

func (c *Config) ActiveCluster() *Cluster {
	for _, c := range c.Clusters {
		if c.Active {
			return &c
		}
	}
	if len(c.Clusters) > 0 {
		return &c.Clusters[0]
	}
	return nil
}

func (c *Config) flush() {
	if err := c.ConfigIO.write(c); err != nil {
		fmt.Println("Unable to write config file")
		os.Exit(-1)
	}
	log.Debug("flushed config")
}

func (c *Config) SwitchCluster(name string) *Cluster {
	var activeCluster *Cluster

	for i := range c.Clusters {
		// deactivate all clusters
		c.Clusters[i].Active = false
		if c.Clusters[i].Name == name {
			c.Clusters[i].Active = true
			activeCluster = &c.Clusters[i]
			log.Debug("switched to cluster " + name)
		}
	}

	c.flush()

	return activeCluster
}

func (c *Config) DeleteCluster(name string) {
	// Find the index of the cluster to delete
	index := -1
	for i, cluster := range c.Clusters {
		if cluster.Name == name {
			index = i
			break
		}
	}

	if index == -1 {
		log.Warn("cluster not found: " + name)
		return
	}

	// Check if the cluster to delete is active
	isActive := c.Clusters[index].Active

	// Remove the cluster
	c.Clusters = append(c.Clusters[:index], c.Clusters[index+1:]...)

	// Reactivate the first cluster if needed
	if isActive && len(c.Clusters) > 0 {
		c.Clusters[0].Active = true
	}

	c.flush()

	log.Debug("deleted cluster: " + name)
}

func (c *Config) FindClusterByName(name string) *Cluster {
	for _, cluster := range c.Clusters {
		if cluster.Name == name {
			return &cluster
		}
	}
	return nil
}

type LoadedMsg struct {
	Config *Config
}

func New(configIO IO) *Config {
	config, err := configIO.read()
	if err != nil {
		fmt.Println("Error reading config file:", err)
		os.Exit(-1)
	}
	config.ConfigIO = configIO
	return config
}
