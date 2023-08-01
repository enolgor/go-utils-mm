package textutils

import "testing"

func TestNormalize(t *testing.T) {
	type testCase struct {
		input   string
		options []NormalizeOption
		output  string
	}
	cases := []testCase{
		{"sömeÑ rAndom ç string", []NormalizeOption{}, "somenrandomcstring"},
		{"sömeÑ rAndom ç string", []NormalizeOption{OptionSpace}, "somen random c string"},
		{"sömeÑ rAndom ç string", []NormalizeOption{OptionSlugSpace}, "somen-random-c-string"},
		{"sömeÑ rAndom ç string//++**", []NormalizeOption{OptionSpace, OptionSlugSpace}, "somen-random-c-string"},
	}
	for _, test := range cases {
		got, err := Normalize(test.input, test.options...)
		if err != nil {
			t.Error(err)
		}
		if got != test.output {
			t.Errorf("got %q, wanted %q", got, test.output)
		}
	}
}
