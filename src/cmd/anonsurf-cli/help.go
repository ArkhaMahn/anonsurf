package main

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"gitlab.com/parrotsec/packages/anonsurf/internal/console"
)

const (
	blueText     = "\x1b[38;2;0;124;166m"
	darkBlueText = "\x1b[38;2;0;94;133m"
	mutedText    = "\x1b[38;5;245m"
	boldText     = "\x1b[1m"
	resetText    = "\x1b[0m"
)

var logoPixels = [...]string{
	"                                ",
	"           WWWWWWWWWW           ",
	"         WWWWWWWWWWWWWW         ",
	"         WWLLLLLLLLLLWW         ",
	"        WWLLLLLLLLLLLLWW        ",
	"        WWLLLLLLLLLLLLWW        ",
	"        WWLLLLLLLLLLLLWW        ",
	"       WWWLLLLLLLLLLLLWWW       ",
	"    WWWWWWLLLLLLLLLLLLWWWWWW    ",
	"  WWWWWLLWLLLLLLLLLLLLWLLWWWWW  ",
	" WWWLLLLLWWWWWWWWWWWWWWLLLLLWWW ",
	" WWLLLLLLLWWWWWWWWWWWWLLLLLLLWW ",
	" WWLLLLLLLLLLWWWWWWLLLLLLLLLLWW ",
	" WWLLLLLLLLLLLLLLLLLLLLLLLLLLWW ",
	"  WWLLLLLLLLLLLLLLLLLLLLLLLLWW  ",
	"   WWWWLDDDDDDDDDDDDDDDDLWWWW   ",
	"  WWWWDDDDDDDDDDDDDDDDDDDDWWWW  ",
	" WWWDDDDDDDDDDDDDDDDDDDDDDDDWWW ",
	"WWWDDDDDDDDDDDDDDDDDDDDDDDDDDWWW",
	"  WWWDDDDDDDDDDDDDDDDDDDDDDWWW  ",
	"   WWWDDDDDDDDDDDDDDDDDDDDWWW   ",
	"    WWWDDDWDDDDDDDDDDWDDDWWW    ",
	"      WWWDWWDDDDDDDDWWDWWW      ",
	"       WWWWWWDDDDDDWWWWWW       ",
	"        WWWWWWWDDWWWWWWW        ",
	"          WWWWWWWWWWWW          ",
	"           WWWWDDWWWW           ",
	"            WWWDDWWW            ",
	"             WWWWWW             ",
	"               WW               ",
}

var rootHelpEntries = [...]struct {
	name    string
	options string
}{
	{name: "start", options: "[-k]"},
	{name: "stop", options: "[-k]"},
	{name: "restart", options: "[-k]"},
	{name: "status", options: "[-n]"},
	{name: "changeid"},
	{name: "ip"},
	{name: "monitor"},
	{name: "boot", options: "<action>"},
}

func configureRootHelp(root *cobra.Command) {
	defaultHelp := root.HelpFunc()
	root.SetHelpFunc(func(command *cobra.Command, args []string) {
		if command != root {
			defaultHelp(command, args)
			return
		}
		writeRootHelp(command.OutOrStdout(), command)
	})
}

func writeRootHelp(output io.Writer, root *cobra.Command) {
	color := console.SupportsColor(output)
	blue, muted, bold, reset := "", "", "", ""
	if color {
		blue, muted, bold, reset = blueText, mutedText, boldText, resetText
		writeLogo(output)
	} else {
		fmt.Fprintf(output, "AnonSurf %s\nSystem-wide Tor routing tool\n", version)
		fmt.Fprintln(output, "Parrot Security team")
		fmt.Fprintln(output, "GPL-3.0")
	}

	descriptions := make(map[string]string)
	for _, command := range root.Commands() {
		descriptions[command.Name()] = command.Short
	}

	fmt.Fprintf(output, "\n  %sUsage%s     anonsurf <command> [flags]\n\n", muted, reset)
	fmt.Fprintf(output, "  %sCOMMANDS%s\n", bold, reset)
	for _, entry := range rootHelpEntries {
		fmt.Fprintf(output, "  %s%-12s%s %s%-13s%s %s\n",
			blue, entry.name, reset,
			muted, entry.options, reset,
			descriptions[entry.name],
		)
	}
	fmt.Fprintf(output, "\n  %sActions%s   boot enable | disable | status\n", muted, reset)
	fmt.Fprintf(output, "  %sFlags%s     -k, --safe-kill   -n, --nyx\n", muted, reset)
	fmt.Fprintln(output, "            -h, --help        -v, --version")
	fmt.Fprintln(output, "            --verbose")
}

func writeLogo(output io.Writer) {
	title := [...]string{
		"",
		"",
		boldText + blueText + "ANONSURF" + resetText,
		mutedText + "System-wide Tor routing tool" + resetText,
		mutedText + "Version " + version + resetText,
		mutedText + "Parrot Security team" + resetText,
		mutedText + "GPL-3.0" + resetText,
	}
	for row := 0; row < len(logoPixels)/2; row++ {
		text := ""
		if row < len(title) {
			text = title[row]
		}
		fmt.Fprintf(output, "%s   %s\n", renderLogoRow(row), text)
	}
}

func renderLogoRow(row int) string {
	upper := logoPixels[row*2]
	lower := logoPixels[row*2+1]
	result := make([]byte, 0, len(upper)*12)
	for column := range upper {
		result = append(result, renderLogoCell(upper[column], lower[column])...)
	}
	return string(result)
}

func renderLogoCell(upper, lower byte) string {
	switch {
	case upper == ' ' && lower == ' ':
		return " "
	case upper == lower:
		return foreground(upper) + "\u2588" + resetText
	case lower == ' ':
		return foreground(upper) + "\u2580" + resetText
	case upper == ' ':
		return foreground(lower) + "\u2584" + resetText
	default:
		return foreground(upper) + background(lower) + "\u2580" + resetText
	}
}

func foreground(pixel byte) string {
	switch pixel {
	case 'W':
		return "\x1b[38;2;245;245;245m"
	case 'D':
		return darkBlueText
	default:
		return blueText
	}
}

func background(pixel byte) string {
	switch pixel {
	case 'W':
		return "\x1b[48;2;245;245;245m"
	case 'D':
		return "\x1b[48;2;0;94;133m"
	default:
		return "\x1b[48;2;0;124;166m"
	}
}
