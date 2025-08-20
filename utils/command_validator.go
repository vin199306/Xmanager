package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

// CommandValidator provides validation for program commands
var CommandValidator = &commandValidator{}

type commandValidator struct {
	// 允许的命令模式
	allowedPatterns []string
	// 禁止的模式
	forbiddenPatterns []string
}

// ValidateCommand validates a program command for security
func (cv *commandValidator) ValidateCommand(command string) error {
	if command == "" {
		return fmt.Errorf("command cannot be empty")
	}
	
	// 去除前后空格
	command = strings.TrimSpace(command)
	
	// 检查命令长度
	if len(command) > 1000 {
		return fmt.Errorf("command too long (max 1000 characters)")
	}
	
	// 检查禁止的模式
	forbiddenPatterns := []string{
		`rm\s+-rf\s+/`,      // 危险的删除命令
		`sudo\s+`,          // 避免提权
		`chmod\s+.*777`,    // 权限提升
		`curl.*\|.*sh`,     // 远程代码执行
		`wget.*\|.*sh`,     // 远程代码执行
		`eval\s+`,          // 代码注入
		`exec\s+`,          // 代码执行
		`system\s*\(`,      // 系统调用
		`\$\(`,             // 命令替换
		`;`,                // 命令链
		`\|\|`,             // 逻辑或
		`&&`,               // 逻辑与
		`>\s*/dev/`,        // 设备文件操作
		`<\s*/dev/`,        // 设备文件操作
	}
	
	for _, pattern := range forbiddenPatterns {
		matched, _ := regexp.MatchString(pattern, command)
		if matched {
			return fmt.Errorf("command contains forbidden pattern: %s", pattern)
		}
	}
	
	// 检查可执行文件是否存在
	if err := cv.validateExecutable(command); err != nil {
		return fmt.Errorf("invalid executable: %v", err)
	}
	
	return nil
}

// validateExecutable checks if the executable part of the command exists
func (cv *commandValidator) validateExecutable(command string) error {
	// 提取命令的第一部分（可执行文件）
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return fmt.Errorf("empty command")
	}
	
	executable := parts[0]
	
	// 处理相对路径
	if strings.Contains(executable, "/") {
		if !filepath.IsAbs(executable) {
			// 相对路径，检查相对于工作目录
			return nil // 将在运行时检查
		}
	}
	
	// 检查绝对路径
	if filepath.IsAbs(executable) {
		if _, err := os.Stat(executable); err != nil {
			return fmt.Errorf("executable not found: %s", executable)
		}
		return nil
	}
	
	// 检查PATH中的可执行文件
	if runtime.GOOS == "windows" {
		// Windows: 检查PATH和扩展名
		extensions := []string{"", ".exe", ".bat", ".cmd"}
		for _, ext := range extensions {
			fullName := executable + ext
			if _, err := exec.LookPath(fullName); err == nil {
				return nil
			}
		}
	} else {
		// Unix-like: 检查PATH
		if _, err := exec.LookPath(executable); err == nil {
			return nil
		}
	}
	
	return fmt.Errorf("executable not found in PATH: %s", executable)
}

// ValidateWorkingDir validates the working directory
func (cv *commandValidator) ValidateWorkingDir(workingDir string) error {
	if workingDir == "" {
		return nil // 允许空值，使用默认目录
	}
	
	// 标准化路径
	workingDir = filepath.Clean(workingDir)
	
	// 检查路径长度
	if len(workingDir) > 500 {
		return fmt.Errorf("working directory path too long")
	}
	
	// 检查是否为绝对路径
	if !filepath.IsAbs(workingDir) {
		return fmt.Errorf("working directory must be absolute path")
	}
	
	// 检查是否存在
	if _, err := os.Stat(workingDir); err != nil {
		if os.IsNotExist(err) {
			// 尝试创建目录
			if err := os.MkdirAll(workingDir, 0755); err != nil {
				return fmt.Errorf("failed to create working directory: %v", err)
			}
		} else {
			return fmt.Errorf("cannot access working directory: %v", err)
		}
	}
	
	// 检查是否为目录
	info, err := os.Stat(workingDir)
	if err != nil {
		return fmt.Errorf("working directory error: %v", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("working directory path is not a directory")
	}
	
	return nil
}

// SanitizeCommand sanitizes a command for safe execution
func (cv *commandValidator) SanitizeCommand(command string) string {
	// 去除控制字符
	command = strings.Map(func(r rune) rune {
		if r < 32 && r != ' ' && r != '\t' && r != '\n' && r != '\r' {
			return -1
		}
		return r
	}, command)
	
	// 去除多余空格
	command = strings.Join(strings.Fields(command), " ")
	
	return command
}

// IsSafeCommand checks if a command is considered safe
func (cv *commandValidator) IsSafeCommand(command string) bool {
	return cv.ValidateCommand(command) == nil
}