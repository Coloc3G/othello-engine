package learning

// Constants for coefficient ranges - keep these
const (
	MaterialMin  = 1
	MaterialMax  = 5000
	MobilityMin  = 1
	MobilityMax  = 5000
	CornersMin   = 1
	CornersMax   = 5000
	ParityMin    = 1
	ParityMax    = 5000
	StabilityMin = 1
	StabilityMax = 5000
	FrontierMin  = 1
	FrontierMax  = 5000
)

// New improved mutation parameters
const (
	// Small random mutations most of the time
	SmallMutationRate = 0.25
	SmallDeltaMax     = 25 // Small adjustments

	// Medium mutations occasionally
	MediumMutationRate = 0.05
	MediumDeltaMax     = 75 // Medium adjustments

	// Large mutations rarely (for exploration)
	LargeMutationRate = 0.02
	LargeDeltaMax     = 200 // Large adjustments for exploration

	// Completely new value generation (very rare)
	RerollRate = 0.01
)
