package platform

// Platform 定义平台特定功能的接口
type Platform interface {
	// GetSystemTheme 获取系统主题 (light/dark)
	GetSystemTheme() string

	// SetStartup 设置开机自启动
	SetStartup(enable bool) error

	// IsStartupEnabled 检查是否已设置开机自启动
	IsStartupEnabled() (bool, error)

	// OpenTaskManager 打开系统任务管理器或活动监视器
	OpenTaskManager() error
}

// NewPlatform 根据当前操作系统创建平台实现
func NewPlatform() Platform {
	// 根据操作系统返回对应实现
	// 在各平台特定文件中实现
	return newPlatform()
}
