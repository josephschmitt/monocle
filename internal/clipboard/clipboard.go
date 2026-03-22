package clipboard

import (
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// Copy writes text to the system clipboard. It uses OSC 52 escape sequences
// (which work over SSH) and also tries platform-specific commands as fallback.
func Copy(text string) error {
	// OSC 52 works over SSH — the terminal emulator interprets the escape
	// sequence and sets the local clipboard. Write directly to the TTY to
	// bypass Bubble Tea's output handling.
	osc52Err := copyOSC52(text)

	// Also try system clipboard commands for terminals that don't support OSC 52.
	sysErr := copySystem(text)

	if osc52Err != nil && sysErr != nil {
		return sysErr
	}
	return nil
}

// copyOSC52 writes an OSC 52 escape sequence to the controlling terminal.
func copyOSC52(text string) error {
	encoded := base64.StdEncoding.EncodeToString([]byte(text))
	// Use BEL (\x07) as terminator for better tmux compatibility.
	sequence := fmt.Sprintf("\x1b]52;c;%s\x07", encoded)

	tty, err := openTTY()
	if err != nil {
		return err
	}
	defer tty.Close()

	_, err = tty.WriteString(sequence)
	return err
}

// openTTY opens the controlling terminal for writing.
func openTTY() (*os.File, error) {
	if runtime.GOOS == "windows" {
		return os.OpenFile("CON", os.O_WRONLY, 0)
	}
	return os.OpenFile("/dev/tty", os.O_WRONLY, 0)
}

// copySystem uses platform-specific clipboard commands.
func copySystem(text string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("pbcopy")
	case "linux":
		if _, err := exec.LookPath("xclip"); err == nil {
			cmd = exec.Command("xclip", "-selection", "clipboard")
		} else if _, err := exec.LookPath("xsel"); err == nil {
			cmd = exec.Command("xsel", "--clipboard", "--input")
		} else {
			return fmt.Errorf("no clipboard tool found (install xclip or xsel)")
		}
	case "windows":
		cmd = exec.Command("clip.exe")
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}
