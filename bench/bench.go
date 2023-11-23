package bench

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"
)

type InResult struct {
	Name  string `json:"name"`
	Bench struct {
		Duration  int `json:"duration"`
		AllocSize int `json:"alloc_size"`
		AllocNb   int `json:"alloc_nb"`
		ReallocNb int `json:"realloc_nb"`
	} `json:"bench"`
}

type OutResult struct {
	Commit string     `json:"commit"`
	Time   time.Time  `json:"time"`
	Benchs []InResult `json:"benchs"`
}

type Append struct{}

func (a *Append) Append(ctx context.Context, out io.Writer, all io.Reader, one io.Reader) error {
	// decode one input
	var inr []InResult
	dec := json.NewDecoder(one)

	if err := dec.Decode(&inr); err != nil {
		return fmt.Errorf("decode one: %w", err)
	}

	// decode all input
	var outr []OutResult
	dec = json.NewDecoder(all)

	if err := dec.Decode(&outr); err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("decode all: %w", err)
	}

	outr = append(outr, OutResult{Commit: "TODO", Time: time.Now(), Benchs: inr})

	// encode output
	enc := json.NewEncoder(out)
	if err := enc.Encode(outr); err != nil {
		return fmt.Errorf("encode out: %w", err)
	}

	return nil
}
