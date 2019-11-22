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
//	oci "github.com/kata-containers/runtime/virtcontainers/pkg/compatoci"
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
//	var execp	*specs.Process
//	var grpcp	*grpc.Process

	ctx := context.Background()
	cli, err := client.NewAgentClient(ctx,
			"hvsock:///tmp/root/kata.hvsock:1024", false)

	if err != nil {
		fmt.Println("open connection fails")
		return errors.New("Open connection fails")
	}

	defer cli.Close()

/*
	bdir := clic.String("bundle")
	id := clic.Args().First()

	ociSpec, err := oci.ParseConfigJSON(bdir)

	if ociSpec.Spec.Process == nil {
		fmt.Println("no Process configuration?")
		if ociSpec.Process != nil {
			fmt.Println("Process configuration in compactOCIPROCESS")
			ociSpec.Spec.Process = &ociSpec.Process.Process
			execp = &ociSpec.Process.Process
			caps, _ := ociSpec.Process.Capabilities.(types.LinuxCapabilities)
			ociSpec.Spec.Process.Capabilities = &specs.LinuxCapabilities {
			Bounding: caps.Bounding,
			Effective: caps.Effective,
			Inheritable: caps.Inheritable,
			Permitted: caps.Permitted,
			Ambient: caps.Ambient,
			}
		}
	}

	fmt.Println("uid mappings:", ociSpec.Linux.UIDMappings)
	fmt.Println("gid mappings:", ociSpec.Linux.GIDMappings)
	fmt.Println("ociSpec: ", ociSpec)
	fmt.Println("Process: ", ociSpec.Spec.Process)
	fmt.Println("caps: ", ociSpec.Spec.Process.Capabilities)
	execp = ociSpec.Spec.Process
//	execp.Args = []string {"testbin"}

	if err != nil {
		fmt.Println("Cannot parse config.json")
		return errors.New("Invalid config.json")
	}

	grpcSpec, err := grpc.OCItoGRPC(&ociSpec.Spec)

	if err != nil {
		fmt.Println("Cannot convert ociSpec to grpcSpec")
		return errors.New("Cannot convert ociSpec to grpcSpec")
	}

	fmt.Println("uid mappings:", grpcSpec.Linux.UIDMappings)
	fmt.Println("gid mappings:", grpcSpec.Linux.GIDMappings)

	// grpcSpec.Linux.UIDMappings[0].Size_ = 500
	// grpcSpec.Linux.GIDMappings[0].Size_ =1000
	grpcSpec.Root.Path = filepath.Join(bdir, grpcSpec.Root.Path)

	execp.Args = []string {"echo", "600s"}

	grpcp, err = grpc.ProcessOCItoGRPC(execp)
	if err != nil {
		fmt.Println("Can not convert OCIProcsess to grpc Process")
		return errors.New("Cannont convert Process")
	}
*/

/*
	var (
	devices []*grpc.Device
	storages []*grpc.Storage
	)
	req := &grpc.CreateContainerRequest {
		ContainerId: id,
		ExecId: id,
		Devices: devices,
		Storages: storages,
		OCI: grpcSpec,
		SandboxPidns: false,
	}

	_, err = cli.CreateContainer(ctx, req)

	if err != nil {
		fmt.Printf("Create container %s failed\n", id)
		return errors.New("Connot create container")
	}

	// to start the container
	var req1 *grpc.StartContainerRequest
	req1 = &grpc.StartContainerRequest {
		ContainerId: id,
	}

	time.Sleep(1 * time.Second)

	_, err = cli.StartContainer(ctx, req1)

	if err != nil {
		fmt.Println("Cannot start container %v", id)
		return errors.New("Cannot start container")
	}

	fmt.Println("started container %v", id)

	time.Sleep(2 * time.Second)

	// handle io
	wg := &sync.WaitGroup {}
	in := &InStream {
		ctx: ctx,
		cid: id,
		eid: id,
		cli: cli,
	}

	out := &OutStream {
		ctx: ctx,
		cid: id,
		eid: id,
		cli: cli,
	}

	errStream := &ErrStream {
		ctx: ctx,
		cid: id,
		eid: id,
		cli: cli,
	}

	// setup terminal befor io.copy
	termios, err := terminal.SetupTerminal(int(os.Stdin.Fd()))
	defer terminal.RestoreTerminal(int(os.Stdin.Fd()), termios)

	wg.Add(1)
	go func() {
		// can we read from stdin?
		_, err := io.Copy(in, os.Stdin)
		if err != nil {
			fmt.Println("write stdin failed")
		}
		wg.Done()
	}()

	go func() {
		_, _ = io.Copy(os.Stdout, out)
		wg.Done()
	}()
	_ = errStream
*/
/*
	go func() {
		_, _ = io.Copy(os.Stderr, errStream)
	}()
*/

/*
	// exec process
	eid := id + "exec"
	execreq := &grpc.ExecProcessRequest {
		ContainerId: id,
		ExecId: eid,
		Process: grpcp,
	}

	_, err = cli.ExecProcess(ctx, execreq)

	if err != nil {
		fmt.Println("Exec Process failes!")
	}

	time.Sleep(1 * time.Second)

	// signal process
	sigreq := &grpc.SignalProcessRequest {
		ContainerId: id,
		ExecId: eid,
		Signal: uint32(syscall.SIGKILL),
	}

	_, err = cli.SignalProcess(ctx, sigreq)
	if err != nil {
		fmt.Println("Can not signal process")
	}

	// wait process
	waitreq := &grpc.WaitProcessRequest {
		ContainerId: id,
		ExecId: eid,
	}

	waitresp, err := cli.WaitProcess(ctx, waitreq)
	if err != nil {
		fmt.Println("Wait process fail")
	} else {
		fmt.Println("exit status: %v", waitresp.Status)
	}
*/
	// check
	fmt.Println("before call check!")
	_, err = cli.HealthClient.Check(ctx, &grpc.CheckRequest{})
	if err != nil {
		fmt.Println("check agent status fail")
	}

	fmt.Println("agent alive!")

	// get guest details
	detailreq := &grpc.GuestDetailsRequest {
		MemBlockSize: true,
		MemHotplugProbe: true,
	}
	detailresp, err := cli.AgentServiceClient.GetGuestDetails(ctx, detailreq)
	if err != nil {
		fmt.Println("Can not get guest details")
	} else  {
		fmt.Println("Agent version: %v", detailresp.AgentDetails.Version)
	}


	return nil
}

