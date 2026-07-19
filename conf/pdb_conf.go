package conf

import (
	"os"
	"time"
)

var (
	PdbDir      string
	PdbServer   string
	PdbCacheTTL time.Duration
	ServerPort  string
)

func init() {
	PdbDir = GetStrEnvWithDefault("PDB_DIR", defaultPdbDir())
	PdbServer = GetStrEnvWithDefault("PDB_SERVER", "https://msdl.microsoft.com/download/symbols")
	PdbCacheTTL = GetDurationEnvWithDefault("PDB_CACHE_TTL", time.Hour)
	ServerPort = GetStrEnvWithDefault("SERVER_PORT", "0.0.0.0:9000")
}

// GetStrEnvWithDefault 获取字符串环境变量，如果未设置则返回默认值
func GetStrEnvWithDefault(key string, defaultValue string) string {
	val, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return val
}

// GetDurationEnvWithDefault 获取 duration 环境变量。0 表示永久有效；非法值使用默认值。
func GetDurationEnvWithDefault(key string, defaultValue time.Duration) time.Duration {
	val, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}

	duration, err := time.ParseDuration(val)
	if err != nil || duration < 0 {
		return defaultValue
	}
	return duration
}
