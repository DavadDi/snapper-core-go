package main

import (
	"flag"

	"github.com/teambition/snapper-core-go/src/app"
	"github.com/teambition/snapper-core-go/src/rpc"
)

var rule = flag.String("rule", "rpc", "rpc or core")
var config = flag.String("config", "././config/default.json", "rpc or core")

func main() {
	flag.Parse()
	if *rule == "rpc" {
		rpc := new(rpc.RPC)
		rpc.Start(*config)
	} else if *rule == "core" {
		app.Start(*config)
	}
}
