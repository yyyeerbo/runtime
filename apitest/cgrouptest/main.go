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
//	"strings"
	_ "syscall"
//	"sync"

	"github.com/urfave/cli"
//	specs "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/kata-containers/agent/protocols/client"
	"github.com/kata-containers/agent/protocols/grpc"
	"github.com/kata-containers/runtime/virtcontainers/pkg/oci"
//	"github.com/kata-containers/runtime/apitest/terminal"
//	"github.com/kata-containers/runtime/virtcontainers/types"
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

	bdir := clic.String("bundle")
	id := clic.Args().First()

	ociSpec, err := oci.ParseConfigJSON(bdir)

	if err != nil {
		fmt.Println("Cannot parse config.json")
		return errors.New("Invalid config.json")
	}

	ociRes := ociSpec.Linux.Resources
	grpcRes,err := grpc.ResourcesOCItoGRPC(ociRes)

	if err != nil {
		fmt.Println("Cannot convert ociResources to grpcResources")
		return errors.New("Convert error")
	}

	if err != nil {
		fmt.Println("Cannot convert ociSpec to grpcSpec")
		return errors.New("Cannot convert ociSpec to grpcSpec")
	}

	// UpdateContainer
	upReq := &grpc.UpdateContainerRequest {
		ContainerId: id,
		Resources: grpcRes,
	}

	_, err = cli.UpdateContainer(ctx, upReq)

	if err != nil {
		fmt.Println("UpdateContainer failed!")
		return errors.New("UpdateContainer fail")
	}

	fmt.Println("UpdateContainer succeeded!")

	// StatsContainer
	stReq := &grpc.StatsContainerRequest {
		ContainerId: id,
	}

	stRes, err := cli.StatsContainer(ctx, stReq)
	if err != nil {
		fmt.Println("StatsContainer failed!")
		return errors.New("StatsContainer fail")
	}

	fmt.Println(stRes)

	// ListProcesses
	listReq := &grpc.ListProcessesRequest {
		ContainerId: id,
		Format: "table",
		Args: []string {},
	}

	listRes, err := cli.ListProcesses(ctx, listReq)
	if err != nil {
		fmt.Println("ListProcesses failed!")
		return errors.New("ListProcesses fail")
	}

	fmt.Println(listRes)

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
