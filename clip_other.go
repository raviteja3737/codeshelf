//go:build !windows

package main

import (
	"errors"
	"os/exec"
	"runtime"
	"strings"
)

func copyToClipboard(text string) error {
	var cmds [][]string
	if runtime.GOOS == "darwin" {
		cmds = [][]string{{"pbcopy"}}
	} else {
		cmds = [][]string{{"wl-copy"}, {"xclip", "-selection", "clipboard"}, {"xsel", "-b", "-i"}}
	}
	for _, c := range cmds {
		cmd := exec.Command(c[0], c[1:]...)
		cmd.Stdin = strings.NewReader(text)
		if cmd.Run() == nil {
			return nil
		}
	}
	return errors.New("no clipboard tool (Linux: install xclip or wl-clipboard)")
}

func enableVT() func() { return func() {} }
