package icons

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/storage"
)

// GetAppIcon returns the application icon resource
func GetAppIcon() fyne.Resource {
	// For now, we'll use a simple placeholder
	// In the future, this can be replaced with a proper icon file
	// The icon should be embedded as a resource using fyne bundle
	
	// Return nil for now - Fyne will use a default icon
	return nil
}

// GetIconURI returns the icon as a URI for window icon
func GetIconURI() fyne.URI {
	// Return nil for now - no custom icon file yet
	return storage.NewFileURI("icon.png") // placeholder
}