package config

type MilvusConfig struct {
	Address  string `json:"address" yaml:"address" mapstructure:"address"`    // Remote address, "localhost:19530".
	Username string `json:"username" yaml:"username" mapstructure:"username"` // Username for auth.
	Password string `json:"password" yaml:"password" mapstructure:"password"` // Password for auth.
}
