# pdb_proxy

## linux 自动部署脚本

执行后自动下载配置systemd服务，开机自启动 监听9000端口

```shell 
curl -o- https://raw.githubusercontent.com/luodaoyi/pdb_proxy/master/install.sh | bash

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

# 可用节点

http://msdl.szdyg.cn/download/symbols

https://msdl.szdyg.cn/download/symbols

[节点测试下载](http://msdl.szdyg.cn/download/symbols/wrpcrt4.pdb/0DBDD41E0805EAAB4F3FE2365B9EC7A91/wrpcrt4.pdb)

[一键配置工具](https://github.com/szdyg/pdb_config_tool)

# other

如果不需要缓存pdb到本地，推荐直接使用nginx反向代理加速。

参考：https://blog.sunflyer.cn/archives/848


