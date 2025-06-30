package models

import (
	"gocv.io/x/gocv"
)

// VehicleView represents the view type of the vehicle
type VehicleView int

const (
	ViewFront VehicleView = iota
	ViewRear
	ViewUnknown
)

// LightingType represents lighting conditions
type LightingType int

const (
	LightingDaylight LightingType = iota
	LightingInfrared
	LightingUnknown
)

// VehicleImage holds image data and metadata
type VehicleImage struct {
	Image          gocv.Mat            `json:"-"`
	View           VehicleView         `json:"view"`
	Lighting       LightingType        `json:"lighting"`
	QualityScore   float64             `json:"quality_score"`
	ProcessingMeta ProcessingMetadata  `json:"processing_meta"`
}

// ProcessingMetadata holds processing information
type ProcessingMetadata struct {
	OriginalWidth    int    `json:"original_width"`
	OriginalHeight   int    `json:"original_height"`
	VehicleBounds    Bounds `json:"vehicle_bounds"`
	NormalizedWidth  int    `json:"normalized_width"`
	NormalizedHeight int    `json:"normalized_height"`
}

// Bounds represents a bounding rectangle
type Bounds struct {
	X, Y, Width, Height int
}

// Point2D represents a 2D coordinate
type Point2D struct {
	X, Y float64
}

// LicensePlateRegion represents a detected license plate area
type LicensePlateRegion struct {
	Bounds        Bounds  `json:"bounds"`
	Confidence    float64 `json:"confidence"`
	AvgBrightness float64 `json:"avg_brightness"`
	IsReflective  bool    `json:"is_reflective"`
}

// IRSignature represents the infrared signature around a license plate
type IRSignature struct {
	PlateRegion          LicensePlateRegion `json:"plate_region"`
	SurroundingRegion    Bounds            `json:"surrounding_region"`
	ReflectivityMap      [][]float64       `json:"reflectivity_map"`
	MaterialSignature    []float64         `json:"material_signature"`
	IlluminationGradient []float64         `json:"illumination_gradient"`
	ShadowPatterns       []Point2D         `json:"shadow_patterns"`
	TextureFeatures      []float64         `json:"texture_features"`
}