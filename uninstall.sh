#!/bin/bash

set -e

# 检查是否为root用户
if [ "$EUID" -ne 0 ]; then
    echo "请使用root权限运行此脚本"
    exit 1
fi

# 检查是否已安装 PDB Proxy
if [ ! -f /usr/bin/pdb-proxy ]; then
    echo "未检测到已安装的 PDB Proxy，退出..."
    exit 0
fi

echo "正在卸载 PDB Proxy..."
# 停止服务
systemctl stop pdb-proxy
# 禁用服务
systemctl disable pdb-proxy
# 删除服务文件
rm -f /etc/systemd/system/pdb-proxy.service
# 重新加载 systemd 配置
systemctl daemon-reload
# 删除二进制文件
rm -f /usr/bin/pdb-proxy
echo "卸载完成！PDB Proxy 已被移除。"
exit 0
