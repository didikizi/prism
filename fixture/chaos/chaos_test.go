package chaos

import "testing"

// TestUnstableOperation demonstrates prism's panic detection.
// The raw panic kills the test binary; prism renders it as a PANIC card.
func TestUnstableOperation(t *testing.T) {
	var s []string
	_ = s[0] // index out of range → panic → test binary crash
}
