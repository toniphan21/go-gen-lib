package color

import "nhatp.com/go/gen-lib/cli"

func Generated(text string) string {
	return cli.ColorGreen(text)
}

func Source(text string) string {
	return cli.ColorBlue(text)
}

func Binary(text string) string {
	return cli.ColorWhite(text)
}

func Version(text string) string {
	return text
}

func Input(text string) string {
	return cli.ColorCyan(text)
}

func Package(text string) string {
	return cli.ColorYellow(text)
}
