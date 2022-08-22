package conf

import "gopkg.in/ini.v1"

var (
	PdbDir     string
	PdbServer  string
	ServerPort string
)

func init() {
	cfg, err := ini.Load("pdb_proxy.ini")
	if err != nil {
		panic("read pdb_proxy.ini error")
	}
	PdbDir = cfg.Section("").Key("pdb_dir").String()
	PdbServer = cfg.Section("").Key("pdb_server").String()
	ServerPort = cfg.Section("").Key("server_port").String()
}
