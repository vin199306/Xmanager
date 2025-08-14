package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Program 表示一个程序的信息
type Program struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Command     string    `json:"command"`
	StartDir    string    `json:"start_dir"`
	Description string    `json:"description"`
	IsRunning   bool      `json:"is_running"`
	PID         int       `json:"pid,omitempty"`
	AutoStart   bool      `json:"auto_start"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Store 数据存储接口实现
type Store struct {
	fileName string
}

// NewStore 创建一个新的存储实例
func NewStore() *Store {
	// 获取应用数据目录
	appDataDir, err := os.UserConfigDir()
	if err != nil {
		appDataDir = os.TempDir()
	}

	// 创建应用特定目录
	appDir := filepath.Join(appDataDir, "program-manager")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		fmt.Printf("创建应用目录失败: %v\n", err)
	}

	return &Store{
		fileName: filepath.Join(appDir, "programs.json"),
	}
}

// LoadPrograms 加载程序列表
func (s *Store) LoadPrograms() ([]*Program, error) {
	// 检查文件是否存在
	if _, err := os.Stat(s.fileName); os.IsNotExist(err) {
		return []*Program{}, nil
	}

	// 读取文件内容
	data, err := os.ReadFile(s.fileName)
	if err != nil {
		return nil, fmt.Errorf("读取程序列表失败: %w", err)
	}

	// 解析JSON
	var programs []*Program
	if err := json.Unmarshal(data, &programs); err != nil {
		return nil, fmt.Errorf("解析程序列表失败: %w", err)
	}

	return programs, nil
}

// SavePrograms 保存程序列表
func (s *Store) SavePrograms(programs []*Program) error {
	// 更新修改时间
	for _, p := range programs {
		p.UpdatedAt = time.Now()
	}

	// 序列化为JSON
	data, err := json.MarshalIndent(programs, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化程序列表失败: %w", err)
	}

	// 写入文件
	if err := os.WriteFile(s.fileName, data, 0644); err != nil {
		return fmt.Errorf("写入程序列表失败: %w", err)
	}

	return nil
}

// AddProgram 添加新程序
func (s *Store) AddProgram(program *Program) error {
	if program == nil {
		return errors.New("程序信息不能为空")
	}

	programs, err := s.LoadPrograms()
	if err != nil {
		return err
	}

	// 检查ID是否已存在
	for _, p := range programs {
		if p.ID == program.ID {
			return fmt.Errorf("程序ID已存在: %s", program.ID)
		}
	}

	// 设置创建和更新时间
	now := time.Now()
	program.CreatedAt = now
	program.UpdatedAt = now

	// 添加到列表
	programs = append(programs, program)

	// 保存
	return s.SavePrograms(programs)
}

// DeleteProgram 删除程序
func (s *Store) DeleteProgram(id string) error {
	programs, err := s.LoadPrograms()
	if err != nil {
		return err
	}

	// 查找并删除
	newPrograms := []*Program{}
	found := false
	for _, p := range programs {
		if p.ID != id {
			newPrograms = append(newPrograms, p)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("未找到程序: %s", id)
	}

	// 保存
	return s.SavePrograms(newPrograms)
}

// ExportPrograms 导出程序列表
func (s *Store) ExportPrograms(filePath string) error {
	programs, err := s.LoadPrograms()
	if err != nil {
		return err
	}

	// 序列化为JSON
	data, err := json.MarshalIndent(programs, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化程序列表失败: %w", err)
	}

	// 写入文件
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("写入导出文件失败: %w", err)
	}

	return nil
}

// ImportPrograms 导入程序列表
func (s *Store) ImportPrograms(filePath string) error {
	// 读取文件内容
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("读取导入文件失败: %w", err)
	}

	// 解析JSON
	var programs []*Program
	if err := json.Unmarshal(data, &programs); err != nil {
		return fmt.Errorf("解析导入文件失败: %w", err)
	}

	// 保存
	return s.SavePrograms(programs)
}