package elastic

import (
	"context"
	"github.com/olivere/elastic/v7"
	"github.com/pkg/errors"
	"main/config"
	"main/pkg/elasticsearch"
	"main/pkg/logger"
)

func InitElasticClient(ctx context.Context, cfg *config.Config, logger logger.Logger) (*elastic.Client, error) {
	elasticClient, err := elasticsearch.NewElasticClient(cfg.Elastic)
	if err != nil {
		return nil, err
	}

	info, code, err := elasticClient.Ping(cfg.Elastic.URL).Do(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "client.Ping")
	}
	logger.Infof("Elasticsearch returned with code {%d} and version {%s}", code, info.Version.Number)

	esVersion, err := elasticClient.ElasticsearchVersion(cfg.Elastic.URL)
	if err != nil {
		return nil, errors.Wrap(err, "client.ElasticsearchVersion")
	}
	logger.Infof("Elasticsearch version {%s}", esVersion)

	return elasticClient, nil
}
