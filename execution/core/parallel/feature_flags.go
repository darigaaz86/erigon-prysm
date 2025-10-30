package parallel

import (
	"fmt"
	"os"
	"strconv"
	"sync/atomic"
)

// FeatureFlags manages feature flags for parallel execution.
type FeatureFlags struct {
	enabled                 atomic.Bool
	enableMetrics           atomic.Bool
	enablePanicRecovery     atomic.Bool
	enableAdaptiveThreshold atomic.Bool
	enableFallback          atomic.Bool
}

// NewFeatureFlags creates a new feature flags manager.
func NewFeatureFlags() *FeatureFlags {
	ff := &FeatureFlags{}
	ff.LoadFromEnvironment()
	return ff
}

// LoadFromEnvironment loads feature flags from environment variables.
func (ff *FeatureFlags) LoadFromEnvironment() {
	// Load PARALLEL_ENABLED
	if val := os.Getenv("PARALLEL_ENABLED"); val != "" {
		if enabled, err := strconv.ParseBool(val); err == nil {
			ff.enabled.Store(enabled)
		}
	}

	// Load PARALLEL_METRICS
	if val := os.Getenv("PARALLEL_METRICS"); val != "" {
		if enabled, err := strconv.ParseBool(val); err == nil {
			ff.enableMetrics.Store(enabled)
		}
	}

	// Load PARALLEL_PANIC_RECOVERY
	if val := os.Getenv("PARALLEL_PANIC_RECOVERY"); val != "" {
		if enabled, err := strconv.ParseBool(val); err == nil {
			ff.enablePanicRecovery.Store(enabled)
		}
	}

	// Load PARALLEL_ADAPTIVE_THRESHOLD
	if val := os.Getenv("PARALLEL_ADAPTIVE_THRESHOLD"); val != "" {
		if enabled, err := strconv.ParseBool(val); err == nil {
			ff.enableAdaptiveThreshold.Store(enabled)
		}
	}

	// Load PARALLEL_FALLBACK
	if val := os.Getenv("PARALLEL_FALLBACK"); val != "" {
		if enabled, err := strconv.ParseBool(val); err == nil {
			ff.enableFallback.Store(enabled)
		}
	}
}

// IsEnabled returns whether parallel execution is enabled.
func (ff *FeatureFlags) IsEnabled() bool {
	return ff.enabled.Load()
}

// SetEnabled sets whether parallel execution is enabled.
func (ff *FeatureFlags) SetEnabled(enabled bool) {
	ff.enabled.Store(enabled)
}

// IsMetricsEnabled returns whether metrics collection is enabled.
func (ff *FeatureFlags) IsMetricsEnabled() bool {
	return ff.enableMetrics.Load()
}

// SetMetricsEnabled sets whether metrics collection is enabled.
func (ff *FeatureFlags) SetMetricsEnabled(enabled bool) {
	ff.enableMetrics.Store(enabled)
}

// IsPanicRecoveryEnabled returns whether panic recovery is enabled.
func (ff *FeatureFlags) IsPanicRecoveryEnabled() bool {
	return ff.enablePanicRecovery.Load()
}

// SetPanicRecoveryEnabled sets whether panic recovery is enabled.
func (ff *FeatureFlags) SetPanicRecoveryEnabled(enabled bool) {
	ff.enablePanicRecovery.Store(enabled)
}

// IsAdaptiveThresholdEnabled returns whether adaptive threshold is enabled.
func (ff *FeatureFlags) IsAdaptiveThresholdEnabled() bool {
	return ff.enableAdaptiveThreshold.Load()
}

// SetAdaptiveThresholdEnabled sets whether adaptive threshold is enabled.
func (ff *FeatureFlags) SetAdaptiveThresholdEnabled(enabled bool) {
	ff.enableAdaptiveThreshold.Store(enabled)
}

// IsFallbackEnabled returns whether fallback is enabled.
func (ff *FeatureFlags) IsFallbackEnabled() bool {
	return ff.enableFallback.Load()
}

// SetFallbackEnabled sets whether fallback is enabled.
func (ff *FeatureFlags) SetFallbackEnabled(enabled bool) {
	ff.enableFallback.Store(enabled)
}

// String returns a string representation of feature flags.
func (ff *FeatureFlags) String() string {
	return fmt.Sprintf("FeatureFlags{Enabled: %v, Metrics: %v, PanicRecovery: %v, AdaptiveThreshold: %v, Fallback: %v}",
		ff.IsEnabled(), ff.IsMetricsEnabled(), ff.IsPanicRecoveryEnabled(),
		ff.IsAdaptiveThresholdEnabled(), ff.IsFallbackEnabled())
}

// ApplyToConfig applies feature flags to a configuration.
func (ff *FeatureFlags) ApplyToConfig(config *Config) {
	config.Enabled = ff.IsEnabled()
	config.EnableMetrics = ff.IsMetricsEnabled()
}

// GlobalFeatureFlags is the global feature flags instance.
var GlobalFeatureFlags = NewFeatureFlags()

// IsParallelEnabled returns whether parallel execution is globally enabled.
func IsParallelEnabled() bool {
	return GlobalFeatureFlags.IsEnabled()
}

// EnableParallel enables parallel execution globally.
func EnableParallel() {
	GlobalFeatureFlags.SetEnabled(true)
}

// DisableParallel disables parallel execution globally.
func DisableParallel() {
	GlobalFeatureFlags.SetEnabled(false)
}

// FeatureFlagManager manages feature flags with persistence.
type FeatureFlagManager struct {
	flags      *FeatureFlags
	configPath string
}

// NewFeatureFlagManager creates a new feature flag manager.
func NewFeatureFlagManager(configPath string) *FeatureFlagManager {
	return &FeatureFlagManager{
		flags:      NewFeatureFlags(),
		configPath: configPath,
	}
}

// Load loads feature flags from configuration file.
func (ffm *FeatureFlagManager) Load() error {
	// In a full implementation, this would load from file
	// For now, load from environment
	ffm.flags.LoadFromEnvironment()
	return nil
}

// Save saves feature flags to configuration file.
func (ffm *FeatureFlagManager) Save() error {
	// In a full implementation, this would save to file
	return nil
}

// GetFlags returns the feature flags.
func (ffm *FeatureFlagManager) GetFlags() *FeatureFlags {
	return ffm.flags
}

// RuntimeToggle allows toggling features at runtime.
type RuntimeToggle struct {
	manager *FeatureFlagManager
}

// NewRuntimeToggle creates a new runtime toggle.
func NewRuntimeToggle(manager *FeatureFlagManager) *RuntimeToggle {
	return &RuntimeToggle{
		manager: manager,
	}
}

// Toggle toggles a feature flag.
func (rt *RuntimeToggle) Toggle(feature string, enabled bool) error {
	flags := rt.manager.GetFlags()

	switch feature {
	case "parallel":
		flags.SetEnabled(enabled)
	case "metrics":
		flags.SetMetricsEnabled(enabled)
	case "panic_recovery":
		flags.SetPanicRecoveryEnabled(enabled)
	case "adaptive_threshold":
		flags.SetAdaptiveThresholdEnabled(enabled)
	case "fallback":
		flags.SetFallbackEnabled(enabled)
	default:
		return fmt.Errorf("unknown feature: %s", feature)
	}

	return rt.manager.Save()
}

// GetStatus returns the status of all features.
func (rt *RuntimeToggle) GetStatus() map[string]bool {
	flags := rt.manager.GetFlags()
	return map[string]bool{
		"parallel":           flags.IsEnabled(),
		"metrics":            flags.IsMetricsEnabled(),
		"panic_recovery":     flags.IsPanicRecoveryEnabled(),
		"adaptive_threshold": flags.IsAdaptiveThresholdEnabled(),
		"fallback":           flags.IsFallbackEnabled(),
	}
}

// FeatureGate controls access to experimental features.
type FeatureGate struct {
	gates map[string]bool
}

// NewFeatureGate creates a new feature gate.
func NewFeatureGate() *FeatureGate {
	return &FeatureGate{
		gates: make(map[string]bool),
	}
}

// Enable enables a feature gate.
func (fg *FeatureGate) Enable(feature string) {
	fg.gates[feature] = true
}

// Disable disables a feature gate.
func (fg *FeatureGate) Disable(feature string) {
	fg.gates[feature] = false
}

// IsEnabled checks if a feature gate is enabled.
func (fg *FeatureGate) IsEnabled(feature string) bool {
	return fg.gates[feature]
}

// List returns all feature gates and their status.
func (fg *FeatureGate) List() map[string]bool {
	result := make(map[string]bool)
	for k, v := range fg.gates {
		result[k] = v
	}
	return result
}

// ExperimentalFeatures defines experimental features.
var ExperimentalFeatures = []string{
	"adaptive_batching",
	"predictive_conflict_detection",
	"dynamic_worker_scaling",
	"cross_block_optimization",
}

// IsExperimental checks if a feature is experimental.
func IsExperimental(feature string) bool {
	for _, exp := range ExperimentalFeatures {
		if exp == feature {
			return true
		}
	}
	return false
}
