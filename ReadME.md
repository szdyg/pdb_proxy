# pdb_proxy

pdb代理服务器，用于加速符号（msdl.microsoft.com）下载，同时可以保留一份符号表存在本地，作为节点提供服务。

# 部署

```shell
git clone https://github.com/szdyg/pdb_proxy.git
cd pdb_proxy

# 启动
docker compose up -d

# 停止
docker compose down

# 更新到最新镜像
docker compose pull && docker compose up -d

# 查看日志
docker compose logs -f
```

默认监听 9000 端口，符号缓存在当前目录下的 `./pdb`。

# 配置说明

通过环境变量配置，在 `docker-compose.yml` 的 `environment` 中设置：

| 变量 | 默认值 | 说明 |
| --- | --- | --- |
| `SERVER_PORT` | `0.0.0.0:9000` | 监听地址和端口 |
| `PDB_DIR` | `/pdb` | 缓存pdb的目录（容器内路径，通过 `volumes` 映射到宿主机） |
| `PDB_SERVER` | `https://msdl.microsoft.com/download/symbols` | 远端pdb服务器 |

# 可用节点

http://msdl.szdyg.cn/download/symbols

https://msdl.szdyg.cn/download/symbols

[节点测试下载](http://msdl.szdyg.cn/download/symbols/wrpcrt4.pdb/0DBDD41E0805EAAB4F3FE2365B9EC7A91/wrpcrt4.pdb)

[一键配置工具](https://github.com/szdyg/pdb_config_tool)

# other

如果不需要缓存pdb到本地，推荐直接使用nginx反向代理加速。

参考：https://blog.sunflyer.cn/archives/848
