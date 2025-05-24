package systray

import (
	"fmt"
	"log"
	"time"

	"github.com/eatmoreapple/go-runcat/internal/platform"
	"github.com/eatmoreapple/go-runcat/internal/resource"
	"github.com/eatmoreapple/go-runcat/internal/theme"
	"github.com/getlantern/systray"
)

// SpeedLimitType 表示速度限制类型
type SpeedLimitType string

const (
	// SpeedDefault 默认速度（根据CPU使用率动态调整）
	SpeedDefault SpeedLimitType = "default"
	// SpeedCPU10 限制为CPU 10%的速度
	SpeedCPU10 SpeedLimitType = "cpu10"
	// SpeedCPU20 限制为CPU 20%的速度
	SpeedCPU20 SpeedLimitType = "cpu20"
	// SpeedCPU30 限制为CPU 30%的速度
	SpeedCPU30 SpeedLimitType = "cpu30"
	// SpeedCPU40 限制为CPU 40%的速度
	SpeedCPU40 SpeedLimitType = "cpu40"
)

// Manager 系统托盘管理器
type Manager struct {
	// 平台实现
	platform platform.Platform
	// 资源管理器
	resourceManager *resource.Manager
	// 主题管理器
	themeManager *theme.Manager

	// 当前选择的角色
	currentRunner resource.RunnerType
	// 当前速度限制
	speedLimit SpeedLimitType
	// 当前CPU使用率
	cpuUsage float64
	// 当前图标索引
	currentIconIndex int
	// 图标切换间隔
	animationInterval time.Duration
	// 最小动画间隔
	minInterval float64

	// 菜单项
	runnerMenu      map[resource.RunnerType]*systray.MenuItem
	themeMenu       map[theme.Type]*systray.MenuItem
	startupMenu     *systray.MenuItem
	speedLimitMenu  map[SpeedLimitType]*systray.MenuItem
	taskManagerMenu *systray.MenuItem

	// 停止动画的通道
	stopAnimationCh chan struct{}
	// 是否正在运行动画
	animationRunning bool

	// 当前图标数据
	currentIcons [][]byte
}

// NewSystrayManager 创建一个新的系统托盘管理器
func NewSystrayManager(
	p platform.Platform,
	rm *resource.Manager,
	tm *theme.Manager,
) *Manager {
	return &Manager{
		platform:          p,
		resourceManager:   rm,
		themeManager:      tm,
		currentRunner:     resource.RunnerCat,
		speedLimit:        SpeedDefault,
		animationInterval: 200 * time.Millisecond,
		minInterval:       25.0,
		runnerMenu:        make(map[resource.RunnerType]*systray.MenuItem),
		themeMenu:         make(map[theme.Type]*systray.MenuItem),
		speedLimitMenu:    make(map[SpeedLimitType]*systray.MenuItem),
		stopAnimationCh:   make(chan struct{}),
	}
}

// Start 启动系统托盘
func (m *Manager) Start() {
	systray.Run(m.onReady, m.onExit)
}

// onReady 系统托盘准备就绪时的回调
func (m *Manager) onReady() {
	// 设置初始图标
	m.updateIcon()

	// 创建菜单项
	m.createMenuItems()

	// 启动动画
	m.startAnimation()

	// 设置主题变化回调
	m.themeManager.SetOnThemeChanged(func(t theme.Type) {
		m.updateIcon()
	})
}

// onExit 系统托盘退出时的回调
func (m *Manager) onExit() {
	m.stopAnimation()
}

// SetCPUUsage 设置CPU使用率
func (m *Manager) SetCPUUsage(usage float64) {
	m.cpuUsage = usage

	// 更新系统托盘提示文本
	systray.SetTooltip(fmt.Sprintf("CPU: %.1f%%", usage))

	// 根据CPU使用率调整动画速度
	if m.speedLimit == SpeedDefault {
		// 根据CPU使用率计算动画间隔
		// 使用与原始RunCat相同的算法
		interval := 200.0 / float64(max(1.0, min(20.0, usage/5.0)))
		m.animationInterval = time.Duration(interval) * time.Millisecond
	}
}

// 创建菜单项
func (m *Manager) createMenuItems() {
	// Runner菜单
	runnerMenuItem := systray.AddMenuItem("Runner", "Select runner")
	m.runnerMenu[resource.RunnerCat] = runnerMenuItem.AddSubMenuItemCheckbox("Cat", "Cat runner", m.currentRunner == resource.RunnerCat)
	m.runnerMenu[resource.RunnerParrot] = runnerMenuItem.AddSubMenuItemCheckbox("Parrot", "Parrot runner", m.currentRunner == resource.RunnerParrot)
	m.runnerMenu[resource.RunnerHorse] = runnerMenuItem.AddSubMenuItemCheckbox("Horse", "Horse runner", m.currentRunner == resource.RunnerHorse)

	// Theme菜单
	themeMenuItem := systray.AddMenuItem("Theme", "Select theme")
	m.themeMenu[theme.AutoType] = themeMenuItem.AddSubMenuItemCheckbox("Auto", "Auto theme", m.themeManager.GetTheme() == theme.AutoType)
	m.themeMenu[theme.LightType] = themeMenuItem.AddSubMenuItemCheckbox("Light", "Light theme", m.themeManager.GetTheme() == theme.LightType)
	m.themeMenu[theme.DarkType] = themeMenuItem.AddSubMenuItemCheckbox("Dark", "Dark theme", m.themeManager.GetTheme() == theme.DarkType)

	// Startup菜单
	startupEnabled, err := m.platform.IsStartupEnabled()
	if err != nil {
		log.Printf("Failed to check startup status: %v", err)
		startupEnabled = false
	}
	m.startupMenu = systray.AddMenuItemCheckbox("Start at Login", "Start at login", startupEnabled)

	// Speed Limit菜单
	speedLimitMenuItem := systray.AddMenuItem("Runner Speed Limit", "Set runner speed limit")
	m.speedLimitMenu[SpeedDefault] = speedLimitMenuItem.AddSubMenuItemCheckbox("Default", "Default speed", m.speedLimit == SpeedDefault)
	m.speedLimitMenu[SpeedCPU10] = speedLimitMenuItem.AddSubMenuItemCheckbox("CPU 10%", "Limit to CPU 10%", m.speedLimit == SpeedCPU10)
	m.speedLimitMenu[SpeedCPU20] = speedLimitMenuItem.AddSubMenuItemCheckbox("CPU 20%", "Limit to CPU 20%", m.speedLimit == SpeedCPU20)
	m.speedLimitMenu[SpeedCPU30] = speedLimitMenuItem.AddSubMenuItemCheckbox("CPU 30%", "Limit to CPU 30%", m.speedLimit == SpeedCPU30)
	m.speedLimitMenu[SpeedCPU40] = speedLimitMenuItem.AddSubMenuItemCheckbox("CPU 40%", "Limit to CPU 40%", m.speedLimit == SpeedCPU40)

	// 分隔线
	systray.AddSeparator()

	// 版本信息
	m.taskManagerMenu = systray.AddMenuItem("Task Manger", "")

	// Author
	systray.AddMenuItem("Author", "eatmoreapple")

	// 退出菜单
	quitItem := systray.AddMenuItem("Quit", "Quit the application")

	// 处理菜单事件
	go m.handleMenuEvents(quitItem)
}

// 处理菜单事件
func (m *Manager) handleMenuEvents(quitItem *systray.MenuItem) {
	// Runner菜单事件
	for runner, item := range m.runnerMenu {
		go func(r resource.RunnerType, i *systray.MenuItem) {
			for range i.ClickedCh {
				m.setRunner(r)
			}
		}(runner, item)
	}

	// Theme菜单事件
	for t, item := range m.themeMenu {
		go func(themeType theme.Type, i *systray.MenuItem) {
			for range i.ClickedCh {
				m.setTheme(themeType)
			}
		}(t, item)
	}

	// Startup菜单事件
	go func() {
		for range m.startupMenu.ClickedCh {
			m.toggleStartup()
		}
	}()

	// Speed Limit菜单事件
	for speed, item := range m.speedLimitMenu {
		go func(s SpeedLimitType, i *systray.MenuItem) {
			for range i.ClickedCh {
				m.setSpeedLimit(s)
			}
		}(speed, item)
	}

	// Task Manager菜单事件
	go func() {
		for range m.taskManagerMenu.ClickedCh {
			if err := m.platform.OpenTaskManager(); err != nil {
				log.Printf("Failed to open task manager: %v", err)
			}
		}
	}()

	// 退出事件
	go func() {
		<-quitItem.ClickedCh
		systray.Quit()
	}()
}

// 设置角色
func (m *Manager) setRunner(runner resource.RunnerType) {
	if m.currentRunner == runner {
		return
	}

	// 更新选中状态
	for r, item := range m.runnerMenu {
		if r == runner {
			item.Check()
		} else {
			item.Uncheck()
		}
	}

	m.currentRunner = runner
	m.currentIconIndex = 0
	m.updateIcon()
}

// 设置主题
func (m *Manager) setTheme(t theme.Type) {
	// 更新选中状态
	for themeType, item := range m.themeMenu {
		if themeType == t {
			item.Check()
		} else {
			item.Uncheck()
		}
	}

	m.themeManager.SetTheme(t)
}

// 切换开机自启动
func (m *Manager) toggleStartup() {
	enabled, _ := m.platform.IsStartupEnabled()
	if err := m.platform.SetStartup(!enabled); err != nil {
		log.Printf("Failed to toggle startup: %v", err)
		return
	}
	// 切换菜单项状态
	if enabled {
		m.startupMenu.Uncheck()
	} else {
		m.startupMenu.Check()
	}
}

// 设置速度限制
func (m *Manager) setSpeedLimit(speed SpeedLimitType) {
	if m.speedLimit == speed {
		return
	}

	// 更新选中状态
	for s, item := range m.speedLimitMenu {
		if s == speed {
			item.Check()
		} else {
			item.Uncheck()
		}
	}

	m.speedLimit = speed

	// 根据速度限制设置动画间隔
	switch speed {
	case SpeedDefault:
		// 使用CPU使用率动态调整
		m.SetCPUUsage(m.cpuUsage)
	case SpeedCPU10:
		m.animationInterval = time.Duration(100) * time.Millisecond
	case SpeedCPU20:
		m.animationInterval = time.Duration(50) * time.Millisecond
	case SpeedCPU30:
		m.animationInterval = time.Duration(33) * time.Millisecond
	case SpeedCPU40:
		m.animationInterval = time.Duration(25) * time.Millisecond
	}
}

// 更新图标
func (m *Manager) updateIcon() {
	// 获取当前主题
	currentTheme := m.themeManager.GetActualTheme()
	// 加载图标
	icons, err := m.resourceManager.LoadIcons(m.currentRunner, currentTheme)
	if err != nil {
		// 图标加载失败，使用默认图标
		log.Printf("Failed to load icons: %v", err)
		return
	}

	// 缓存图标
	m.currentIcons = icons

	// 设置当前图标
	if m.currentIconIndex >= len(icons) {
		m.currentIconIndex = 0
	}

	// 设置系统托盘图标
	systray.SetIcon(icons[m.currentIconIndex])
}

// 启动动画
func (m *Manager) startAnimation() {
	if m.animationRunning {
		return
	}

	m.animationRunning = true
	go func() {
		for {
			select {
			case <-m.stopAnimationCh:
				m.animationRunning = false
				return
			case <-time.After(m.animationInterval):
				// 更新图标索引
				m.currentIconIndex++
				iconCount := m.resourceManager.GetIconCount(m.currentRunner, m.themeManager.GetActualTheme())
				if m.currentIconIndex >= iconCount {
					m.currentIconIndex = 0
				}

				// 更新图标
				m.updateIcon()
			}
		}
	}()
}

// 停止动画
func (m *Manager) stopAnimation() {
	if !m.animationRunning {
		return
	}

	m.stopAnimationCh <- struct{}{}
}
