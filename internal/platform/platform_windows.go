//go:build windows

package platform

import (
	"errors"
	"os"
	"os/exec"

	"golang.org/x/sys/windows/registry"
)

type windowsPlatform struct{}

func newPlatform() Platform {
	return &windowsPlatform{}
}

// GetSystemTheme 获取Windows系统主题 (light/dark)
func (p *windowsPlatform) GetSystemTheme() string {
	k, err := registry.OpenKey(registry.CURRENT_USER, `SOFTWARE\Microsoft\Windows\CurrentVersion\Themes\Personalize`, registry.QUERY_VALUE)
	if err != nil {
		return "light" // 默认返回light主题
	}
	defer func() { _ = k.Close() }()

	val, _, err := k.GetIntegerValue("SystemUsesLightTheme")
	if err != nil {
		return "light" // 默认返回light主题
	}
	if val == 0 {
		return "dark"
	}
	return "light"
}

// SetStartup 设置Windows开机自启动
func (p *windowsPlatform) SetStartup(enable bool) error {
	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Run`, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer func() { _ = k.Close() }()

	execPath, err := os.Executable()
	if err != nil {
		return err
	}

	if enable {
		return k.SetStringValue("GoRunCat", execPath)
	} else {
		return k.DeleteValue("GoRunCat")
	}
}

// IsStartupEnabled 检查Windows是否已设置开机自启动
func (p *windowsPlatform) IsStartupEnabled() (bool, error) {
	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Run`, registry.QUERY_VALUE)
	if err != nil {
		return false, err
	}
	defer func() { _ = k.Close() }()

	_, _, err = k.GetStringValue("GoRunCat")
	if errors.Is(err, registry.ErrNotExist) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

// OpenTaskManager 打开Windows任务管理器
func (p *windowsPlatform) OpenTaskManager() error {
	cmd := exec.Command("powershell", "-c", "Start-Process", "taskmgr.exe")
	return cmd.Start()
}
