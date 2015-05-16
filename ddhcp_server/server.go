package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/elasti-rans/ddhcp/ddhcp"
)

type CliIp net.IP

func (c CliIp) String() string {
	return "ip"
}

func (c CliIp) Set(s string) error {
	if parsedIp := net.ParseIP(s); parsedIp == nil {
		return errors.New(fmt.Sprintf("failed to parsed %s into op object", s))
	} else {
		c = CliIp(parsedIp)
		return nil
	}
}

type serverArgs struct {
	Ip     CliIp
	Subnet CliIp

	PoolStartIp CliIp
	PoolEndIp   CliIp
}

func cliParse() serverArgs {
	var args serverArgs
	flag.Var(&args.Ip, "ip", "ip to listen for request")
	flag.Var(&args.Subnet, "subnet", "subnet to listen")
	flag.Var(&args.PoolStartIp, "pool-start", "first ip in the dhcp pool")
	flag.Var(&args.PoolEndIp, "pool-end", "last ip in the dhcp pool")

	flag.Parse()
	return args
}

func main() {
	args := cliParse()

	pool, err := ddhcp.NewLeasePool(net.IP(args.PoolStartIp), net.IP(args.PoolEndIp))
	if err != nil {
		log.Fatal(err)
	}

	server, err := ddhcp.New(net.IP(args.Ip), pool)
	if err != nil {
		log.Fatal(err)
	}

	defer server.Close()
	server.Serve()
	// TODO: need to add signal handler, to terminate the server
}
