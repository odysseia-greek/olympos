package configs

import "github.com/elastic/go-elasticsearch/v7"

type ParmenidesConfig struct {
	Index         string
	Created       int
	ElasticClient elasticsearch.Client
}