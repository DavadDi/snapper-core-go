package main

import (
	"flag"

	"github.com/teambition/snapper-core-go/rpc"
)

var rule = flag.String("rule", "rpc", "rpc or core")
var config = flag.String("config", "../config/default.json", "rpc or core")

func main() {
	rpc := new(rpc.RPC)
	flag.Parse()
	if *rule == "rpc" {
		rpc.Start(*config)
	} else if *rule == "core" {

	}
}
