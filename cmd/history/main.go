package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"syscall"

	"github.com/lightpanda-io/perf-fmt/bench"
	jsrbench "github.com/lightpanda-io/perf-fmt/bench/jsruntime"
	"github.com/lightpanda-io/perf-fmt/s3"
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

	AWSRegion = "eu-west-3"
	AWSBucket = "lpd-perf"
)

// run configures the flags and starts the HTTP API server.
func run(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	// declare runtime flag parameters.
	flags := flag.NewFlagSet(args[0], flag.ExitOnError)
	// usage func declaration.
	exec := args[0]
	flags.Usage = func() {
		fmt.Fprintf(stderr, "usage: %s <dir>\n", exec)
		fmt.Fprintf(stderr, "\nRead, format and save perf bench results history.\n")
		fmt.Fprintf(stderr, "\nTo upload data in AWS S3, the program uses env var:\n")
		fmt.Fprintf(stderr, "\tAWS_ACCESS_KEY_ID\t\trequired\n")
		fmt.Fprintf(stderr, "\tAWS_SECRET_ACCESS_KEY\t\trequired\n")
		fmt.Fprintf(stderr, "\tAWS_REGION\t\t\tdefault value: %s\n", AWSRegion)
		fmt.Fprintf(stderr, "\tAWS_BUCKET\t\t\tdefault value: %s\n", AWSBucket)
	}
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}

	args = flags.Args()
	if len(args) != 1 {
		flags.Usage()
		return errors.New("bad arguments")
	}

	dirname := args[0]
	files, err := os.ReadDir(dirname)
	if err != nil {
		return fmt.Errorf("opendir: %w", err)
	}

	var res []jsrbench.OutResult
	for _, file := range files {
		fmt.Fprintln(os.Stderr, file.Name())
		// ignore subdirs
		if file.IsDir() {
			continue
		}

		b, err := os.ReadFile(filepath.Join(dirname, file.Name()))
		if err != nil {
			return fmt.Errorf("readfile: %w", err)
		}

		out, err := jsrbench.ParseTxtData(file.Name(), b)
		if err != nil {
			return fmt.Errorf("parse file: %w", err)
		}

		res = append(res, out)
	}

	// pull history.json file
	fio, err := s3.NewS3IO(env("AWS_BUCKET", AWSBucket), "bench/history.json")
	if err != nil {
		return fmt.Errorf("news3io single result: %w", err)
	}

	// pull the whole data
	ball, err := fio.Pull(ctx)
	if err != nil {
		return fmt.Errorf("pull all file: %w", err)
	}
	defer ball.Close()

	// decode the result
	dec := json.NewDecoder(ball)
	var all []jsrbench.OutResult
	if err := dec.Decode(&all); err != nil {
		return fmt.Errorf("decode all history: %w", err)
	}

	// let merge all and res
APPEND:
	for _, in := range res {
		for i, a := range all {
			if in.Hash == a.Hash {
				// the commit already exists in all we replace the values.
				all[i] = in
				continue APPEND
			}
		}

		all = append(all, in)
	}

	// reorder all slice
	sort.Slice(all, func(i, j int) bool {
		return all[i].Time.Before(all[j].Time)
	})

	// push the history.json
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	if err := enc.Encode(all); err != nil {
		return fmt.Errorf("json encode: %w", err)
	}

	// push output
	if err := fio.Push(ctx, &buf); err != nil {
		return fmt.Errorf("push single result : %w", err)
	}

	for _, v := range res {
		in := []bench.InResult{
			{Name: "With Isolate", Bench: bench.InItem(v.Data.WithIsolate)},
			{Name: "Without Isolate", Bench: bench.InItem(v.Data.WithoutIsolate)},
		}

		bin, err := json.Marshal(in)
		if err != nil {
			return fmt.Errorf("json encode single result: %w", err)
		}

		// upload all individual file to s3
		filename := fmt.Sprintf("%s_%v.json", v.Time.Format("2006-01-02_15-04"), v.Hash)
		fio, err := s3.NewS3IO(env("AWS_BUCKET", AWSBucket), "bench/"+filename)
		if err != nil {
			return fmt.Errorf("news3io single result: %w", err)
		}

		// push output
		if err := fio.Push(ctx, bytes.NewReader(bin)); err != nil {
			return fmt.Errorf("push single result : %w", err)
		}
	}

	return nil
}

func env(key, dflt string) string {
	val, ok := os.LookupEnv(key)
	if !ok {
		return dflt
	}

	return val
}
