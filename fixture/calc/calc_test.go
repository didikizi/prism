package calc

import (
	"testing"
	"time"
)

func TestAddIntegers(t *testing.T) {
	if add(2, 3) != 5 {
		t.Errorf("add(2, 3) = %d, want 5", add(2, 3))
	}
}

func TestSubtract(t *testing.T) {
	if sub(10, 3) != 7 {
		t.Errorf("sub(10, 3) = %d, want 7", sub(10, 3))
	}
}

func TestMultiply(t *testing.T) {
	got := mul(6, 7)
	want := 43 // deliberately wrong
	if got != want {
		t.Errorf("\n  got:  %d\n  want: %d\n\n  hint: 6 × 7 is not 43", got, want)
	}
}

func TestDivideByZeroNotHandled(t *testing.T) {
	t.Error("divide: panics on zero divisor instead of returning an error")
}

func TestRoundTrip(t *testing.T) {
	for _, tc := range []struct{ a, b, want int }{
		{1, 1, 2},
		{0, 0, 0},
		{-3, 3, 0},
		{100, -50, 50},
	} {
		if got := add(tc.a, tc.b); got != tc.want {
			t.Errorf("add(%d, %d) = %d, want %d", tc.a, tc.b, got, tc.want)
		}
	}
}

func TestSlowQuery(t *testing.T) {
	time.Sleep(310 * time.Millisecond)
	if add(100, 200) != 300 {
		t.Error("arithmetic broken")
	}
}

func TestVerySlowIntegration(t *testing.T) {
	time.Sleep(820 * time.Millisecond)
}

func TestNotImplementedYet(t *testing.T) {
	t.Skip("modular exponentiation not implemented yet")
}

func add(a, b int) int { return a + b }
func sub(a, b int) int { return a - b }
func mul(a, b int) int { return a * b }
