package startup

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows/registry"
)

const (
	// 注册表中开机启动项的路径
	autoStartRegPath = `Software\Microsoft\Windows\CurrentVersion\Run`
	// 程序在注册表中的键名
	appRegKey = "ProgramManager"
)

// IsAutoStartEnabled 检查程序是否已设置为开机启动
func IsAutoStartEnabled() (bool, error) {
	// 打开注册表项
	key, err := registry.OpenKey(registry.CURRENT_USER, autoStartRegPath, registry.QUERY_VALUE)
	if err != nil {
		return false, fmt.Errorf("打开注册表项失败: %w", err)
	}
	defer key.Close()

	// 尝试获取值
	_, _, err = key.GetStringValue(appRegKey)
	if err == nil {
		// 值存在，说明已设置开机启动
		return true, nil
	} else if errors.Is(err, registry.ErrNotExist) {
		// 值不存在，说明未设置开机启动
		return false, nil
	} else {
		// 其他错误
		return false, fmt.Errorf("读取注册表值失败: %w", err)
	}
}

// EnableAutoStart 启用程序开机启动
func EnableAutoStart() error {
	// 获取程序可执行文件路径
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("获取程序路径失败: %w", err)
	}

	// 确保路径是绝对路径
	exePath, err = filepath.Abs(exePath)
	if err != nil {
		return fmt.Errorf("获取绝对路径失败: %w", err)
	}

	// 打开注册表项
	key, err := registry.OpenKey(registry.CURRENT_USER, autoStartRegPath, registry.SET_VALUE)
	if err != nil {
		// 尝试创建注册表项
		key, err = registry.CreateKey(registry.CURRENT_USER, autoStartRegPath, registry.SET_VALUE)
		if err != nil {
			return fmt.Errorf("创建注册表项失败: %w", err)
		}
	}
	defer key.Close()

	// 设置注册表值
	if err := key.SetStringValue(appRegKey, exePath); err != nil {
		return fmt.Errorf("设置注册表值失败: %w", err)
	}

	return nil
}

// DisableAutoStart 禁用程序开机启动
func DisableAutoStart() error {
	// 打开注册表项
	key, err := registry.OpenKey(registry.CURRENT_USER, autoStartRegPath, registry.DELETE_VALUE)
	if err != nil {
		return fmt.Errorf("打开注册表项失败: %w", err)
	}
	defer key.Close()

	// 删除注册表值
	if err := key.DeleteValue(appRegKey); err != nil {
		if errors.Is(err, registry.ErrNotExist) {
			// 值不存在，无需删除
			return nil
		}
		return fmt.Errorf("删除注册表值失败: %w", err)
	}

	return nil
}

// CheckAdminRights 检查是否具有管理员权限
func CheckAdminRights() bool {
	var tokenHandle syscall.Token
	if err := syscall.OpenProcessToken(syscall.GetCurrentProcess(), syscall.TOKEN_QUERY, &tokenHandle); err != nil {
		return false
	}
	defer tokenHandle.Close()

	var elevation uint32
	size := uint32(unsafe.Sizeof(elevation))
	if err := syscall.GetTokenInformation(tokenHandle, syscall.TokenElevation, (*byte)(unsafe.Pointer(&elevation)), size, &size); err != nil {
		return false
	}

	return elevation != 0
}

// RunAsAdmin 以管理员权限重新启动程序
func RunAsAdmin() error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("获取程序路径失败: %w", err)
	}

	verb := "runas"
	cmd := exePath
	args := ""
	curDir, _ := os.Getwd()

	var showCmd int32 = 1 // SW_NORMAL

	return syscall.ShellExecute(0, syscall.StringToUTF16Ptr(verb), syscall.StringToUTF16Ptr(cmd), syscall.StringToUTF16Ptr(args), syscall.StringToUTF16Ptr(curDir), showCmd)
}