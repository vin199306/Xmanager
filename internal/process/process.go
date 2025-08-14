package process

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

// Manager 进程管理器
type Manager struct {
	// 可以添加进程管理相关的字段
}

// NewManager 创建一个新的进程管理器
func NewManager() *Manager {
	return &Manager{}
}

// StartProcess 启动一个新进程
func (m *Manager) StartProcess(command string, startDir string) (int, error) {
	cmd := exec.Command("cmd.exe", "/c", command)
	cmd.Dir = startDir
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true,
	}

	if err := cmd.Start(); err != nil {
		return 0, fmt.Errorf("启动进程失败: %w", err)
	}

	return cmd.Process.Pid, nil
}

// StopProcess 停止进程
func (m *Manager) StopProcess(pid int) error {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("查找进程失败: %w", err)
	}

	return proc.Kill()
}