package bench

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
