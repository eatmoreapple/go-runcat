package app

import (
	"io/fs"
	"time"

	"github.com/eatmoreapple/go-runcat/internal/monitor"
	"github.com/eatmoreapple/go-runcat/internal/platform"
	"github.com/eatmoreapple/go-runcat/internal/resource"
	"github.com/eatmoreapple/go-runcat/internal/systray"
	"github.com/eatmoreapple/go-runcat/internal/theme"
)

// App 应用程序
type App struct {
	// 配置管理器
	configManager *ConfigManager
	// 平台实现
	platform platform.Platform
	// 资源管理器
	resourceManager *resource.ResourceManager
	// 主题管理器
	themeManager *theme.Manager
	// 系统托盘管理器
	systrayManager *systray.Manager
	// CPU监控器
	cpuMonitor *monitor.CPUMonitor
}

// NewApp 创建一个新的应用程序实例
func NewApp(fs fs.FS) (*App, error) {
	// 创建配置管理器
	configManager, err := NewConfigManager()
	if err != nil {
		return nil, err
	}

	// 创建平台实现
	p := platform.NewPlatform()

	// 创建资源管理器
	rm := resource.NewResourceManager(fs)

	// 创建主题管理器
	tm := theme.NewManager(p)

	// 设置主题
	config := configManager.GetConfig()
	tm.SetTheme(theme.Type(config.Theme))

	// 创建系统托盘管理器
	sm := systray.NewSystrayManager(p, rm, tm)

	// 创建CPU监控器
	cm := monitor.NewCPUMonitor(3 * time.Second)

	return &App{
		configManager:   configManager,
		platform:        p,
		resourceManager: rm,
		themeManager:    tm,
		systrayManager:  sm,
		cpuMonitor:      cm,
	}, nil
}

// Run 运行应用程序
func (a *App) Run() error {
	// 设置CPU使用率更新回调
	a.cpuMonitor.OnUpdate = func(usage float64) {
		a.systrayManager.SetCPUUsage(usage)
	}

	// 启动CPU监控
	a.cpuMonitor.Start()

	// 启动系统托盘
	a.systrayManager.Start()

	return nil
}
