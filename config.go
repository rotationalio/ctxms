package ctxms

import (
	"errors"
	"fmt"
	"time"
)

type Config struct {
	Name         string
	Port         uint16
	Delay        time.Duration
	Terminal     bool
	StartingPort uint16
}

func NewConfig() *Config {
	return &Config{
		StartingPort: 9000,
	}
}

func (c *Config) Validate() error {
	if c.Name == "" {
		return errors.New("a unique server name is required")
	}

	if c.Port < c.StartingPort {
		return errors.New("the port cannot be less than the starting port")
	}

	return nil
}

func (c *Config) Addr() string {
	return fmt.Sprintf(":%d", c.Port)
}

func (c *Config) NextHop() string {
	if c.Terminal {
		return fmt.Sprintf(":%d", c.StartingPort)
	}
	return fmt.Sprintf(":%d", c.Port+1)
}
