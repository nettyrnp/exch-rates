package config

import (
	"fmt"
	"log"
	"os/exec"
	"path"
	"reflect"
	"strings"
	"time"

	"github.com/caarlos0/env"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
)

const (
	AppEnvDev = "development"
)

type Config struct {
	AppEnv   string `env:"APP_ENV"`
	Port     string `env:"PORT"`
	Protocol string `env:"PROTOCOL"`

	LogDir      string `env:"LOG_DIR"`
	LogMaxSize  int    `env:"LOG_MAX_SIZE"`
	LogBackups  int    `env:"LOG_BACKUPS"`
	LogMaxAge   int    `env:"LOG_MAX_AGE"`
	LogCompress bool   `env:"LOG_COMPRESS"`

	RepositoryDriver string `env:"CUSTOMER_REPOSITORY_DRIVER"`
	RepositoryDSN    string `env:"CUSTOMER_REPOSITORY_DSN"`

	PollerInterval       time.Duration `env:"POLLER_INTERVAL"`
	PollerBaseCurrencies []string      `env:"POLLER_BASE_CURRENCIES"`
	PollerURL            string        `env:"POLLER_URL"`
	PollerTimeout        time.Duration `env:"POLLER_TIMEOUT"`
}

func Load(filenames ...string) Config {
	if err := godotenv.Load(filenames...); err != nil {
		log.Printf("loading environment file: %s\n", err)
	}

	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("failed to parse environment: %s\n", err)
	}

	// todo: make as absolute paths in '.env'
	rootDir, err := getRootDir()
	if err != nil {
		log.Fatalf(err.Error())
	}
	cfg.LogDir = path.Join(rootDir, cfg.LogDir)

	return cfg
}

func (c Config) Print(fname string) {
	fmt.Println("-------------------------------------------------")
	fmt.Printf("loading environment configuration from %s\n", fname)
	fmt.Println("-------------------------------------------------")

	s := reflect.ValueOf(&c).Elem()
	typeOfT := s.Type()

	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		fmt.Printf("%s=%v\n", typeOfT.Field(i).Name, f.Interface())
	}

	fmt.Println("-------------------------------------------------")
}

func getRootDir() (string, error) {
	out, err := exec.Command("pwd").Output()
	if err != nil {
		return "", errors.New("getting root dir")
	}
	rootDir := string(out)
	suffix := "\n"
	if len(rootDir) == 0 || !strings.HasSuffix(rootDir, suffix) {
		log.Fatalf("failed to get root dir")
	}
	rootDir = rootDir[:len(rootDir)-len(suffix)] // strip suffix
	return rootDir, nil
}
