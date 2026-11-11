// Copyright 2023-2026 Lightpanda (Selecy SAS)
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

// Package scenarios formats the output of the demo repo's perf harness
// (perf/, `go run . -bin base=...,new=... -out perf.json`).
//
// The harness compares two binaries. The history keeps the medians of the
// one named "new", the commit being tested.
package scenarios

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"time"

	"github.com/lightpanda-io/perf-fmt/git"
)

// Bin is the binary of the harness output that is recorded.
const Bin = "new"

type Summary struct {
	// Median instructions minus the idle scenario's: startup and the CDP
	// connection excluded. The number to plot.
	WorkInstructions int64   `json:"work_instructions"`
	Instructions     int64   `json:"instructions"`
	TaskClockMs      float64 `json:"task_clock_ms"`
	WallMs           float64 `json:"wall_ms"`
	PeakRSSKB        int64   `json:"peak_rss_kb"`
	Requests         int64   `json:"requests"`
}

type InResult struct {
	Env     map[string]string             `json:"env"`
	Summary map[string]map[string]Summary `json:"summary"`
}

type OutResult struct {
	Hash git.CommitHash `json:"commit"`
	Time time.Time      `json:"datetime"`
	// What else the numbers depend on (demo commit, node, puppeteer, kernel,
	// cpu): a change here explains a jump that isn't the browser's.
	Env       map[string]string  `json:"env"`
	Scenarios map[string]Summary `json:"scenarios"`
}

type Append struct{}

func (a *Append) Append(
	ctx context.Context,
	hash git.CommitHash, datetime time.Time,
	out io.Writer,
	all io.Reader, one io.Reader,
) error {
	// decode one input
	var inr InResult
	dec := json.NewDecoder(one)

	if err := dec.Decode(&inr); err != nil {
		return fmt.Errorf("decode one: %w", err)
	}

	scenarios, ok := inr.Summary[Bin]
	if !ok || len(scenarios) == 0 {
		return fmt.Errorf("no results for bin %q", Bin)
	}

	// decode all input
	var allres []OutResult
	dec = json.NewDecoder(all)

	if err := dec.Decode(&allres); err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("decode all: %w", err)
	}

	// search if the commit already exists in the all results to avoid duplication.
	for _, v := range allres {
		if hash == v.Hash {
			return errors.New("hash exists")
		}
	}

	allres = append(allres, OutResult{
		Hash:      hash,
		Time:      datetime,
		Env:       inr.Env,
		Scenarios: scenarios,
	})

	// reorder slice
	sort.Slice(allres, func(i, j int) bool {
		return allres[i].Time.Before(allres[j].Time)
	})

	// encode output
	enc := json.NewEncoder(out)
	if err := enc.Encode(allres); err != nil {
		return fmt.Errorf("encode out: %w", err)
	}

	return nil
}
