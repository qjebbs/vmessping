package main

import (
	"fmt"
	"os"
)

var (
	MAINVER = "0.0.0-src"
)

const usage = `
v2tool is a helper for using v2ray.

Usage:

        v2tool <command> [arguments]

The commands are:

        ping        ping a vmess link / json outbound config file (vmessping)
        outbound    add / remove outbounds through v2ray api server

Use "v2tool help <command>" for more information about a command.
`

func main() {
	if len(os.Args) < 2 {
		usageAndExit(0)
		return
	}
	showHelp := false
	cmd := os.Args[1]
	if cmd == "help" {
		showHelp = true
		if len(os.Args) < 3 {
			usageAndExit(1)
			return
		}
		cmd = os.Args[2]
	}
	args := []string{"-h"}
	if !showHelp {
		args = os.Args[2:]
	}
	switch cmd {
	case "ping":
		ping(args)
	case "outbound":
		outbound(args)
	default:
		usageAndExit(1)
	}
}

func usageAndExit(code int) {
	fmt.Println(usage)
	os.Exit(code)
}
