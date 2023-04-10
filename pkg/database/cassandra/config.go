package cassandra

type Config struct {
	Hosts    []string `yaml:"hosts"`
	Keyspace string   `yaml:"keyspace"`
}
