package cfg

import "k8s.io/client-go/rest"

type Config struct {
	Kubeconfig           *rest.Config
	PromMetricsIntervals int
	*MysqlConfig
}
