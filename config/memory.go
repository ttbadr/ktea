package config

type InMemoryConfigIO struct {
	config *Config
}

func (i *InMemoryConfigIO) write(config *Config) error {
	i.config = config
	return nil
}

func (i InMemoryConfigIO) read() (*Config, error) {
	if i.config == nil {
		i.config = &Config{}
	}
	return i.config, nil
}
