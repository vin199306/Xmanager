# 程序管理工具编译说明

## 环境要求
- Go 1.21或更高版本
- GCC编译器 (Windows上可通过MinGW安装)
- Fyne依赖: `go get fyne.io/fyne/v2`

## 编译步骤

### 本地编译
1. 克隆代码仓库
   ```bash
   git clone https://github.com/example/program-manager.git
   cd program-manager
   ```

2. 安装依赖
   ```bash
   go mod tidy
   ```

3. 编译程序
   ```bash
   # Windows (生成无控制台窗口的GUI程序)
   go build -ldflags="-s -w -H=windowsgui" -o program-manager.exe ./cmd/app
   
   # Linux/macOS (仅支持基本功能，无系统托盘和开机启动)
   go build -o program-manager ./cmd/app
   ```

### 修改版本号
版本号在编译时通过ldflags参数设置，可以修改GitHub Actions工作流文件或本地编译命令：
```bash
# 在编译命令中添加版本信息
go build -ldflags="-s -w -H=windowsgui -X main.version=1.0.0" -o program-manager.exe ./cmd/app
```

### 添加自定义图标
1. 准备一个PNG格式的图标文件（建议尺寸为256x256像素）
2. 使用Fyne提供的`fyne bundle`工具将图标打包：
   ```bash
   # 安装fyne命令行工具
   go install fyne.io/fyne/v2/cmd/fyne@latest
   
   # 打包图标
   fyne bundle -o internal/ui/icon.go icon.png
   ```
3. 在代码中使用打包的图标（修改internal/ui/ui.go）：
   ```go
   // 导入图标
   import "github.com/example/program-manager/internal/ui"
   
   // 在BuildUI函数中设置窗口图标
   ui.window.SetIcon(ui.MyIcon)
   ```

## GitHub Actions自动编译
本项目包含GitHub Actions工作流配置（.github/workflows/build.yml），支持以下功能：
- 推送到main分支时自动编译
- 手动触发编译
- 生成带版本信息的可执行文件
- 打包为ZIP格式并上传为artifacts

### 工作流参数
- `go-version`: Go语言版本，默认为1.21
- 编译产物会包含版本号、提交哈希和构建时间

## 常见问题

### 1. 编译时提示缺少GCC
在Windows上安装MinGW：
```bash
# 使用choco包管理器
choco install mingw -y
# 或手动下载安装: https://sourceforge.net/projects/mingw/
```

### 2. 程序运行时提示缺少DLL
确保使用静态链接编译：
```bash
go build -ldflags="-s -w -H=windowsgui" -o program-manager.exe ./cmd/app
```

### 3. 开机启动功能不生效
- 确保程序具有管理员权限
- 检查注册表项 `HKEY_CURRENT_USER\Software\Microsoft\Windows\CurrentVersion\Run` 中是否存在ProgramManager项

### 4. 系统托盘不显示
- Windows 10/11可能会将托盘图标隐藏在溢出菜单中
- 尝试自定义任务栏设置，将程序图标固定到任务栏