package services

import (
	"strings"
	"time"

	"github.com/spf13/viper"
)

type viperConfigProvider struct {
	prefix string
}

var initialized = false

func NewConfigProvider() ConfigProvider {
	return NewConfigProviderPrefix("")()
}

func NewConfigProviderPrefix(prefix string) func() ConfigProvider {
	if !initialized {
		initialized = true

		viper.AutomaticEnv()
		viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))
		viper.AutomaticEnv()
	}
	return func() ConfigProvider { return &viperConfigProvider{prefix: prefix} }
}

func (c *viperConfigProvider) SetDefault(key string, value any) {
	if c.prefix != "" {
		key = c.prefix + "." + key
	}
	viper.SetDefault(key, value)
}

func (c *viperConfigProvider) Get(key string) any {
	if c.prefix != "" {
		key = c.prefix + "." + key
	}
	return viper.Get(key)
}

func (c *viperConfigProvider) GetBool(key string) bool {
	if c.prefix != "" {
		key = c.prefix + "." + key
	}
	return viper.GetBool(key)
}

func (c *viperConfigProvider) GetInt(key string) int {
	if c.prefix != "" {
		key = c.prefix + "." + key
	}
	return viper.GetInt(key)
}

func (c *viperConfigProvider) GetInt64(key string) int64 {
	if c.prefix != "" {
		key = c.prefix + "." + key
	}
	return viper.GetInt64(key)
}

func (c *viperConfigProvider) GetFloat64(key string) float64 {
	if c.prefix != "" {
		key = c.prefix + "." + key
	}
	return viper.GetFloat64(key)
}

func (c *viperConfigProvider) GetString(key string) string {
	if c.prefix != "" {
		key = c.prefix + "." + key
	}
	return viper.GetString(key)
}

func (c *viperConfigProvider) GetStringSlice(key string) []string {
	if c.prefix != "" {
		key = c.prefix + "." + key
	}
	return viper.GetStringSlice(key)
}

func (c *viperConfigProvider) GetStringMap(key string) map[string]interface{} {
	if c.prefix != "" {
		key = c.prefix + "." + key
	}
	return viper.GetStringMap(key)
}

func (c *viperConfigProvider) GetDuration(key string) time.Duration {
	if c.prefix != "" {
		key = c.prefix + "." + key
	}
	return viper.GetDuration(key)
}
