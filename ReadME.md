# pdb_proxy

pdb代理服务器，用于加速符号（msdl.microsoft.com）下载，同时可以保留一份符号表存在本地，作为节点提供服务。

# 可用节点
msdl.szdyg.icu

[测试下载](http://msdl.szdyg.icu/download/symbols/wrpcrt4.pdb/0DBDD41E0805EAAB4F3FE2365B9EC7A91/wrpcrt4.pdb)

# other
如果不需要缓存pdb到本地，推荐直接使用nginx反向代理加速。
参考：https://blog.sunflyer.cn/archives/848


