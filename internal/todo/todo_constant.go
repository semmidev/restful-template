package todo

import "time"

// HTTP constants
const (
	// maxCoverSize is the upper bound for uploaded cover images.
	// previously io.ReadAll with no limit — an attacker could upload
	// an arbitrary-size payload to exhaust server memory.
	maxCoverSize = 5 << 20 // 5 MB
)

// Service constants
const (
	todoCacheTTL = 5 * time.Minute
)
