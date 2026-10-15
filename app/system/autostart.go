package system

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	AppName   = "GoClip"
	PlistName = "com.goclip.app.plist"
)

// IsAutoStartEnabled 检查当前系统是否已配置开机自启
func IsAutoStartEnabled() bool {
	switch runtime.GOOS {
	case "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			return false
		}
		plistPath := filepath.Join(home, "Library", "LaunchAgents", PlistName)
		_, err = os.Stat(plistPath)
		return err == nil
	case "windows":
		cmd := exec.Command("reg", "query", `HKCU\Software\Microsoft\Windows\CurrentVersion\Run`, "/v", AppName)
		out, err := cmd.Output()
		return err == nil && strings.Contains(string(out), AppName)
	case "linux":
		home, err := os.UserHomeDir()
		if err != nil {
			return false
		}
		desktopPath := filepath.Join(home, ".config", "autostart", "goclip.desktop")
		_, err = os.Stat(desktopPath)
		return err == nil
	default:
		return false
	}
}

// SetAutoStart 设置或移除系统级开机启动项
func SetAutoStart(enabled bool, silentMode bool) error {
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("获取程序路径失败: %w", err)
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("解析程序绝对路径失败: %w", err)
	}

	switch runtime.GOOS {
	case "darwin":
		return setAutoStartDarwin(enabled, execPath, silentMode)
	case "windows":
		return setAutoStartWindows(enabled, execPath, silentMode)
	case "linux":
		return setAutoStartLinux(enabled, execPath, silentMode)
	default:
		return fmt.Errorf("当前操作系统不支持自动配置开机启动: %s", runtime.GOOS)
	}
}

func setAutoStartDarwin(enabled bool, execPath string, silentMode bool) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	agentDir := filepath.Join(home, "Library", "LaunchAgents")
	plistPath := filepath.Join(agentDir, PlistName)

	if !enabled {
		if _, err := os.Stat(plistPath); err == nil {
			_ = exec.Command("launchctl", "unload", plistPath).Run()
			return os.Remove(plistPath)
		}
		return nil
	}

	if err := os.MkdirAll(agentDir, 0755); err != nil {
		return err
	}

	silentArg := ""
	if silentMode {
		silentArg = "\n        <string>--silent</string>"
	}

	// 生成标准 LaunchAgent plist
	plistContent := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.goclip.app</string>
    <key>ProgramArguments</key>
    <array>
        <string>%s</string>%s
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>ProcessType</key>
    <string>Interactive</string>
</dict>
</plist>`, execPath, silentArg)

	if err := os.WriteFile(plistPath, []byte(plistContent), 0644); err != nil {
		return err
	}

	// 注册到 launchctl
	_ = exec.Command("launchctl", "load", plistPath).Run()
	return nil
}

func setAutoStartWindows(enabled bool, execPath string, silentMode bool) error {
	if !enabled {
		cmd := exec.Command("reg", "delete", `HKCU\Software\Microsoft\Windows\CurrentVersion\Run`, "/v", AppName, "/f")
		return cmd.Run()
	}

	cmdLine := fmt.Sprintf(`"%s"`, execPath)
	if silentMode {
		cmdLine += " --silent"
	}

	cmd := exec.Command("reg", "add", `HKCU\Software\Microsoft\Windows\CurrentVersion\Run`, "/v", AppName, "/t", "REG_SZ", "/d", cmdLine, "/f")
	return cmd.Run()
}

func setAutoStartLinux(enabled bool, execPath string, silentMode bool) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	autostartDir := filepath.Join(home, ".config", "autostart")
	desktopPath := filepath.Join(autostartDir, "goclip.desktop")

	if !enabled {
		if _, err := os.Stat(desktopPath); err == nil {
			return os.Remove(desktopPath)
		}
		return nil
	}

	if err := os.MkdirAll(autostartDir, 0755); err != nil {
		return err
	}

	cmdLine := execPath
	if silentMode {
		cmdLine += " --silent"
	}

	desktopContent := fmt.Sprintf(`[Desktop Entry]
Type=Application
Version=1.0
Name=GoClip
Comment=GoClip Clipboard Manager
Exec=%s
StartupNotify=false
Terminal=false
`, cmdLine)

	return os.WriteFile(desktopPath, []byte(desktopContent), 0644)
}
