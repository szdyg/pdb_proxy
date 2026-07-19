# pdb_proxy

## linux 自动部署脚本

执行后自动下载配置systemd服务，开机自启动 监听9000端口

```shell 
# 安装/更新
curl -o- https://raw.githubusercontent.com/luodaoyi/pdb_proxy/master/install.sh | bash

#卸载
curl -o- https://raw.githubusercontent.com/luodaoyi/pdb_proxy/master/uninstall.sh | bash


#启动： 
systemctl start pdb-proxy
#停止： 
systemctl stop pdb-proxy
#重启： 
systemctl restart pdb-proxy
```

pdb代理服务器，用于加速符号（msdl.microsoft.com）下载，同时可以保留一份符号表存在本地，作为节点提供服务。

# 配置说明
server_port  监听端口

pdb_dir      缓存pdb的目录

pdb_server   远端pdb服务器

pdb_cache_ttl  缓存有效期，对应环境变量 `PDB_CACHE_TTL`，使用 Go duration 格式，例如 `1h`、`30m`；设置为 `0` 表示永久缓存，默认 `1h`

例如 Docker Compose 使用永久缓存：

```shell
PDB_CACHE_TTL=0 docker compose up -d
```

Linux 安装脚本也可以直接指定：

```shell
PDB_CACHE_TTL=24h bash install.sh
```

# 可用节点

http://msdl.szdyg.cn/download/symbols

https://msdl.szdyg.cn/download/symbols

[节点测试下载](http://msdl.szdyg.cn/download/symbols/wrpcrt4.pdb/0DBDD41E0805EAAB4F3FE2365B9EC7A91/wrpcrt4.pdb)

[一键配置工具](https://github.com/szdyg/pdb_config_tool)

# other

如果不需要缓存pdb到本地，推荐直接使用nginx反向代理加速。

参考：https://blog.sunflyer.cn/archives/848


