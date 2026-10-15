package system

import _ "embed"

//go:embed tray_icon@2x.png
var TrayIconData []byte

type TrayHandler struct {
	OnToggle   func()
	OnSettings func()
	OnQuit     func()
}
