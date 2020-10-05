package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"path/filepath"

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
	c, err := resolvePath(*conf)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	d, err := resolvePath(*outdir)
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

func resolvePath(p string) (string, error) {
	if filepath.HasPrefix(p, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("Cannot resolve path %s: %v", p, err)
		}
		return path.Join(home, p[2:]), nil
	} else if !filepath.IsAbs(p) {
		wd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		return path.Join(wd, p), nil
	}
	return p, nil
}
