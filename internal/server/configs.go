// Main package with  metric server app configuration
package server

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/alexflint/go-arg"
	"github.com/screamsoul/go-metrics-tpl/pkg/ipmask"
	"github.com/screamsoul/go-metrics-tpl/pkg/utils"
)

type Postgres struct {
	DatabaseDSN      string          `arg:"-d,env:DATABASE_DSN" default:"" help:"Строка подключения к базе Postgres" json:"database_dsn"`
	BackoffIntervals []time.Duration `arg:"--b-intervals,env:BACKOFF_INTERVALS" help:"Интервалы повтора запроса (обязательно если (default=1s,3s,5s)"`
	BackoffRetries   bool            `arg:"--backoff,env:BACKOFF_RETRIES" default:"true" help:"Повтор запроса при разрыве соединения"`
}

type CryptoPublicKey struct {
	Key *rsa.PrivateKey
}

type Config struct {
	Postgres
	ListenAddress     string          `arg:"-a,env:ADDRESS" default:"localhost:8080" help:"Адрес и порт сервера" json:"address"`
	LogLevel          string          `arg:"--ll,env:LOG_LEVEL" default:"INFO" help:"Уровень логирования"`
	StoreInterval     int             `arg:"-i,env:STORE_INTERVAL" default:"300" help:"Интервал времени в секундах, по истечении которого текущие показания сервера сохраняются на диск" json:"store_interval"`
	FileStoragePath   string          `arg:"-f,env:FILE_STORAGE_PATH" default:"/tmp/metrics-db.json" help:"Полное имя файла, куда сохраняются текущие значения" json:"store_file"`
	Restore           bool            `arg:"-r,env:RESTORE" default:"true" help:"Загружать или нет ранее сохранённые значения из указанного файла при старте сервера" json:"restore"`
	HashBodyKey       string          `arg:"-k,env:KEY" default:"" help:"hash key"`
	Debug             bool            `arg:"--debug,env:DEBUG" default:"false" help:"debug mode"`
	CryptoKey         CryptoPublicKey `arg:"--crypto-key,env:CRYPTO_KEY" default:"" help:"the path to the file with the public key" json:"crypto_key"`
	TrustedSubnetCIDR ipmask.CIDRIP   `arg:"-t,env:TRUSTED_SUBNET" default:"" help:"allowed subnet in the classless addressing string format (CIDR)" json:"trusted_subnet"`
}

func (cpk *CryptoPublicKey) UnmarshalText(b []byte) error {
	keyData, err := os.ReadFile(string(b))
	if err != nil {
		return err
	}

	block, _ := pem.Decode(keyData)
	if block == nil || !strings.Contains(block.Type, "PRIVATE KEY") {
		return fmt.Errorf("not find private key in file")
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return err
	}
	rsaPrivateKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return fmt.Errorf("not an RSA private key")
	}

	cpk.Key = rsaPrivateKey
	return nil
}

func NewConfig() (*Config, error) {
	var cfg Config

	err := utils.FillFromFile(&cfg)
	if err != nil {
		return nil, err
	}

	if err := arg.Parse(&cfg); err != nil {
		return nil, err
	}

	if cfg.Postgres.BackoffIntervals == nil && cfg.Postgres.BackoffRetries {
		cfg.Postgres.BackoffIntervals = []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}
	} else if !cfg.Postgres.BackoffRetries {
		cfg.Postgres.BackoffIntervals = nil
	}

	return &cfg, nil
}
