#!/bin/bash

set -euo pipefail

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

# 检查基础工具
if ! command -v curl &> /dev/null; then
    echo "未找到 curl，请先安装后重试"
    exit 1
fi
if ! command -v tar &> /dev/null; then
    echo "未找到 tar，请先安装后重试"
    exit 1
fi
if ! command -v systemctl &> /dev/null; then
    echo "未找到 systemctl，无法安装 systemd 服务"
    exit 1
fi

# 交互设置端口
DEFAULT_PORT="9000"
INPUT_PORT=""
if ! read -r -p "请输入监听端口(默认${DEFAULT_PORT}): " INPUT_PORT; then
    INPUT_PORT=""
fi
INPUT_PORT=$(echo "$INPUT_PORT" | tr -d '[:space:]')
if [ -z "$INPUT_PORT" ]; then
    PORT="$DEFAULT_PORT"
elif [[ "$INPUT_PORT" =~ ^[0-9]+$ ]] && [ "$INPUT_PORT" -ge 1 ] && [ "$INPUT_PORT" -le 65535 ]; then
    PORT="$INPUT_PORT"
else
    echo "端口不合法，使用默认端口 ${DEFAULT_PORT}"
    PORT="$DEFAULT_PORT"
fi

# 默认配置
PDB_DIR="/opt/pdb"
PDB_SERVER="https://msdl.microsoft.com/download/symbols"
PDB_CACHE_TTL="${PDB_CACHE_TTL:-1h}"
SERVER_PORT="0.0.0.0:${PORT}"

# 确保缓存目录存在
mkdir -p "$PDB_DIR"

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
    if systemctl is-active --quiet pdb-proxy; then
        systemctl stop pdb-proxy
    fi
fi

# 创建临时目录
TMP_DIR=$(mktemp -d -t pdb-proxy.XXXXXX)
cleanup_tmp_dir() {
    if [ -n "${TMP_DIR:-}" ] && [ -d "$TMP_DIR" ]; then
        case "$TMP_DIR" in
            /tmp/pdb-proxy.*|/var/tmp/pdb-proxy.*)
                rm -rf -- "$TMP_DIR"
                ;;
            *)
                echo "警告：临时目录路径异常，跳过清理：$TMP_DIR"
                ;;
        esac
    fi
}
trap cleanup_tmp_dir EXIT
cd "$TMP_DIR" || exit 1

# 获取最新版本信息
echo "正在获取最新版本信息..."
LATEST_RELEASE=$(curl -fsSL https://api.github.com/repos/luodaoyi/pdb_proxy/releases/latest)

# 解析版本号
VERSION=$(echo "$LATEST_RELEASE" | jq -r '.tag_name // empty')
if [ -z "$VERSION" ]; then
    echo "解析版本号失败，请稍后重试或检查网络"
    exit 1
fi
echo "最新版本: $VERSION"

# 查找匹配当前架构的资源
echo "正在解析下载链接..."
ASSET_URL=$(echo "$LATEST_RELEASE" | jq -r ".assets[] | select(.name | contains(\"linux-${ARCH_NAME}\") and contains(\".tar.gz\") and (contains(\".md5\") | not)) | .browser_download_url")

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
curl -fL -o pdb-proxy.tar.gz "$ASSET_URL"

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
Environment="PDB_DIR=${PDB_DIR}"
Environment="PDB_SERVER=${PDB_SERVER}"
Environment="PDB_CACHE_TTL=${PDB_CACHE_TTL}"
Environment="SERVER_PORT=${SERVER_PORT}"
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
if ! systemctl --no-pager --full status pdb-proxy; then
    echo "服务状态检查失败，请手动执行：systemctl status pdb-proxy"
fi
