package config

import "cluster-scheduler/proto"

type Config struct {
	ListenAddr    string
	JobAlgorithm  proto.JobAlgorithm
	NodeAlgorithm proto.NodeAlgorithm
}

func Default() *Config {
	return &Config{
		ListenAddr:    ":8080",
		JobAlgorithm:  proto.JobAlgoPriority,
		NodeAlgorithm: proto.NodeAlgoLeastLoaded,
	}
}
