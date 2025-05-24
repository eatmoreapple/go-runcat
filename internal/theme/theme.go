package theme

import (
	"sync"

	"github.com/eatmoreapple/go-runcat/internal/platform"
)

// Type 表示主题类型
type Type string

const (
	// AutoType 自动根据系统主题切换
	AutoType Type = "auto"
	// LightType 浅色主题
	LightType Type = "light"
	// DarkType 深色主题
	DarkType Type = "dark"
)

// Manager 主题管理器
type Manager struct {
	// 当前主题设置
	currentTheme Type
	// 实际使用的主题（考虑到自动主题设置）
	actualTheme Type
	// 平台实现
	platform platform.Platform
	// 主题变化时的回调函数
	onThemeChanged func(theme Type)
	// 互斥锁
	mu sync.RWMutex
}

// NewManager 创建一个新的主题管理器
func NewManager(p platform.Platform) *Manager {
	return &Manager{
		currentTheme: AutoType,
		platform:     p,
		actualTheme:  LightType,
	}
}

// SetTheme 设置主题
func (m *Manager) SetTheme(theme Type) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.currentTheme == theme {
		return
	}

	m.currentTheme = theme
	m.updateActualTheme()
}

// GetTheme 获取当前设置的主题
func (m *Manager) GetTheme() Type {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentTheme
}

// GetActualTheme 获取实际使用的主题
func (m *Manager) GetActualTheme() Type {
	// may cause deadlock here
	//m.mu.RLock()
	//defer m.mu.RUnlock()
	return m.actualTheme
}

// SetOnThemeChanged 设置主题变化时的回调函数
func (m *Manager) SetOnThemeChanged(callback func(theme Type)) {
	m.onThemeChanged = callback
}

// UpdateSystemTheme 更新系统主题
func (m *Manager) UpdateSystemTheme() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.currentTheme == AutoType {
		oldTheme := m.actualTheme
		m.updateActualTheme()

		if oldTheme != m.actualTheme && m.onThemeChanged != nil {
			m.onThemeChanged(m.actualTheme)
		}
	}
}

// 更新实际使用的主题
func (m *Manager) updateActualTheme() {
	if m.currentTheme == AutoType {
		// 获取系统主题
		sysTheme := m.platform.GetSystemTheme()
		if sysTheme == "dark" {
			m.actualTheme = DarkType
		} else {
			m.actualTheme = LightType
		}
	} else {
		m.actualTheme = m.currentTheme
	}
	// 如果设置了回调函数，则调用
	if m.onThemeChanged != nil {
		m.onThemeChanged(m.actualTheme)
	}
}
