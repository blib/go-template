package services

import "time"

type ConfigProvider interface {
	SetDefault(key string, value any)
	Get(key string) any
	GetBool(key string) bool
	GetInt(key string) int
	GetInt64(key string) int64
	GetFloat64(key string) float64
	GetString(key string) string
	GetDuration(key string) time.Duration
	GetStringSlice(key string) []string
	GetStringMap(key string) map[string]interface{}
}
