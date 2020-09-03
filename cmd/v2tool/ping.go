package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/v2fly/vmessping"
)

func ping(args []string) {
	pingCmd := flag.NewFlagSet("v2tool ping", flag.ExitOnError)
	verbose := pingCmd.Bool("v", false, "verbose (debug log)")
	showNode := pingCmd.Bool("n", false, "show node location/outbound ip")
	usemux := pingCmd.Bool("m", false, "use mux outbound")
	desturl := pingCmd.String("dest", "http://www.google.com/gen_204", "the test destination url, need 204 for success return")
	count := pingCmd.Uint("c", 9999, "Count. Stop after sending COUNT requests")
	timeout := pingCmd.Uint("o", 10, "timeout seconds for each request")
	inteval := pingCmd.Uint("i", 1, "inteval seconds between pings")
	quit := pingCmd.Uint("q", 0, "fast quit on error counts")
	pingCmd.Parse(args)

	var vmess string
	if pingCmd.NArg() == 0 {
		if vmess = os.Getenv("VMESS"); vmess == "" {
			fmt.Println("To ping a vmess link:")
			fmt.Println(os.Args[0], "vmess://....")
			fmt.Println()
			fmt.Println("To ping a config file:")
			fmt.Println(os.Args[0], "path/to/config.json")
			fmt.Println()
			pingCmd.Usage()
			os.Exit(1)
		}
	} else {
		vmess = pingCmd.Args()[0]
	}

	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, os.Interrupt, os.Kill, syscall.SIGTERM)

	vmessping.PrintVersion(MAINVER)
	ps, err := vmessping.Ping(vmess, *count, *desturl, *timeout, *inteval, *quit, osSignals, *showNode, *verbose, *usemux)
	if err != nil {
		os.Exit(1)
	}
	ps.PrintStats()
	if ps.IsErr() {
		os.Exit(1)
	}
}
