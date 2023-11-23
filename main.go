package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/browsercore/perf-fmt/bench"
)

const (
	exitOK   = 0
	exitFail = 1
)

// main starts interruptable context and runs the program.
func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	err := run(ctx, os.Args, os.Stdout, os.Stderr)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(exitFail)
	}

	os.Exit(exitOK)
}

const (
	SourceBench = "bench"
	SourceWPT   = "wpt"
)

// run configures the flags and starts the HTTP API server.
func run(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	// declare runtime flag parameters.
	flags := flag.NewFlagSet(args[0], flag.ExitOnError)
	// usage func declaration.
	flags.Usage = func() {
		fmt.Fprintf(stderr, "usage: %s <source> <result.json>\n", args[0])
		fmt.Fprintf(stderr, "\nRead, format and save performance results.\n")
		fmt.Fprintf(stderr, "\nThe sources avalaible are:\n")
		fmt.Fprintf(stderr, "\t%s\tjsruntime-lib benchmark json result.\n", SourceBench)
		fmt.Fprintf(stderr, "\t%s\tbrowsercore WPT test result.\n", SourceWPT)
	}
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}

	args = flags.Args()
	if len(args) != 2 {
		flags.Usage()
		return errors.New("bad arguments")
	}

	var append Append

	switch args[0] {
	case SourceBench:
		append = &bench.Append{}
	case SourceWPT:
		return errors.New("not implemented source")
	default:
		flags.Usage()
		return errors.New("bad source")
	}

	// open one
	one, err := os.Open(args[1])
	if err != nil {
		return fmt.Errorf("open input file: %w", err)
	}
	defer one.Close()

	// pull the all
	fio := FileIO{Path: "/tmp/append.json"}
	all, err := fio.Pull(ctx)
	if err != nil {
		return fmt.Errorf("pull all file: %w", err)
	}
	defer all.Close()

	var out bytes.Buffer

	// append input to output
	if err := append.Append(ctx, &out, all, one); err != nil {
		return fmt.Errorf("append result: %w", err)
	}

	// push output
	if err := fio.Push(ctx, &out); err != nil {
		return fmt.Errorf("push result: %w", err)
	}

	return nil
}
