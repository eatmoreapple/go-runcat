package resource

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/eatmoreapple/go-runcat/internal/theme"
)

// RunnerType 定义支持的动画角色类型
type RunnerType string

const (
	RunnerCat    RunnerType = "cat"
	RunnerParrot RunnerType = "parrot"
	RunnerHorse  RunnerType = "horse"
)

// Manager 资源管理器
type Manager struct {
	// 嵌入的资源文件
	fs fs.FS
	// 缓存的图标资源
	icons map[string][][]byte
	// 图标计数
	iconCounts map[RunnerType]int
}

// NewResourceManager 创建一个新的资源管理器
func NewResourceManager(fs fs.FS) *Manager {
	rm := &Manager{
		fs:         fs,
		icons:      make(map[string][][]byte),
		iconCounts: make(map[RunnerType]int),
	}

	// 初始化图标计数
	rm.initIconCounts()

	return rm
}

// 初始化图标计数
func (m *Manager) initIconCounts() {
	// 默认值
	m.iconCounts[RunnerCat] = 5
	m.iconCounts[RunnerParrot] = 10
	m.iconCounts[RunnerHorse] = 14

	// 尝试从文件系统中获取实际计数
	for runner := range m.iconCounts {
		// 检查light主题目录
		pattern := fmt.Sprintf("*_%s_*.ico", strings.ToLower(string(runner)))
		lightPath := fmt.Sprintf("assets/%s/light", runner)

		files, err := fs.Glob(m.fs, filepath.Join(lightPath, pattern))
		if err == nil && len(files) > 0 {
			m.iconCounts[runner] = len(files)
		}
	}
}

// LoadIcons 加载指定角色和主题的图标
func (m *Manager) LoadIcons(runner RunnerType, themeType theme.Type) ([][]byte, error) {
	// 转换主题类型为字符串
	themeStr := "light"
	if themeType == theme.DarkType {
		themeStr = "dark"
	}

	// 生成缓存键
	key := fmt.Sprintf("%s_%s", themeStr, runner)

	// 检查缓存
	if icons, ok := m.icons[key]; ok {
		return icons, nil
	}

	// 获取图标数量
	count := m.GetIconCount(runner)
	if count == 0 {
		return nil, fmt.Errorf("no icons found for runner: %s", runner)
	}
	// 加载图标

	readFromFs := func(path string) ([]byte, error) {
		file, err := m.fs.Open(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read icon %s: %w", path, err)
		}
		defer func() { _ = file.Close() }()
		return io.ReadAll(file)
	}

	icons := make([][]byte, count)

	// 遍历图标索引
	for i := 0; i < count; i++ {
		// 构建资源路径
		path := fmt.Sprintf("assets/%s/%s/%s_%s_%d.ico", runner, themeStr, themeStr, strings.ToLower(string(runner)), i)

		// 读取资源文件
		data, err := readFromFs(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read icon %s: %w", path, err)
		}

		icons[i] = data
	}

	// 缓存图标
	m.icons[key] = icons

	return icons, nil
}

// GetIconCount 获取指定角色的图标数量
func (m *Manager) GetIconCount(runner RunnerType) int {
	count, ok := m.iconCounts[runner]
	if !ok {
		return 0
	}
	return count
}

// GetIcon 获取指定角色、主题和索引的图标
func (m *Manager) GetIcon(runner RunnerType, themeType theme.Type, index int) ([]byte, error) {
	icons, err := m.LoadIcons(runner, themeType)
	if err != nil {
		return nil, err
	}

	if index < 0 || index >= len(icons) {
		return nil, fmt.Errorf("icon index out of range: %d", index)
	}

	return icons[index], nil
}

// GetIconReader 获取指定角色、主题和索引的图标读取器
func (m *Manager) GetIconReader(runner RunnerType, themeType theme.Type, index int) (*bytes.Reader, error) {
	icon, err := m.GetIcon(runner, themeType, index)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(icon), nil
}
