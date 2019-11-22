package main

import (
	"context"
	"errors"
//	"path/filepath"
	"fmt"
	"io"
	"os"
	"time"
//	"os/signal"
//	"runtime"
//	"strings"
	"syscall"
	"sync"

	"github.com/urfave/cli"
	specs "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/kata-containers/agent/protocols/client"
	"github.com/kata-containers/agent/protocols/grpc"
	"github.com/kata-containers/runtime/virtcontainers/pkg/oci"
	"github.com/kata-containers/runtime/apitest/terminal"
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
	var execp	*specs.Process
	var grpcp	*grpc.Process

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

	if ociSpec.Spec.Process == nil {
		fmt.Println("no Process configuration?")
		if ociSpec.Process != nil {
			fmt.Println("Process configuration in compactOCIPROCESS")
			ociSpec.Spec.Process = &ociSpec.Process.Process
			execp = &ociSpec.Process.Process
		}
	}

	execp = ociSpec.Spec.Process
//	execp.Args = []string {"testbin"}

	if err != nil {
		fmt.Println("Cannot parse config.json")
		return errors.New("Invalid config.json")
	}

//	execp.Args = []string {"echo", "600s"}

	grpcp, err = grpc.ProcessOCItoGRPC(execp)
	if err != nil {
		fmt.Println("Can not convert OCIProcsess to grpc Process")
		return errors.New("Cannont convert Process")
	}

	time.Sleep(2 * time.Second)

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

	wg := &sync.WaitGroup {}
	in := &InStream {
		ctx: ctx,
		cid: id,
		eid: eid,
		cli: cli,
	}

	out := &OutStream {
		ctx: ctx,
		cid: id,
		eid: eid,
		cli: cli,
	}

	errStream := &ErrStream {
		ctx: ctx,
		cid: id,
		eid: eid,
		cli: cli,
	}

	// setup termios beforeio.Copy
	termios, err := terminal.SetupTerminal(int(os.Stdin.Fd()))
	defer terminal.RestoreTerminal(int(os.Stdin.Fd()), termios)

	wg.Add(1)
	go func() {
		// can we read from stdin?
		/*
		buf := make([]byte, 100)
		l, err := os.Stdin.Read(buf)
		if err != nil {
			fmt.Println("Read stdin error")
			return
		}
		fmt.Println("Copy stdin routine, read", l, "byte:", buf[:l])
		/*
		buf1 := make([]byte, l)
		copy(buf1, buf)
		fmt.Println("buf1:", string(buf1))
		buf1 = []byte("asd\n")
		*/
		/*
		buf1 := []byte("asdsdf\n")
		// write ls to server
		stdinreq := &grpc.WriteStreamRequest {
			ContainerId: id,
			ExecId: id,
			Data: buf1, //[]byte("asdsds\n"),
		}

		stdinresp, err := in.cli.WriteStdin(ctx, stdinreq)

		if err !=nil {
			fmt.Println("write stdin error!")
			return
		}

		fmt.Println("go write ", stdinresp.Len, "bytes")
		*/
		_, err := io.Copy(in, os.Stdin)
		if err != nil {
			fmt.Println("write stdin failed")
		}

		closereq := &grpc.CloseStdinRequest {
			ContainerId: id,
			ExecId: eid,
		}

		_, err1 := in.cli.CloseStdin(ctx, closereq)
		if err1 != nil {
			fmt.Println("close stdin failed")
		}
		wg.Done()
	}()

	go func() {
		_, _ = io.Copy(os.Stdout, out)
		wg.Done()
	}()
	_ = errStream

	go func() {
		_, _ = io.Copy(os.Stderr, errStream)
	}()

/*
	time.Sleep(600 * time.Second)


	// signal process
	sigreq := &grpc.SignalProcessRequest {
		ContainerId: id,
		ExecId: eid,
		Signal: uint32(syscall.SIGTERM),
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

	_ = syscall.SIGKILL
	_ = grpcp
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
	/*
	_, err = cli.Check(ctx, &grpc.CheckRequest{})
	if err != nil {
		fmt.Println("check agent status fail")
	}

	// get guest details
	detailreq := &grpc.GuestDetailsRequest {
		MemBlockSize: true,
		MemHotplugProbe: true,
	}
	detailresp, err := cli.GetGuestDetails(ctx, detailreq)
	if err != nil {
		fmt.Println("Can not get guest details")
	} else  {
		fmt.Println("Agent version: %v", detailresp.AgentDetails.Version)
	}
	*/

	wg.Wait()

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
