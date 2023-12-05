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
		FreeNb    int `json:"free_nb"`
	} `json:"bench"`
}

type OutItem struct {
	Duration  int `json:"duration"`
	AllocSize int `json:"allocationSize"`
	AllocNb   int `json:"alloccation"`
	ReallocNb int `json:"reallocation"`
	FreeNb    int `json:"free"`
}

type OutResult struct {
	Commit string    `json:"commitHash"`
	Time   time.Time `json:"dateTime"`
	Data   struct {
		WithIsolate    OutItem `json:"withIsolate"`
		WithoutIsolate OutItem `json:"withoutIsolate"`
	} `json:"data"`
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
	var allres []OutResult
	dec = json.NewDecoder(all)

	if err := dec.Decode(&allres); err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("decode all: %w", err)
	}

	outres := OutResult{
		Commit: "TODO",
		Time:   time.Now(),
	}

	for _, v := range inr {
		switch v.Name {
		case "With Isolate":
			outres.Data.WithIsolate = OutItem(v.Bench)
		case "Without Isolate":
			outres.Data.WithoutIsolate = OutItem(v.Bench)
		default:
			return fmt.Errorf("unhandled bench result: %s", v.Name)
		}
	}

	allres = append(allres, outres)

	// encode output
	enc := json.NewEncoder(out)
	if err := enc.Encode(allres); err != nil {
		return fmt.Errorf("encode out: %w", err)
	}

	return nil
}
