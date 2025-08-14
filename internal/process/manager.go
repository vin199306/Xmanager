package process

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/example/program-manager/internal/storage"
)

// Manager 进程管理器
type Manager struct {
	programs map[string]*storage.Program
}

// NewManager 创建进程管理器
func NewManager() *Manager {
	return &Manager{
		programs: make(map[string]*storage.Program),
	}
}

// StartProgram 启动程序
func (m *Manager) StartProgram(program *storage.Program) error {
	if program == nil {
		return fmt.Errorf("程序信息不能为空")
	}

	// 检查程序是否已在运行
	if program.IsRunning {
		return fmt.Errorf("程序 %s 已在运行", program.Name)
	}

	// 创建命令
	cmd := exec.Command("cmd.exe", "/c", program.Command)

	// 设置工作目录
	if program.StartDir != "" {
		cmd.Dir = program.StartDir
	}

	// 隐藏控制台窗口
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true,
	}

	// 创建上下文，支持取消操作
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 启动命令
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动程序失败: %w", err)
	}

	// 更新程序状态
	program.IsRunning = true
	program.PID = cmd.Process.Pid
	m.programs[program.ID] = program

	// 监控进程状态
	go func() {
		// 等待进程结束
		if err := cmd.Wait(); err != nil {
			// 记录错误，但不阻塞主流程
			fmt.Printf("程序 %s 运行出错: %v\n", program.Name, err)
		}

		// 更新程序状态
		program.IsRunning = false
		program.PID = 0
		d delete(m.programs, program.ID)
	}()

	return nil
}

// StopProgram 停止程序
func (m *Manager) StopProgram(program *storage.Program) error {
	if program == nil {
		return fmt.Errorf("程序信息不能为空")
	}

	// 检查程序是否在运行
	if !program.IsRunning || program.PID <= 0 {
		return fmt.Errorf("程序 %s 未在运行", program.Name)
	}

	// 获取进程
	process, err := os.FindProcess(program.PID)
	if err != nil {
		return fmt.Errorf("查找进程失败: %w", err)
	}

	// 终止进程
	if err := process.Kill(); err != nil {
		return fmt.Errorf("终止进程失败: %w", err)
	}

	// 更新程序状态
	program.IsRunning = false
	program.PID = 0
	delete(m.programs, program.ID)

	return nil
}

// StopAllPrograms 停止所有运行中的程序
func (m *Manager) StopAllPrograms() {
	for _, program := range m.programs {
		if program.IsRunning {
			if err := m.StopProgram(program); err != nil {
				fmt.Printf("停止程序 %s 失败: %v\n", program.Name, err)
			}
		}
	}
}

// RefreshStatus 刷新程序状态
func (m *Manager) RefreshStatus(programs []*storage.Program) {
	for _, program := range programs {
		// 检查程序是否在我们的管理列表中
		if p, exists := m.programs[program.ID]; exists {
			// 更新状态
			program.IsRunning = p.IsRunning
			program.PID = p.PID
		} else if program.IsRunning {
			// 如果不在管理列表但标记为运行中，则检查实际状态
			process, err := os.FindProcess(program.PID)
			if err != nil || process.Signal(syscall.Signal(0)) != nil {
				// 进程不存在或无法访问，标记为未运行
				program.IsRunning = false
				program.PID = 0
			}
		}
	}
}

// StartAutoStartPrograms 启动所有设置为开机启动的程序
func (m *Manager) StartAutoStartPrograms(programs []*storage.Program) {
	for _, program := range programs {
		if program.AutoStart && !program.IsRunning {
			if err := m.StartProgram(program); err != nil {
				fmt.Printf("启动开机启动程序 %s 失败: %v\n", program.Name, err)
			}
		}
	}
}