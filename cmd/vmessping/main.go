package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	vmessping "github.com/qjebbs/v2tool/vmessping"
)

var (
	MAINVER = "0.0.0-src"
)

func main() {
	verbose := flag.Bool("v", false, "verbose (debug log)")
	showNode := flag.Bool("n", false, "show node location/outbound ip")
	usemux := flag.Bool("m", false, "use mux outbound")
	desturl := flag.String("dest", "http://www.google.com/gen_204", "the test destination url, need 204 for success return")
	count := flag.Uint("c", 9999, "Count. Stop after sending COUNT requests")
	timeout := flag.Uint("o", 10, "timeout seconds for each request")
	inteval := flag.Uint("i", 1, "inteval seconds between pings")
	quit := flag.Uint("q", 0, "fast quit on error counts")
	flag.Parse()

	var vmess string
	if flag.NArg() == 0 {
		if vmess = os.Getenv("VMESS"); vmess == "" {
			fmt.Println("To ping a vmess link:")
			fmt.Println(os.Args[0], "vmess://....")
			fmt.Println()
			fmt.Println("To ping a config file:")
			fmt.Println(os.Args[0], "path/to/config.json")
			fmt.Println()
			flag.Usage()
			os.Exit(1)
		}
	} else {
		vmess = flag.Args()[0]
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
