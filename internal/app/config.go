package app

import (
	"os"
	"path/filepath"

	"github.com/eatmoreapple/go-runcat/internal/resource"
	"github.com/eatmoreapple/go-runcat/internal/systray"
	"github.com/eatmoreapple/go-runcat/internal/theme"
	"github.com/spf13/viper"
)

// Config 应用程序配置
type Config struct {
	// 当前选择的角色
	Runner string `mapstructure:"runner"`
	// 当前主题设置
	Theme string `mapstructure:"theme"`
	// 当前速度限制
	SpeedLimit string `mapstructure:"speed_limit"`
}

// ConfigManager 配置管理器
type ConfigManager struct {
	// 配置文件路径
	configPath string
	// viper实例
	viper *viper.Viper
	// 当前配置
	config Config
}

// NewConfigManager 创建一个新的配置管理器
func NewConfigManager() (*ConfigManager, error) {
	// 获取用户配置目录
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}

	// 创建应用程序配置目录
	appConfigDir := filepath.Join(configDir, "go-runcat")
	if err := os.MkdirAll(appConfigDir, 0755); err != nil {
		return nil, err
	}

	// 配置文件路径
	configPath := filepath.Join(appConfigDir, "config.yaml")

	// 创建viper实例
	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	// 设置默认值
	v.SetDefault("runner", string(resource.RunnerCat))
	v.SetDefault("theme", string(theme.AutoType))
	v.SetDefault("speed_limit", string(systray.SpeedDefault))

	// 创建配置管理器
	cm := &ConfigManager{
		configPath: configPath,
		viper:      v,
	}

	// 加载配置
	if err := cm.Load(); err != nil {
		// 如果配置文件不存在，使用默认配置
		if os.IsNotExist(err) {
			cm.config = Config{
				Runner:     string(resource.RunnerCat),
				Theme:      string(theme.AutoType),
				SpeedLimit: string(systray.SpeedDefault),
			}
			// 保存默认配置
			if err := cm.Save(); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	return cm, nil
}

// Load 加载配置
func (m *ConfigManager) Load() error {
	// 检查配置文件是否存在
	if _, err := os.Stat(m.configPath); os.IsNotExist(err) {
		return err
	}

	// 读取配置文件
	if err := m.viper.ReadInConfig(); err != nil {
		return err
	}

	// 解析配置
	if err := m.viper.Unmarshal(&m.config); err != nil {
		return err
	}

	return nil
}

// Save 保存配置
func (m *ConfigManager) Save() error {
	// 更新viper配置
	m.viper.Set("runner", m.config.Runner)
	m.viper.Set("theme", m.config.Theme)
	m.viper.Set("speed_limit", m.config.SpeedLimit)

	// 写入配置文件
	return m.viper.WriteConfig()
}

// GetConfig 获取当前配置
func (m *ConfigManager) GetConfig() Config {
	return m.config
}

// SetRunner 设置当前角色
func (m *ConfigManager) SetRunner(runner resource.RunnerType) error {
	m.config.Runner = string(runner)
	return m.Save()
}

// SetTheme 设置当前主题
func (m *ConfigManager) SetTheme(theme theme.Type) error {
	m.config.Theme = string(theme)
	return m.Save()
}

// SetSpeedLimit 设置当前速度限制
func (m *ConfigManager) SetSpeedLimit(speedLimit systray.SpeedLimitType) error {
	m.config.SpeedLimit = string(speedLimit)
	return m.Save()
}
