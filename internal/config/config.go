package config

import (
	"errors"
	"flag"
	"fmt"
	"runtime"
	"time"
)

const (
	minPort          = 1
	maxPort          = 65535
	defaultTimeout   = 200 * time.Millisecond
	defaultHost      = "127.0.0.1"
	defaultLogFormat = false
)

type Config struct {
	Host      string
	StartPort int
	EndPort   int
	Workers   int
	Timeout   time.Duration
	JSONLogs  bool
}

func Parse(args []string) (Config, error) {
	fs := flag.NewFlagSet("portscope", flag.ContinueOnError)

	host := fs.String("host", defaultHost, "target host")
	start := fs.Int("start", minPort, "start port")
	end := fs.Int("end", maxPort, "end port")
	workers := fs.Int("workers", runtime.NumCPU()*256, "number of concurrent workers")
	timeout := fs.Duration("timeout", defaultTimeout, "per-connection timeout")
	jsonLogs := fs.Bool("json", defaultLogFormat, "enable JSON logs")

	if err := fs.Parse(args); err != nil {
		return Config{}, err
	}

	cfg := Config{
		Host:      *host,
		StartPort: *start,
		EndPort:   *end,
		Workers:   *workers,
		Timeout:   *timeout,
		JSONLogs:  *jsonLogs,
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (c Config) Validate() error {
	if c.Host == "" {
		return errors.New("host must not be empty")
	}

	if c.StartPort < minPort || c.StartPort > maxPort {
		return fmt.Errorf("start port must be in range [%d, %d]", minPort, maxPort)
	}

	if c.EndPort < minPort || c.EndPort > maxPort {
		return fmt.Errorf("end port must be in range [%d, %d]", minPort, maxPort)
	}

	if c.StartPort > c.EndPort {
		return errors.New("start port must be less than or equal to end port")
	}

	if c.Workers <= 0 {
		return errors.New("workers must be greater than 0")
	}

	if c.Timeout <= 0 {
		return errors.New("timeout must be greater than 0")
	}

	return nil
}
