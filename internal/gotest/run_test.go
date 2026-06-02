package gotest

import "testing"

func feed(t *testing.T, r *Run, lines ...string) {
	t.Helper()
	for _, line := range lines {
		ev, ok := Decode(line)
		if !ok {
			t.Fatalf("Decode(%q) = not ok", line)
		}
		r.Add(ev)
	}
}

func TestRunAggregates(t *testing.T) {
	r := NewRun()
	feed(t, r,
		`{"Action":"run","Package":"x","Test":"TestA"}`,
		`{"Action":"pass","Package":"x","Test":"TestA","Elapsed":0.01}`,
		`{"Action":"run","Package":"x","Test":"TestB"}`,
		`{"Action":"output","Package":"x","Test":"TestB","Output":"    x_test.go:5: boom\n"}`,
		`{"Action":"fail","Package":"x","Test":"TestB","Elapsed":0.02}`,
		`{"Action":"fail","Package":"x","Elapsed":0.03}`,
	)

	if r.Pass != 1 || r.Fail != 1 || r.Skip != 0 {
		t.Fatalf("counts: pass=%d fail=%d skip=%d", r.Pass, r.Fail, r.Skip)
	}
	if !r.Failed() {
		t.Fatal("Failed() = false, want true")
	}

	f := r.Failures()
	if len(f) != 1 {
		t.Fatalf("Failures() len = %d, want 1 (no duplicate package-level entry)", len(f))
	}
	if f[0].Test != "TestB" || f[0].Kind != Assertion {
		t.Fatalf("failure = %+v, want TestB/Assertion", f[0])
	}
}

func TestSlowestOrder(t *testing.T) {
	r := NewRun()
	feed(t, r,
		`{"Action":"pass","Package":"x","Test":"Fast","Elapsed":0.10}`,
		`{"Action":"pass","Package":"x","Test":"Slow","Elapsed":0.90}`,
	)
	s := r.Slowest(5)
	if len(s) != 2 || s[0].Name != "Slow" {
		t.Fatalf("Slowest = %v, want [Slow Fast]", s)
	}
}

func TestDecodeSkipsNonJSON(t *testing.T) {
	if _, ok := Decode("# example.com/x [build failed]"); ok {
		t.Fatal("Decode(non-JSON) = ok, want skipped")
	}
	if _, ok := Decode(""); ok {
		t.Fatal("Decode(empty) = ok, want skipped")
	}
}

func TestClassify(t *testing.T) {
	cases := []struct {
		out  []string
		want FailKind
	}{
		{[]string{"panic: boom\n", "goroutine 1\n"}, Panic},
		{[]string{"==================\n", "WARNING: DATA RACE\n"}, Race},
		{[]string{"    x_test.go:5: not equal\n"}, Assertion},
	}
	for _, c := range cases {
		if got := classify(c.out); got != c.want {
			t.Errorf("classify(%v) = %v, want %v", c.out, got, c.want)
		}
	}
}
