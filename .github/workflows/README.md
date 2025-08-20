# GitHub Actions 工作流指南

本指南说明如何使用配置的 GitHub Actions 工作流进行自动构建和发布。

## 🚀 快速开始

### 1. 推送标签触发发布

要创建一个新的发布版本，只需推送一个以 `v` 开头的标签：

```bash
# 创建新版本标签
git tag v1.0.0
git push origin v1.0.0
```

### 2. 手动触发构建

也可以手动触发构建：

1. 进入 GitHub 仓库的 **Actions** 页面
2. 选择 **Build and Release** 工作流
3. 点击 **Run workflow** 按钮

## 📋 工作流说明

### 主要工作流

#### 🔨 `release.yml` - 构建和发布
- **触发条件**: 推送 `v*` 标签或手动触发
- **功能**:
  - 编译 Linux AMD64 版本
  - 创建发布包
  - 上传到 GitHub Releases
  - 构建并推送 Docker 镜像

#### 🧪 `test.yml` - 测试和构建验证
- **触发条件**: 推送到 `main`/`develop` 分支或创建 PR
- **功能**:
  - 运行单元测试
  - 代码质量检查
  - 安全扫描
  - 构建验证

## 🏗️ 构建产物

### 可执行文件
- `program-manager-linux-amd64`: Linux AMD64 可执行文件
- `program-manager-{version}-linux-amd64.tar.gz`: 完整发布包

### Docker 镜像
- `ghcr.io/{owner}/{repository}:{tag}`: 对应版本标签
- `ghcr.io/{owner}/{repository}:latest`: 最新版本

## 🔧 本地测试

在推送标签前，可以在本地测试构建：

```bash
# 测试构建
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -trimpath -o program-manager-linux .

# 测试 Docker 构建
docker build -t program-manager:test .
```

## 📊 构建状态

可以在 GitHub 仓库的 **Actions** 页面查看所有工作流的运行状态。

## 📝 发布说明

发布说明会自动生成，包含：
- 提交信息
- 构建时间
- 系统要求
- 使用方法

## 🔒 权限要求

确保仓库设置中启用了以下权限：
- **Actions**: 读写权限
- **Contents**: 读写权限
- **Packages**: 写入权限（用于 Docker 镜像）

## 🐛 故障排除

### 构建失败

1. 检查 Go 版本兼容性
2. 验证所有依赖是否可用
3. 查看 Actions 日志获取详细信息

### 发布失败

1. 确认标签格式正确（必须以 `v` 开头）
2. 检查 GitHub Token 权限
3. 验证仓库设置

## 📞 支持

如有问题，请：
1. 查看 Actions 日志
2. 检查工作流配置文件
3. 创建 Issue 寻求帮助