package testsetup

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/morehao/go-ark-template/pkg/dbclient"
	"github.com/morehao/golib/dbaccess/dbes"
	"github.com/morehao/golib/dbaccess/dbgorm"
	"github.com/morehao/golib/dbaccess/dbredis"
	"github.com/morehao/golib/glog"
	_ "github.com/morehao/golib/glog/driver/zap"
	"github.com/morehao/golib/gutil"
)

type baseAppInitializer struct {
	AppName    string
	ConfigPath string

	Log         map[string]glog.LogConfig
	DBConfigs   []dbgorm.Config
	RedisConfig dbredis.RedisConfig
	ESConfigs   []dbes.ESConfig
}

type AppConfig struct {
	Log         map[string]glog.LogConfig `yaml:"log"`
	DBConfigs   []dbgorm.Config           `yaml:"db_configs"`
	RedisConfig dbredis.RedisConfig       `yaml:"redis_config"`
	ESConfigs   []dbes.ESConfig           `yaml:"es_configs"`
}

func newBaseAppInitializer(appName string) (*baseAppInitializer, error) {
	configPath := findConfigPath(appName)
	if configPath == "" {
		return nil, fmt.Errorf("cannot find config path for app: %s", appName)
	}

	return &baseAppInitializer{
		AppName:    appName,
		ConfigPath: configPath,
	}, nil
}

func findConfigPath(appName string) string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return ""
	}

	pkgDir := filepath.Dir(filename)
	projectRoot := filepath.Dir(filepath.Dir(pkgDir))

	return filepath.Join(projectRoot, "apps", appName, "config", "config.yaml")
}

func (i *baseAppInitializer) Init() error {
	if _, err := os.Stat(i.ConfigPath); err != nil {
		return fmt.Errorf("config file not found: %s, error: %w", i.ConfigPath, err)
	}

	var cfg AppConfig
	gutil.LoadYamlConfig(i.ConfigPath, &cfg)

	i.Log = cfg.Log
	i.DBConfigs = cfg.DBConfigs
	i.RedisConfig = cfg.RedisConfig
	i.ESConfigs = cfg.ESConfigs

	if err := i.initLog(); err != nil {
		return fmt.Errorf("init log: %w", err)
	}

	if err := i.initResources(); err != nil {
		return fmt.Errorf("init resources: %w", err)
	}

	return nil
}

func (i *baseAppInitializer) initLog() error {
	logCfg, ok := i.Log["default"]
	if !ok {
		for _, c := range i.Log {
			logCfg = c
			break
		}
	}
	return glog.InitLogger(&logCfg)
}

func (i *baseAppInitializer) initResources() error {
	var gormLogConfig *glog.LogConfig
	if c, ok := i.Log["gorm"]; ok {
		gormLogConfig = &c
	}
	if err := dbclient.InitMultiDB(i.DBConfigs, gormLogConfig); err != nil {
		return fmt.Errorf("init db: %w", err)
	}

	if i.RedisConfig.Addr != "" {
		var redisLogConfig *glog.LogConfig
		if c, ok := i.Log["redis"]; ok {
			redisLogConfig = &c
		}
		if err := dbclient.InitRedis(i.RedisConfig, redisLogConfig); err != nil {
			return fmt.Errorf("init redis: %w", err)
		}
	}

	if len(i.ESConfigs) > 0 {
		var esLogConfig *glog.LogConfig
		if c, ok := i.Log["es"]; ok {
			esLogConfig = &c
		}
		if err := dbclient.InitMultiEs(i.ESConfigs, esLogConfig); err != nil {
			return fmt.Errorf("init elasticsearch: %w", err)
		}
	}

	return nil
}

func (i *baseAppInitializer) Close() error {
	if err := glog.Close(); err != nil {
		return fmt.Errorf("close logger: %w", err)
	}
	return nil
}
