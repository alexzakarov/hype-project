package eventstroredb

import (
	"github.com/kurrent-io/KurrentDB-Client-Go/kurrentdb"
)

func NewEventStoreDB(cfg EventStoreConfig) (*kurrentdb.Client, error) {
	settings, err := kurrentdb.ParseConnectionString(cfg.ConnectionString)
	if err != nil {
		return nil, err
	}

	return kurrentdb.NewClient(settings)
}
