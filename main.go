// Copyright 2023-2024 Lightpanda (Selecy SAS)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
	"time"

	browserbench "github.com/lightpanda-io/perf-fmt/bench/browser"
	jsrbench "github.com/lightpanda-io/perf-fmt/bench/jsruntime"
	"github.com/lightpanda-io/perf-fmt/cdp"
	"github.com/lightpanda-io/perf-fmt/git"
	"github.com/lightpanda-io/perf-fmt/hyperfine"
	"github.com/lightpanda-io/perf-fmt/s3"
	"github.com/lightpanda-io/perf-fmt/wpt"
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
	SourceBenchJSRuntime = "bench-jsruntime"
	SourceBenchBrowser   = "bench-browser"
	SourceWPT            = "wpt"
	SourceCDP            = "cdp"
	SourceHyperfine      = "hyperfine"

	AWSRegion = "eu-west-3"
	AWSBucket = "lpd-perf"

	PathBenchJSRuntime = "bench/jsruntime"
	PathBenchBrowser   = "bench/browser"
	PathCDP            = "cdp"
	PathWPT            = "wpt"
	PathHyperfine      = "hyperfine"
)

// run configures the flags and starts the HTTP API server.
func run(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	// declare runtime flag parameters.
	flags := flag.NewFlagSet(args[0], flag.ExitOnError)

	var (
		dev = flags.Bool("dev", false, "use dev/ dir storage prefix")
	)

	// usage func declaration.
	exec := args[0]
	flags.Usage = func() {
		fmt.Fprintf(stderr, "usage: %s <source> <commit> <result.json>\n", exec)
		fmt.Fprintf(stderr, "\nRead, format and save performance results.\n")
		fmt.Fprintf(stderr, "\nThe sources avalaible are:\n")
		fmt.Fprintf(stderr, "\t%s\tDEPRECATED jsruntime-lib benchmark json result.\n", SourceBenchJSRuntime)
		fmt.Fprintf(stderr, "\t%s\tlightpanda browser test benchmark json result.\n", SourceBenchBrowser)
		fmt.Fprintf(stderr, "\t%s\tlightpanda browser CDP benchmark json result.\n", SourceCDP)
		fmt.Fprintf(stderr, "\t%s\tlightpanda browser WPT test result.\n", SourceWPT)
		fmt.Fprintf(stderr, "\t%s\tlightpanda browser cold start.\n", SourceHyperfine)
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
	if len(args) != 3 {
		flags.Usage()
		return errors.New("bad arguments")
	}

	var (
		append Append
		path   string
	)

	switch args[0] {
	case SourceBenchJSRuntime:
		append = &jsrbench.Append{}
		path = PathBenchJSRuntime
	case SourceBenchBrowser:
		append = &browserbench.Append{}
		path = PathBenchBrowser
	case SourceCDP:
		append = &cdp.Append{}
		path = PathCDP
	case SourceWPT:
		append = &wpt.Append{}
		path = PathWPT
	case SourceHyperfine:
		append = &hyperfine.Append{}
		path = PathHyperfine
	default:
		flags.Usage()
		return errors.New("bad source")
	}

	// If dev flag is active, use the `dev/` dir prefix.
	if *dev {
		path = "dev/" + path
		fmt.Fprintf(os.Stderr, "⚠️  Dev mode enabled, result will be stored in %q\n", path)
	}

	hash := git.CommitHash(args[1])
	// TODO check commit format

	now := time.Now().UTC()

	// open one
	one, err := os.Open(args[2])
	if err != nil {
		return fmt.Errorf("open input file: %w", err)
	}
	defer one.Close()

	// prepare S3 connection
	// set default env region if not already set.
	if _, ok := os.LookupEnv("AWS_REGION"); !ok {
		os.Setenv("AWS_REGION", AWSRegion)
	}
	fio, err := s3.NewS3IO(env("AWS_BUCKET", AWSBucket), path+"/history.json")
	if err != nil {
		return fmt.Errorf("new s3 io: %w", err)
	}

	// pull the all
	all, err := fio.Pull(ctx)
	if err != nil {
		return fmt.Errorf("pull all file: %w", err)
	}
	defer all.Close()

	var out bytes.Buffer

	// append input to output
	if err := append.Append(ctx, hash, now, &out, all, one); err != nil {
		return fmt.Errorf("append result: %w", err)
	}

	// push output
	if err := fio.Push(ctx, &out); err != nil {
		return fmt.Errorf("push result: %w", err)
	}

	// push the single result file
	// Reset the file handler to the begining of the file
	if _, err := one.Seek(0, 0); err != nil {
		return fmt.Errorf("reset file: %w", err)
	}

	filename := fmt.Sprintf("%s_%v.json", now.Format("2006-01-02_15-04"), hash)
	fio, err = s3.NewS3IO(env("AWS_BUCKET", AWSBucket), path+"/"+filename)
	if err != nil {
		return fmt.Errorf("news3io single result: %w", err)
	}

	// push output
	if err := fio.Push(ctx, one); err != nil {
		return fmt.Errorf("push single result : %w", err)
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
