package config

import "cluster-scheduler/proto"

type Config struct {
	ListenAddr string
	Algorithm  proto.AlgorithmType
}

func Default() *Config {
	return &Config{
		ListenAddr: ":8080",
		Algorithm:  proto.AlgorithmFIFO,
	}
}
