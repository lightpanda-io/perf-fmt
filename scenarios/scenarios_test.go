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

package scenarios

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

const one = `{
  "env": {"node": "v26.13.0", "demo_commit": "abc"},
  "runs": [{"scenario": "campfire", "bin": "new", "round": 0, "instructions": 1030553162}],
  "summary": {
    "base": {"campfire": {"work_instructions": 1, "requests": 80}},
    "new": {
      "idle": {"work_instructions": 177316108, "instructions": 177316108},
      "campfire": {"work_instructions": 853237054, "instructions": 1030553162, "task_clock_ms": 201.49,
        "wall_ms": 202.68, "peak_rss_kb": 35844, "requests": 80}
    }
  }
}`

func TestAppend(t *testing.T) {
	ctx := context.Background()
	t1 := time.Date(2026, 11, 11, 10, 0, 0, 0, time.UTC)

	var first bytes.Buffer
	if err := (&Append{}).Append(ctx, "aaa", t1, &first, strings.NewReader(""), strings.NewReader(one)); err != nil {
		t.Fatal(err)
	}

	var second bytes.Buffer
	if err := (&Append{}).Append(ctx, "bbb", t1.Add(time.Hour), &second, &first, strings.NewReader(one)); err != nil {
		t.Fatal(err)
	}

	var res []OutResult
	if err := json.Unmarshal(second.Bytes(), &res); err != nil {
		t.Fatal(err)
	}
	if len(res) != 2 || res[0].Hash != "aaa" || res[1].Hash != "bbb" {
		t.Fatalf("unexpected history %+v", res)
	}
	c := res[1].Scenarios["campfire"]
	// The "new" binary's numbers, not the baseline's.
	if c.WorkInstructions != 853237054 || c.Requests != 80 || c.PeakRSSKB != 35844 {
		t.Fatalf("unexpected campfire %+v", c)
	}
	if res[1].Env["demo_commit"] != "abc" {
		t.Fatalf("unexpected env %+v", res[1].Env)
	}
}

func TestAppendDuplicate(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	var first bytes.Buffer
	if err := (&Append{}).Append(ctx, "aaa", now, &first, strings.NewReader(""), strings.NewReader(one)); err != nil {
		t.Fatal(err)
	}
	var second bytes.Buffer
	if err := (&Append{}).Append(ctx, "aaa", now, &second, &first, strings.NewReader(one)); err == nil {
		t.Fatal("expected hash exists")
	}
}

func TestAppendNoNewBin(t *testing.T) {
	in := `{"summary": {"base": {"campfire": {"work_instructions": 1}}}}`
	var out bytes.Buffer
	err := (&Append{}).Append(context.Background(), "aaa", time.Now(), &out, strings.NewReader(""), strings.NewReader(in))
	if err == nil {
		t.Fatal("expected an error without a new bin")
	}
}
