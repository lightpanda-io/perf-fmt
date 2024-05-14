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

package jsruntime

import "testing"

func TestParseFilename(t *testing.T) {
	for _, tc := range []string{
		"2022-11-07_22-58_0883be2_main.txt", "2022-11-07_13-29_0dd2b63_optional_arg.txt",
	} {
		name := tc
		t.Run(tc, func(t *testing.T) {
			dt, c, err := parseTxtName(name)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			t.Log(dt, c)
		})
	}
}

func TestParseLine(t *testing.T) {
	for _, tc := range []string{
		"  | Without Isolateªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªª |              178usªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªª  |            2821ªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªª  |     48kbªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªª  |",
		"  | With Isolateªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªª    |              736usªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªª  |            2908ªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªª  |     54kbªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªªª  |",
		"  | With Isolate    |              850us  |               3  |               1084  |     72kb  |",
		"  | Without Isolate |              328us  |               2  |                977  |     24kb  |",
	} {
		data := []byte(tc)
		t.Run("", func(t *testing.T) {
			v, err := parseLine(data)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			t.Log(v)
		})
	}
}
