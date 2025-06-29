package learning

// Constants for coefficient ranges - keep these
const (
	MaterialMin  = 1
	MaterialMax  = 100
	MobilityMin  = 1
	MobilityMax  = 100
	CornersMin   = 1
	CornersMax   = 100
	ParityMin    = 1
	ParityMax    = 100
	StabilityMin = 1
	StabilityMax = 100
	FrontierMin  = 1
	FrontierMax  = 100
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

// Hybrid training parameters
const (
	// Seuil minimum de positions pour utiliser le GPU
	DefaultGPUThreshold = 1000

	// Batch GPU après chaque génération
	DefaultBatchFrequency = 1

	// Collecte des positions activée par défaut
	DefaultPositionCollection = true
)
