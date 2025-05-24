//go:build darwin

package platform

import (
	"os"
	"os/exec"
	"path/filepath"
)

type darwinPlatform struct{}

func newPlatform() Platform {
	return &darwinPlatform{}
}

// GetSystemTheme 获取macOS系统主题 (light/dark)
func (p *darwinPlatform) GetSystemTheme() string {
	cmd := exec.Command("defaults", "read", "-g", "AppleInterfaceStyle")
	output, err := cmd.Output()
	if err != nil {
		// 如果命令失败，通常意味着系统处于浅色模式
		return "light"
	}

	// 如果输出包含"Dark"，则系统处于深色模式
	if len(output) >= 4 && string(output[:4]) == "Dark" {
		return "dark"
	}
	return "light"
}

// SetStartup 设置macOS开机自启动
func (p *darwinPlatform) SetStartup(enable bool) error {
	// 获取应用程序路径
	execPath, err := os.Executable()
	if err != nil {
		return err
	}

	// 用户的启动项目录
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	launchAgentsDir := filepath.Join(homeDir, "Library", "LaunchAgents")

	// 确保目录存在
	if err := os.MkdirAll(launchAgentsDir, 0755); err != nil {
		return err
	}

	plistPath := filepath.Join(launchAgentsDir, "com.eatmoreapple.goruncat.plist")

	if enable {
		// 创建plist文件内容
		plistContent := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.eatmoreapple.goruncat</string>
    <key>ProgramArguments</key>
    <array>
        <string>` + execPath + `</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
</dict>
</plist>`

		// 写入plist文件
		if err := os.WriteFile(plistPath, []byte(plistContent), 0644); err != nil {
			return err
		}

		// 加载启动项
		cmd := exec.Command("launchctl", "load", plistPath)
		return cmd.Run()
	} else {
		// 卸载启动项
		if _, err := os.Stat(plistPath); err == nil {
			cmd := exec.Command("launchctl", "unload", plistPath)
			if err := cmd.Run(); err != nil {
				return err
			}
			return os.Remove(plistPath)
		}
		return nil
	}
}

// IsStartupEnabled 检查macOS是否已设置开机自启动
func (p *darwinPlatform) IsStartupEnabled() (bool, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false, err
	}

	plistPath := filepath.Join(homeDir, "Library", "LaunchAgents", "com.eatmoreapple.goruncat.plist")
	_, err = os.Stat(plistPath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// OpenTaskManager 打开macOS活动监视器
func (p *darwinPlatform) OpenTaskManager() error {
	cmd := exec.Command("open", "/System/Applications/Utilities/Activity Monitor.app")
	return cmd.Start()
}
