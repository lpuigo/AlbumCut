package duration

import "testing"

func TestFormat(t *testing.T) {
	cases := []struct {
		seconds int
		want    string
	}{
		{0, "00:00:00"},
		{59, "00:00:59"},
		{60, "00:01:00"},
		{3661, "01:01:01"},
	}
	for _, c := range cases {
		if got := Format(c.seconds); got != c.want {
			t.Errorf("Format(%d) = %q, want %q", c.seconds, got, c.want)
		}
	}
}
