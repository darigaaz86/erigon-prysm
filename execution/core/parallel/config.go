package parallel

import "errors"

// Error constants
var (
	ErrInvalidWorkerCount       = errors.New("invalid worker count: must be >= 0")
	ErrInvalidConflictThreshold = errors.New("invalid conflict threshold: must be between 0.0 and 1.0")
	ErrInvalidMaxReexecutions   = errors.New("invalid max reexecutions: must be >= 1")
)

// Config holds configuration parameters for parallel execution.
type Config struct {
	// Enabled controls whether parallel execution is active
	Enabled bool

	// WorkerCount specifies the number of worker threads
	// If 0, defaults to runtime.NumCPU() - 1
	WorkerCount int

	// ConflictThreshold is the conflict rate threshold for fallback to sequential execution
	// Value between 0.0 and 1.0 (default: 0.5 = 50%)
	ConflictThreshold float64

	// MaxReexecutions is the maximum number of re-execution attempts per transaction
	// Default: 3
	MaxReexecutions int

	// EnableMetrics controls whether detailed metrics are collected
	EnableMetrics bool
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		Enabled:           false, // Disabled by default for safety
		WorkerCount:       0,     // Auto-detect
		ConflictThreshold: 0.5,   // 50%
		MaxReexecutions:   3,
		EnableMetrics:     true,
	}
}

// Validate checks if the configuration is valid.
func (c *Config) Validate() error {
	if c.WorkerCount < 0 {
		return ErrInvalidWorkerCount
	}
	if c.ConflictThreshold < 0.0 || c.ConflictThreshold > 1.0 {
		return ErrInvalidConflictThreshold
	}
	if c.MaxReexecutions < 1 {
		return ErrInvalidMaxReexecutions
	}
	return nil
}
