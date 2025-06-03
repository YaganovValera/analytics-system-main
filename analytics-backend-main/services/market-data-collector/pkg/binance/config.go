package binance

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/YaganovValera/analytics-system/common/backoff"
)

type Config struct {
	URL              string         `mapstructure:"ws_url"`
	Streams          []string       `mapstructure:"streams"`
	BufferSize       int            `mapstructure:"buffer_size"`
	ReadTimeout      time.Duration  `mapstructure:"read_timeout"`
	SubscribeTimeout time.Duration  `mapstructure:"subscribe_timeout"`
	BackoffConfig    backoff.Config `mapstructure:"backoff"`
}

func (c *Config) ApplyDefaults() {
	if c.BufferSize <= 0 {
		c.BufferSize = 10000
	}
	if c.ReadTimeout <= 0 {
		c.ReadTimeout = 30 * time.Second
	}
	if c.SubscribeTimeout <= 0 {
		c.SubscribeTimeout = 5 * time.Second
	}
	for i, s := range c.Streams {
		c.Streams[i] = strings.ToLower(strings.TrimSpace(s))
	}
}

func (c *Config) Validate() error {
	if c.URL == "" {
		return fmt.Errorf("binance: URL is required")
	}
	u, err := url.Parse(c.URL)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return fmt.Errorf("binance: invalid URL: %v", err)
	}
	if len(c.Streams) == 0 {
		return fmt.Errorf("binance: at least one stream is required")
	}
	if c.BufferSize > 50000 {
		return fmt.Errorf("binance: buffer_size too large (>50000)")
	}
	return nil
}
