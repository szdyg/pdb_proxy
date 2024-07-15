package conf

import (
	"os"
)

var (
	PdbDir     string
	PdbServer  string
	ServerPort string
)

func init() {

	PdbDir = GetStrEnvWithDefault("PDB_DIR", "/pdb")
	PdbServer = GetStrEnvWithDefault("PDB_SERVER", "https://msdl.microsoft.com/download/symbols")
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
