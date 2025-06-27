package cfg

import (
	"k8s.io/client-go/rest"
	"os"
)

func InitDefaultConfig(c *rest.Config) *Config {
	m := &MysqlConfig{
		Username: "root",
		Password: "dangerous",
		Host:     "127.0.0.1",
		Port:     3306,
		Database: "thanos",
	}
	cfg := &Config{
		Kubeconfig:           c,
		PromMetricsIntervals: 5,
	}
	cfg.MysqlConfig = m
	ParseConfig(cfg)
	return cfg
}

func ParseConfig(cfg *Config) {
	if val, ok := os.LookupEnv("MYSQL_HOST"); ok {
		cfg.MysqlConfig.Host = val
	}
}
