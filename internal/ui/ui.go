package ui

import (
	"fmt"
	"github.com/example/program-manager/internal/process"
	"github.com/example/program-manager/internal/startup"
	"github.com/example/program-manager/internal/storage"
	"github.com/google/uuid"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// UI 表示应用程序的用户界面
type UI struct {
	window    fyne.Window
	store     *storage.Store
	programs  []*storage.Program
	processManager *process.Manager

	// UI组件
	programList *widget.List
	statusLabel *widget.Label
}

// NewUI 创建一个新的UI实例
func NewUI(window fyne.Window, store *storage.Store, programs []*storage.Program) *UI {
	return &UI{
	window:    window,
	store:     store,
	programs:  programs,
	processManager: process.NewManager(),
	}
}

// BuildUI 构建用户界面
func (ui *UI) BuildUI() {
	// 创建主布局
	mainLayout := container.NewBorder(
		ui.buildToolbar(),
		ui.buildStatusBar(),
		nil,
		nil,
		ui.buildProgramList(),
	)

	// 设置主窗口内容
	ui.window.SetContent(mainLayout)

	// 设置窗口大小
	ui.window.Resize(fyne.NewSize(800, 600))

	// 启动所有设置为开机启动的程序
	ui.processManager.StartAutoStartPrograms(ui.programs)
}

// buildToolbar 构建工具栏
func (ui *UI) buildToolbar() fyne.CanvasObject {
	toolbar := widget.NewToolbar(
		widget.NewToolbarAction(theme.ContentAddIcon(), ui.showAddProgramDialog),
		widget.NewToolbarSeparator(),
		widget.NewToolbarAction(theme.ViewRefreshIcon(), ui.refreshProgramStatus),
		widget.NewToolbarSeparator(),
		widget.NewToolbarAction(theme.SettingsIcon(), ui.showStartupSettingsDialog),
	)

	return toolbar
}

// buildStatusBar 构建状态栏
func (ui *UI) buildStatusBar() fyne.CanvasObject {
	ui.statusLabel = widget.NewLabel("就绪")
	statusBar := container.NewHBox(ui.statusLabel)

	return statusBar
}

// buildProgramList 构建程序列表
func (ui *UI) buildProgramList() fyne.CanvasObject {
	ui.programList = widget.NewList(
		func() int {
			return len(ui.programs)
		},
		func() fyne.CanvasObject {
			return container.NewBorder(
				nil,
				nil,
				container.NewHBox(
					widget.NewLabel("程序名称"),
					widget.NewLabel("状态"),
				),
				container.NewHBox(
					widget.NewButton("启动", func() {}),
					widget.NewButton("停止", func() {}),
					widget.NewButton("删除", func() {}),
				),
			)
		},
		func(i widget.ListItemID, item fyne.CanvasObject) {
			program := ui.programs[i]

			// 更新项目内容
			content := item.(*fyne.Container).Objects[1].(*fyne.Container)
			buttons := item.(*fyne.Container).Objects[3].(*fyne.Container)

			// 更新程序名称和状态
			content.RemoveAll()
			nameLabel := widget.NewLabel(program.Name)
			nameLabel.TextStyle = fyne.TextStyle{Bold: true}

			// 状态标签
			statusLabel := widget.NewLabel("已停止")
			statusLabel.TextStyle = fyne.TextStyle{Bold: true}
			statusLabel.Color = theme.Color(theme.ColorNameDisabled)

			if program.IsRunning {
				statusLabel.SetText("运行中")
				statusLabel.Color = theme.Color(theme.ColorNameSuccess)
			}

			content.Add(nameLabel)
			content.Add(statusLabel)

			// 更新按钮
			buttons.RemoveAll()

			// 启动按钮
			startButton := widget.NewButton("启动", func() {
				ui.startProgram(program)
			})
			startButton.Disable()
			if !program.IsRunning {
				startButton.Enable()
			}

			// 停止按钮
			stopButton := widget.NewButton("停止", func() {
				ui.stopProgram(program)
			})
			stopButton.Disable()
			if program.IsRunning {
				stopButton.Enable()
			}

			// 删除按钮
			deleteButton := widget.NewButton("删除", func() {
				ui.deleteProgram(program)
			})

			buttons.Add(startButton)
			buttons.Add(stopButton)
			buttons.Add(deleteButton)
		},
	)

	return ui.programList
}

// showAddProgramDialog 显示添加程序对话框
func (ui *UI) showAddProgramDialog() {
	// 创建表单
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("输入程序名称")

	commandEntry := widget.NewEntry()
	commandEntry.SetPlaceHolder("输入执行命令")

	startDirEntry := widget.NewEntry()
	startDirEntry.SetPlaceHolder("输入启动目录")

	descEntry := widget.NewMultiLineEntry()
	descEntry.SetPlaceHolder("输入程序描述")

	// 创建表单项目
	items := []*widget.FormItem{
		widget.NewFormItem("程序名称", nameEntry),
		widget.NewFormItem("执行命令", commandEntry),
		widget.NewFormItem("启动目录", startDirEntry),
		widget.NewFormItem("描述", descEntry),
	}

	// 创建对话框
	dialog.ShowForm("添加程序", "确定", "取消", items, func(ok bool) {
		if ok {
			// 验证输入
			if nameEntry.Text == "" || commandEntry.Text == "" {
				dialog.ShowError(fmt.Errorf("程序名称和执行命令不能为空"), ui.window)
				return
			}

			// 创建新程序
			program := &storage.Program{
				ID:          uuid.New().String(),
				Name:        nameEntry.Text,
				Command:     commandEntry.Text,
				StartDir:    startDirEntry.Text,
				Description: descEntry.Text,
				IsRunning:   false,
				AutoStart:   false,
			}

			// 保存程序
			if err := ui.store.AddProgram(program); err != nil {
				dialog.ShowError(fmt.Errorf("添加程序失败: %w", err), ui.window)
				return
			}

			// 更新程序列表
			ui.programs = append(ui.programs, program)
			ui.programList.Refresh()

			// 更新状态
			ui.statusLabel.SetText(fmt.Sprintf("已添加程序: %s", program.Name))
		}
	}, ui.window)
}

// refreshProgramStatus 刷新程序状态
func (ui *UI) refreshProgramStatus() {
	// 刷新程序状态
	ui.processManager.RefreshStatus(ui.programs)

	// 更新列表
	ui.programList.Refresh()

	// 更新状态
	ui.statusLabel.SetText("程序状态已刷新")
}

// showStartupSettingsDialog 显示开机启动设置对话框
func (ui *UI) showStartupSettingsDialog() {
	// 检查当前开机启动状态
	enabled, err := startup.IsAutoStartEnabled()
	if err != nil {
		dialog.ShowError(fmt.Errorf("检查开机启动状态失败: %w", err), ui.window)
		return
	}

	// 创建开关
	autoStartSwitch := widget.NewCheck("开机自动启动", func(checked bool) {
		// 保存设置
		if checked {
			if err := startup.EnableAutoStart(); err != nil {
				dialog.ShowError(fmt.Errorf("启用开机启动失败: %w", err), ui.window)
				autoStartSwitch.SetChecked(false)
				return
			}
		} else {
			if err := startup.DisableAutoStart(); err != nil {
				dialog.ShowError(fmt.Errorf("禁用开机启动失败: %w", err), ui.window)
				autoStartSwitch.SetChecked(true)
				return
			}
		}

		// 更新状态
		if checked {
			ui.statusLabel.SetText("已启用开机启动")
		} else {
			ui.statusLabel.SetText("已禁用开机启动")
		}
	})
	autoStartSwitch.SetChecked(enabled)

	// 创建对话框
	content := container.NewVBox(
		autoStartSwitch,
		widget.NewLabel("重启电脑后设置生效"),
	)
	dialog.ShowCustom("开机启动设置", "关闭", content, ui.window)
}

// startProgram 启动程序
func (ui *UI) startProgram(program *storage.Program) {
	if err := ui.processManager.StartProgram(program); err != nil {
		dialog.ShowError(fmt.Errorf("启动程序失败: %w", err), ui.window)
		return
	}

	// 更新列表
	ui.programList.Refresh()

	// 更新状态
	ui.statusLabel.SetText(fmt.Sprintf("已启动程序: %s", program.Name))
}

// stopProgram 停止程序
func (ui *UI) stopProgram(program *storage.Program) {
	if err := ui.processManager.StopProgram(program); err != nil {
		dialog.ShowError(fmt.Errorf("停止程序失败: %w", err), ui.window)
		return
	}

	// 更新列表
	ui.programList.Refresh()

	// 更新状态
	ui.statusLabel.SetText(fmt.Sprintf("已停止程序: %s", program.Name))
}

// deleteProgram 删除程序
func (ui *UI) deleteProgram(program *storage.Program) {
	// 确认对话框
	dialog.ShowConfirm("确认删除", fmt.Sprintf("确定要删除程序 '%s' 吗?", program.Name), func(ok bool) {
		if ok {
			// 如果程序正在运行，先停止
			if program.IsRunning {
				if err := ui.processManager.StopProgram(program); err != nil {
					dialog.ShowError(fmt.Errorf("停止程序失败: %w", err), ui.window)
					return
				}
			}

			// 删除程序
			if err := ui.store.DeleteProgram(program.ID); err != nil {
				dialog.ShowError(fmt.Errorf("删除程序失败: %w", err), ui.window)
				return
			}

			// 更新程序列表
			newPrograms := []*storage.Program{}
			for _, p := range ui.programs {
				if p.ID != program.ID {
					newPrograms = append(newPrograms, p)
				}
			}
			ui.programs = newPrograms
			ui.programList.Refresh()

			// 更新状态
			ui.statusLabel.SetText(fmt.Sprintf("已删除程序: %s", program.Name))
		}
	}, ui.window)
}