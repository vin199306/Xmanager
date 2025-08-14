package app

import (
	"github.com/example/program-manager/internal/ui"
	"github.com/example/program-manager/internal/storage"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

// Run 启动应用程序
func Run() {
	// 创建Fyne应用
	myApp := app.New()
	myWindow := myApp.NewWindow("程序管理工具")

	// 初始化数据存储
	store := storage.NewStore()

	// 加载程序列表
	programs, err := store.LoadPrograms()
	if err != nil {
		// 处理错误
	}

	// 创建UI
	ui := ui.NewUI(myWindow, store, programs)
	ui.BuildUI()

	// 显示窗口并运行应用
	myWindow.ShowAndRun()
}