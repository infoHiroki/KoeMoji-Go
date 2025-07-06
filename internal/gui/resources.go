package gui

import (
	_ "embed"
	"fyne.io/fyne/v2"
)

//go:embed icon.png
var iconData []byte

// GetAppIcon returns the embedded application icon
func GetAppIcon() fyne.Resource {
	return &fyne.StaticResource{
		StaticName:    "icon.png",
		StaticContent: iconData,
	}
}