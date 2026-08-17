package dbclient

import (
	"fmt"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/morehao/golib/dbaccess/dbes"
	"github.com/morehao/golib/glog"
)

var (
	DemoES *elasticsearch.Client
)

const (
	ESServiceDemo = "demo"
)

func InitMultiEs(configs []dbes.ESConfig, logConfig *glog.LogConfig) error {
	if len(configs) == 0 {
		return fmt.Errorf("es config is empty")
	}
	var opts []dbes.Option
	if logConfig != nil {
		opts = append(opts, dbes.WithLogConfig(logConfig))
	}
	for _, cfg := range configs {
		client, _, err := dbes.New(&cfg, opts...)
		if err != nil {
			return err
		}
		switch cfg.Service {
		case ESServiceDemo:
			DemoES = client
		default:
			return fmt.Errorf("unknown es service name: %s", cfg.Service)
		}
	}
	return nil
}
