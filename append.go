package main

import (
	"context"
	"io"

	"github.com/browsercore/perf-fmt/git"
)

type Append interface {
	Append(ctx context.Context,
		hash git.CommitHash,
		out io.Writer,
		all io.Reader,
		one io.Reader,
	) error
}
