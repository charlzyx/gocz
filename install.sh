#!/bin/bash

# 检查系统是否是 Linux、macOS 或其他
OS_TYPE=$(uname -s)
ARCH_TYPE=$(uname -m)

# 设置二进制包的名称
REPO="charlzyx/gocz"
BIN_NAME="gocz"

# 获取最新的 release 版本号
LATEST_TAG=$(curl -s https://api.github.com/repos/${REPO}/releases/latest | jq -r .tag_name)

# 根据操作系统和架构选择下载地址
if [[ "$OS_TYPE" == "Darwin" ]]; then
  if [[ "$ARCH_TYPE" == "x86_64" ]]; then
    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${LATEST_TAG}/${BIN_NAME}-${LATEST_TAG}-darwin-amd64"
  elif [[ "$ARCH_TYPE" == "arm64" ]]; then
    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${LATEST_TAG}/${BIN_NAME}-${LATEST_TAG}-darwin-arm64"
  else
    echo "Unsupported architecture: $ARCH_TYPE"
    exit 1
  fi
elif [[ "$OS_TYPE" == "Linux" ]]; then
  if [[ "$ARCH_TYPE" == "x86_64" ]]; then
    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${LATEST_TAG}/${BIN_NAME}-${LATEST_TAG}-linux-amd64"
  else
    echo "Unsupported architecture: $ARCH_TYPE"
    exit 1
  fi
elif [[ "$OS_TYPE" == "Windows" ]]; then
  echo "Windows is not supported by this script. Please manually download from GitHub."
  exit 1
else
  echo "Unsupported OS: $OS_TYPE"
  exit 1
fi

# 下载二进制文件
echo "Downloading ${BIN_NAME} ${LATEST_TAG} for ${OS_TYPE}-${ARCH_TYPE}..."
curl -sSL -o /tmp/${BIN_NAME} ${DOWNLOAD_URL}

# 安装到 /usr/local/bin
echo "Installing ${BIN_NAME} to /usr/local/bin..."
sudo mv /tmp/${BIN_NAME} /usr/local/bin/${BIN_NAME}

# 确保二进制文件可执行
sudo chmod +x /usr/local/bin/${BIN_NAME}

# 检查安装是否成功
if command -v ${BIN_NAME} &>/dev/null; then
  echo "${BIN_NAME} installed successfully!"
else
  echo "Failed to install ${BIN_NAME}. Please check the error messages."
  exit 1
fi
