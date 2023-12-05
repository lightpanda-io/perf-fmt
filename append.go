package main

import (
	"context"
	"io"
	"time"

	"github.com/browsercore/perf-fmt/git"
)

type Append interface {
	Append(ctx context.Context,
		hash git.CommitHash, datetime time.Time,
		out io.Writer,
		all io.Reader, one io.Reader,
	) error
}
