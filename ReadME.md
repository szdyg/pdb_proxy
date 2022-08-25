# pdb_proxy

pdb代理服务器，用于加速符号（msdl.microsoft.com）下载，同时可以保留一份符号表存在本地，作为节点提供服务。

# 配置说明
server_port  监听端口

pdb_dir      缓存pdb的目录

pdb_server   远端pdb服务器

# 可用节点

http://msdl.szdyg.icu/download/symbols

https://msdl.szdyg.icu/download/symbols

[节点测试下载](http://msdl.szdyg.icu/download/symbols/wrpcrt4.pdb/0DBDD41E0805EAAB4F3FE2365B9EC7A91/wrpcrt4.pdb)

[一键配置工具](https://github.com/szdyg/pdb_config_tool)

# other

如果不需要缓存pdb到本地，推荐直接使用nginx反向代理加速。

参考：https://blog.sunflyer.cn/archives/848


