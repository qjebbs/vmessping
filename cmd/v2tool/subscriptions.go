package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/qjebbs/v2tool/files"
	"github.com/qjebbs/v2tool/vmess"
)

func subscriptionsCmd(args []string) {
	subsCmd := flag.NewFlagSet("v2tool subscriptions", flag.ExitOnError)
	conf := subsCmd.String("c", "", "subscriptions config file")
	outdir := subsCmd.String("o", ".", "output dir")
	socketMark := subsCmd.Int("m", 0, "SO_MARK for outbounds")
	err := subsCmd.Parse(args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if *conf == "" {
		subsCmd.Usage()
		os.Exit(1)
	}
	c, err := files.ResolvePath(*conf)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	d, err := files.ResolvePath(*outdir)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	err = vmess.FetchSubscriptions(c, d, int32(*socketMark))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
