#!/bin/bash

set -e

# 检查是否为root用户
if [ "$EUID" -ne 0 ]; then
    echo "请使用root权限运行此脚本"
    exit 1
fi

# 检查依赖
if ! command -v jq &> /dev/null; then
    echo "未找到 jq，正在安装..."
    if command -v apt-get &> /dev/null; then
        apt-get update && apt-get install -y jq
    elif command -v yum &> /dev/null; then
        yum install -y jq
    elif command -v dnf &> /dev/null; then
        dnf install -y jq
    else
        echo "无法安装 jq，请手动安装后重试"
        exit 1
    fi
fi

# 获取系统架构
ARCH=$(uname -m)
case $ARCH in
    x86_64)
        ARCH_NAME="amd64"
        ;;
    aarch64)
        ARCH_NAME="arm64"
        ;;
    *)
        echo "不支持的架构: $ARCH"
        exit 1
        ;;
esac

# 检查是否为更新操作
IS_UPDATE=false
if [ -f /usr/bin/pdb-proxy ]; then
    IS_UPDATE=true
    echo "检测到已安装的 PDB Proxy，将进行更新..."
    # 停止服务
    systemctl stop pdb-proxy
fi

# 创建临时目录
TMP_DIR=$(mktemp -d)
cd $TMP_DIR || exit 1

# 获取最新版本信息
echo "正在获取最新版本信息..."
LATEST_RELEASE=$(curl -s https://api.github.com/repos/luodaoyi/pdb_proxy/releases/latest)
if [ $? -ne 0 ]; then
    echo "获取版本信息失败"
    exit 1
fi

# 解析版本号
VERSION=$(echo $LATEST_RELEASE | grep -o '"tag_name": "[^"]*' | cut -d'"' -f4)
echo "最新版本: $VERSION"

# 查找匹配当前架构的资源
echo "正在解析下载链接..."
ASSET_URL=$(echo $LATEST_RELEASE | jq -r ".assets[] | select(.name | contains(\"linux-${ARCH_NAME}\") and contains(\".tar.gz\") and (contains(\".md5\") | not)) | .browser_download_url")

# 打印调试信息
echo "解析到的下载链接: $ASSET_URL"

if [ -z "$ASSET_URL" ]; then
    echo "未找到适配 linux-${ARCH_NAME} 架构的版本"
    echo "完整的release信息："
    echo "$LATEST_RELEASE" | jq '.'
    exit 1
fi

# 下载压缩文件
echo "正在下载 linux-${ARCH_NAME} 版本..."
echo "下载链接: $ASSET_URL"
curl -L -o pdb-proxy.tar.gz "$ASSET_URL"
if [ $? -ne 0 ]; then
    echo "下载失败"
    exit 1
fi

# 解压文件
echo "正在解压文件..."
tar xzf pdb-proxy.tar.gz
if [ ! -f "pdb_proxy" ]; then
    echo "解压后未找到 pdb_proxy 文件"
    ls -la
    exit 1
fi

# 安装二进制文件
chmod +x pdb_proxy
mv pdb_proxy /usr/bin/pdb-proxy

# 创建systemd服务文件
cat > /etc/systemd/system/pdb-proxy.service << EOF
[Unit]
Description=PDB Proxy Service
After=network.target

[Service]
Type=simple
ExecStart=/usr/bin/pdb-proxy
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

# 重新加载 systemd 配置
systemctl daemon-reload

if [ "$IS_UPDATE" = true ]; then
    # 启动服务
    systemctl start pdb-proxy
    echo "更新完成！PDB Proxy 服务已重新启动。"
else
    # 启用并启动服务
    systemctl enable pdb-proxy
    systemctl start pdb-proxy
    echo "安装完成！PDB Proxy 服务已启动并设置为开机自启动。"
fi

# 检查服务状态
systemctl status pdb-proxy

# 清理临时目录
cd /
rm -rf "$TMP_DIR"
