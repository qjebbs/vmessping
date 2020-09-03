package main

import (
	"flag"
	"os"
	"strings"
)

type stringArrayFlags []string

func (i *stringArrayFlags) String() string {
	return strings.Join(*i, ",")
}

func (i *stringArrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}
func outbound(args []string) {
	cmd := ""
	if len(args) > 0 {
		cmd = args[0]
	}
	switch cmd {
	case "add":
		outboundAdd(args[1:])
	case "remove":
		outboundRemove(args[1:])
	default:
		args := []string{"-h"}
		outboundRemove(args)
		outboundAdd(args)
	}
}

func outboundRemove(args []string) {
	cmd := flag.NewFlagSet("v2tool outbound remove", flag.ContinueOnError)
	var tags stringArrayFlags
	var files stringArrayFlags
	host := cmd.String("s", "127.0.0.1", "api server address")
	port := cmd.Uint("p", 10085, "api server port")
	cmd.Var(&tags, "t", "the tags of outbounds to remove")
	cmd.Var(&files, "f", "json file to remove (by the tag of the file)")
	err := cmd.Parse(args)
	if err != nil {
		return
	}
	server := APIServer{
		Host: *host,
		Port: uint16(*port),
	}
	defer server.Close()
	if len(files) > 0 {
		err = server.RemoveOutboundFiles(files)
	} else {
		err = server.RemoveOutbounds(tags)
	}
	if err != nil {
		os.Stderr.WriteString(err.Error())
		os.Exit(1)
	}
}
func outboundAdd(args []string) {
	cmd := flag.NewFlagSet("v2tool outbound add", flag.ContinueOnError)
	var jsons stringArrayFlags
	host := cmd.String("s", "127.0.0.1", "api server address")
	port := cmd.Uint("p", 10085, "api server port")
	cmd.Var(&jsons, "f", "json file to add")
	err := cmd.Parse(args)
	if err != nil {
		return
	}
	server := APIServer{
		Host: *host,
		Port: uint16(*port),
	}
	defer server.Close()
	err = server.AddOutboundFiles(jsons)
	if err != nil {
		os.Stderr.WriteString(err.Error())
		os.Exit(1)
	}
}
