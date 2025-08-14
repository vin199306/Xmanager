# 设置环境变量
$env:CGO_ENABLED = '1'
$env:CGO_LDFLAGS = '-lopengl32'

# 输出环境变量值进行调试
Write-Host "CGO_ENABLED: $env:CGO_ENABLED"
Write-Host "CGO_LDFLAGS: $env:CGO_LDFLAGS"

# 构建应用程序
Write-Host "开始构建程序..."
Set-Location -Path "e:\Users\Administrator\Desktop\soft_trae"

try {
    go build -tags 'windows' -ldflags '-H windowsgui' -o program-manager.exe ./cmd/app
    if ($LASTEXITCODE -eq 0) {
        Write-Host "构建成功！程序已保存为 program-manager.exe"
    } else {
        Write-Host "构建失败，退出代码: $LASTEXITCODE"
        exit 1
    }
} catch {
    Write-Host "构建过程中发生错误: $_"
    exit 1
}