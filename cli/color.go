package cli

import "fmt"

var enableColor = true

func DisableColor() {
	enableColor = false
}

func ColorBlack(text string) string {
	if !enableColor {
		return text
	}
	return fmt.Sprintf("\u001B[0;30m%s\u001B[0m", text)
}

func ColorRed(text string) string {
	if !enableColor {
		return text
	}
	return fmt.Sprintf("\u001B[0;31m%s\u001B[0m", text)
}

func ColorRedBold(text string) string {
	if !enableColor {
		return text
	}
	return fmt.Sprintf("\u001B[1;31m%s\u001B[0m", text)
}

func ColorGreen(text string) string {
	if !enableColor {
		return text
	}
	return fmt.Sprintf("\u001B[0;32m%s\u001B[0m", text)
}

func ColorYellow(text string) string {
	if !enableColor {
		return text
	}
	return fmt.Sprintf("\u001B[0;33m%s\u001B[0m", text)
}

func ColorBlue(text string) string {
	if !enableColor {
		return text
	}
	return fmt.Sprintf("\u001B[0;34m%s\u001B[0m", text)
}

func ColorPurple(text string) string {
	if !enableColor {
		return text
	}
	return fmt.Sprintf("\u001B[0;35m%s\u001B[0m", text)
}

func ColorCyan(text string) string {
	if !enableColor {
		return text
	}
	return fmt.Sprintf("\u001B[0;36m%s\u001B[0m", text)
}

func ColorWhite(text string) string {
	if !enableColor {
		return text
	}
	return fmt.Sprintf("\u001B[0;37m%s\u001B[0m", text)
}

// ---

// Deprecated: use color.Generated
func ColorGenerated(text string) string {
	return ColorGreen(text)
}

// Deprecated: use color.Source
func ColorSource(text string) string {
	return ColorBlue(text)
}

// Deprecated: use color.Binary
func ColorBinary(text string) string {
	return ColorWhite(text)
}

// Deprecated: use color.Version
func ColorVersion(text string) string {
	return text
}

// Deprecated: use color.Input
func ColorInput(text string) string {
	return ColorCyan(text)
}

// Deprecated: use color.Package
func ColorPackage(text string) string {
	return ColorYellow(text)
}
