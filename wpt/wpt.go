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

package wpt

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
	Pass  bool   `json:"pass"`
	Crash bool   `json:"crash"`
	Name  string `json:"name"`
	Cases []struct {
		Pass    bool   `json:"pass"`
		Name    string `json:"name"`
		Message string `json:"message"`
	} `json:"cases"`
}

type OutResult struct {
	Hash git.CommitHash `json:"commit"`
	Time time.Time      `json:"datetime"`
	Data struct {
		Pass  int `json:"pass"`
		Fail  int `json:"fail"`
		Crash int `json:"crash"`
	} `json:"data"`
}

type Append struct{}

func (a *Append) Append(
	ctx context.Context,
	hash git.CommitHash, datetime time.Time,
	out io.Writer,
	all io.Reader, one io.Reader,
) error {
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

	// search if the commit already exists in the all results to avoid duplication.
	for _, v := range allres {
		if hash == v.Hash {
			return errors.New("hash exists")
		}
	}

	outres := OutResult{
		Hash: hash,
		Time: datetime,
	}

	for _, t := range inr {
		if t.Crash {
			outres.Data.Crash += 1
			continue
		}

		for _, tc := range t.Cases {
			if tc.Pass {
				outres.Data.Pass += 1
			} else {
				outres.Data.Fail += 1
			}
		}
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
