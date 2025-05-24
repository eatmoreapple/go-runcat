package resource

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
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

var supportedRunners = []RunnerType{
	RunnerCat,
	RunnerParrot,
	RunnerHorse,
}

// Manager 资源管理器
type Manager struct {
	// 嵌入的资源文件
	fs fs.FS
	// 缓存的图标资源
	icons map[string][][]byte
	// 图标计数
	iconCounts map[RunnerType]map[theme.Type]int
}

// NewResourceManager 创建一个新的资源管理器
func NewResourceManager(fs fs.FS) *Manager {
	rm := &Manager{
		fs:         fs,
		icons:      make(map[string][][]byte),
		iconCounts: make(map[RunnerType]map[theme.Type]int),
	}

	// 初始化图标计数
	rm.initIconCounts()

	return rm
}

// 初始化图标计数
func (m *Manager) initIconCounts() {
	for _, runner := range supportedRunners {
		m.iconCounts[runner] = make(map[theme.Type]int)
	}

	var supportThemes = []theme.Type{
		theme.LightType,
		theme.DarkType,
	}

	// 尝试从文件系统中获取实际计数
	for runner := range m.iconCounts {
		for _, t := range supportThemes {
			// note: do not use filepath.Join here
			pattern := fmt.Sprintf("assets/%s/%s/*.ico", runner, t)
			files, err := fs.Glob(m.fs, pattern)
			if err == nil && len(files) > 0 {
				m.iconCounts[runner][t] = len(files)
			}
		}
	}
}

// LoadIcons 加载指定角色和主题的图标
func (m *Manager) LoadIcons(runner RunnerType, themeType theme.Type) ([][]byte, error) {
	// 生成缓存键
	key := fmt.Sprintf("%s_%s", themeType, runner)

	// 检查缓存
	if icons, ok := m.icons[key]; ok {
		return icons, nil
	}

	// 获取图标数量
	count := m.GetIconCount(runner, themeType)
	if count == 0 {
		return nil, fmt.Errorf("no icons found for runner: %s, %s", runner, themeType)
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
		path := fmt.Sprintf("assets/%s/%s/%s_%s_%d.ico", runner, themeType, themeType, strings.ToLower(string(runner)), i)

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
func (m *Manager) GetIconCount(runner RunnerType, themeType theme.Type) int {
	count, ok := m.iconCounts[runner]
	if !ok {
		return 0
	}
	return count[themeType]
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
