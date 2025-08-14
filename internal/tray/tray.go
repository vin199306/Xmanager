package tray

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"
)

// Manager 系统托盘管理器
type Manager struct {
	app    fyne.App
	window fyne.Window
}

// NewManager 创建系统托盘管理器
func NewManager(myApp fyne.App, myWindow fyne.Window) *Manager {
	return &Manager{
		app:    myApp,
	window: myWindow,
	}
}

// SetupTray 设置系统托盘
func (m *Manager) SetupTray() {
	// 创建系统托盘图标
	ray := m.app.NewSystemTray()

	// 设置托盘图标（可以使用默认图标或自定义图标）
	// 这里使用Fyne的默认图标

	// 处理窗口关闭事件
	m.window.SetCloseIntercept(func() {
		// 最小化到托盘
		m.window.Hide()
	})

	// 设置托盘点击事件
	ray.OnTapped = func() {
		// 显示窗口
		m.window.Show()
		m.window.RequestFocus()
	}

	// 创建托盘菜单
	menu := fyne.NewMenu("程序管理工具")

	// 添加"显示窗口"菜单项
	showItem := fyne.NewMenuItem("显示窗口", func() {
		m.window.Show()
		m.window.RequestFocus()
	})
	menu.Items = append(menu.Items, showItem)

	// 添加"退出"菜单项
	exitItem := fyne.NewMenuItem("退出", func() {
		m.app.Quit()
	})
	menu.Items = append(menu.Items, exitItem)

	// 设置托盘菜单
	ray.SetMenu(menu)
}

// ShowNotification 显示托盘通知
func (m *Manager) ShowNotification(title, content string) {
	m.app.SendNotification(&fyne.Notification{
		Title:   title,
		Content: content,
	})
}