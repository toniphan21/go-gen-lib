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

func ColorGenerated(text string) string {
	return ColorGreen(text)
}

func ColorSource(text string) string {
	return ColorBlue(text)
}

func ColorBinary(text string) string {
	return ColorWhite(text)
}

func ColorVersion(text string) string {
	return text
}

func ColorInput(text string) string {
	return ColorCyan(text)
}

func ColorPackage(text string) string {
	return ColorYellow(text)
}
