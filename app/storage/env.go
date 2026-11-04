package storage

import "os"

// isDevelopmentMode 判断当前是否处于开发环境
func isDevelopmentMode() bool {
	if isBuildDevTag {
		return true
	}
	if os.Getenv("APP_ENV") == "development" {
		return true
	}
	if os.Getenv("WAILS_ENV") == "dev" {
		return true
	}
	for _, arg := range os.Args {
		if arg == "--dev" {
			return true
		}
	}
	return false
}

// GetAppDataDir 获取当前环境的数据存储目录
func GetAppDataDir() string {
	if isDevelopmentMode() {
		return ".goclip-dev"
	}
	return ".goclip"
}
