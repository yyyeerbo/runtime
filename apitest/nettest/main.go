package main

import (
	"context"
	"errors"
//	"path/filepath"
	"fmt"
//	"io"
	"os"
//	"time"
//	"os/signal"
//	"runtime"
	"strings"
	_ "syscall"
//	"sync"

	"github.com/urfave/cli"
//	specs "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/kata-containers/agent/protocols/client"
	"github.com/kata-containers/agent/protocols/grpc"
	"github.com/kata-containers/agent/pkg/types"
//	"github.com/kata-containers/runtime/virtcontainers/pkg/oci"
//	"github.com/kata-containers/runtime/apitest/terminal"
)

// create client with NewAgentClient
// use ParseConfigJson to extract ocispec from config.json
// call CreateContainer and StartContainer to test

func main() {
	app := cli.NewApp()

	app.Flags = []cli.Flag {
		cli.StringFlag {
			Name: "bundle, b",
			Value: "",
			Usage: "path to bundle directory, default to cwd",
		},
	}

	app.Action = RunTest

	app.Run(os.Args)
}

func RunTest(clic *cli.Context) error {

	ctx := context.Background()
	cli, err := client.NewAgentClient(ctx,
			"unix:///tmp/testagent:1", false)

	if err != nil {
		fmt.Println("open connection fails")
		return errors.New("Open connection fails")
	}

	defer cli.Close()

	ifres, err := cli.ListInterfaces(ctx, &grpc.ListInterfacesRequest {})

	if err != nil {
		fmt.Println("list interfaces failed!")
	}

	fmt.Println(ifres)

	rtres, err := cli.ListRoutes(ctx, &grpc.ListRoutesRequest {})

	if err != nil {
		fmt.Println("list routes failed!")
	}

	fmt.Println(rtres)

	// update interface
	var iface *types.Interface

	for _, iface = range ifres.Interfaces {
		if strings.HasPrefix(iface.Name, "test") {
			fmt.Println(iface)
			break
		}
	}

	if iface.Mtu == 1500 {
		iface.Mtu = 1460
	} else {
		iface.Mtu = 1500
	}

	for _, addr := range iface.IPAddresses {
		if strings.HasPrefix(addr.Address, "192.168.42") {
			if addr.Address == "192.168.42.203" {
				addr.Address = "192.168.42.229"
			} else {
				addr.Address = "192.168.42.203"
			}
		}
	}

	upifreq := &grpc.UpdateInterfaceRequest {
		Interface: iface,
	}

	retif, _ := cli.UpdateInterface(ctx, upifreq)

	fmt.Println(retif)

	// update routes
	var uprts []*types.Route
	for i, rt := range rtres.Routes {
		if i == 1 {
			continue;
		}
		uprts = append(uprts, rt)
	}

//	uprts[0].Dest = ""

	fmt.Println(uprts)


	uprtreq := &grpc.UpdateRoutesRequest {
		Routes: &grpc.Routes {
			Routes: uprts,
		},
	}

	retrt, _ := cli.UpdateRoutes(ctx, uprtreq)

	fmt.Println(retrt)

	return nil
}

type InStream struct {
	ctx		context.Context
	cid		string
	eid		string
	cli		*client.AgentClient
}

type OutStream struct {
	ctx		context.Context
	cid		string
	eid		string
	cli		*client.AgentClient
}

type ErrStream struct {
	ctx		context.Context
	cid		string
	eid		string
	cli		*client.AgentClient
}

func (in *InStream) Write(data []byte) (n int, err error) {
	req := &grpc.WriteStreamRequest {
		ContainerId: in.cid,
		ExecId: in.eid,
		Data: data,
	}

	resp, err := in.cli.WriteStdin(in.ctx, req)

	if err !=nil {
		return 0, err
	}

	return int(resp.Len), nil
}

func (out *OutStream) Read(data []byte) (n int, err error) {
	req := &grpc.ReadStreamRequest {
		ContainerId: out.cid,
		ExecId:	out.eid,
		Len: uint32(len(data)),
	}

	resp, err := out.cli.ReadStdout(out.ctx, req)

	if err != nil {
		return 0, nil
	}

	copy(data, resp.Data)
	return len(resp.Data), nil
}

func (errStream *ErrStream) Read(data []byte) (n int, err error) {
	req := &grpc.ReadStreamRequest {
		ContainerId: errStream.cid,
		ExecId:	errStream.eid,
		Len: uint32(len(data)),
	}

	resp,err := errStream.cli.ReadStderr(errStream.ctx, req)

	if err != nil {
		return 0, nil
	}

	copy(data, resp.Data)
	return len(resp.Data), nil
}
