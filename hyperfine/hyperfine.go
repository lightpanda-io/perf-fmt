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

package hyperfine

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

type InResult struct {
	Results []struct {
		Mean float64 `json:"mean"`
		Min  float64 `json:"min"`
		Max  float64 `json:"max"`
	} `json:"results"`
}

type OutResult struct {
	Hash git.CommitHash `json:"commit"`
	Time time.Time      `json:"datetime"`
	Mean float64        `json:"mean"`
	Min  float64        `json:"min"`
	Max  float64        `json:"max"`
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

	if len(inr.Results) != 1 {
		return errors.New("unexpected results size")
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

	outres := OutResult{
		Hash: hash,
		Time: datetime,
		Mean: inr.Results[0].Mean,
		Min:  inr.Results[0].Min,
		Max:  inr.Results[0].Max,
	}

	allres = append(allres, outres)

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
