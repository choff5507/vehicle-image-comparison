package models

// VehicleFeatures holds all extracted features for a vehicle image
type VehicleFeatures struct {
	View              VehicleView         `json:"view"`
	Lighting          LightingType        `json:"lighting"`
	
	// Universal features (work in all lighting)
	GeometricFeatures GeometricFeatures   `json:"geometric_features"`
	
	// View-specific features
	LightPatterns     LightPatternFeatures `json:"light_patterns"`
	BumperFeatures    BumperFeatures      `json:"bumper_features"`
	
	// Lighting-optimized features
	DaylightFeatures  *DaylightFeatures   `json:"daylight_features,omitempty"`
	InfraredFeatures  *InfraredFeatures   `json:"infrared_features,omitempty"`
	
	ExtractionQuality float64             `json:"extraction_quality"`
}

// GeometricFeatures - work in all lighting conditions
type GeometricFeatures struct {
	VehicleProportions VehicleProportions `json:"vehicle_proportions"`
	StructuralElements []StructuralElement `json:"structural_elements"`
	ReferencePoints    []Point2D          `json:"reference_points"`
}

// VehicleProportions holds dimensional ratios
type VehicleProportions struct {
	WidthHeightRatio  float64 `json:"width_height_ratio"`
	UpperLowerRatio   float64 `json:"upper_lower_ratio"`
	LicensePlateRatio float64 `json:"license_plate_ratio"`
}

// StructuralElement represents a structural component
type StructuralElement struct {
	Type     string  `json:"type"`
	Position Point2D `json:"position"`
	Size     float64 `json:"size"`
}

// LightPatternFeatures for headlights/taillights
type LightPatternFeatures struct {
	LightElements      []LightElement     `json:"light_elements"`
	PatternSignature   []float64          `json:"pattern_signature"`
	LightConfiguration LightConfiguration `json:"light_configuration"`
}

// LightElement represents individual light components
type LightElement struct {
	Position  Point2D    `json:"position"`
	Shape     LightShape `json:"shape"`
	Size      float64    `json:"size"`
	Intensity float64    `json:"intensity"`
	Type      LightType  `json:"type"`
}

type LightShape int
const (
	ShapeRectangular LightShape = iota
	ShapeRound
	ShapeAngular
	ShapeCustom
)

type LightType int
const (
	TypeHeadlight LightType = iota
	TypeTaillight
	TypeDRL
	TypeFogLight
	TypeBrakeLight
)

// LightConfiguration represents the overall light setup
type LightConfiguration struct {
	NumElements int     `json:"num_elements"`
	Symmetry    float64 `json:"symmetry"`
	Spacing     float64 `json:"spacing"`
}

// BumperFeatures for bumper analysis
type BumperFeatures struct {
	ContourSignature []Point2D `json:"contour_signature"`
	TextureFeatures  []float64 `json:"texture_features"`
	MountingPoints   []Point2D `json:"mounting_points"`
	LicensePlateArea Bounds    `json:"license_plate_area"`
}

// DaylightFeatures - only available in daylight
type DaylightFeatures struct {
	ColorProfile    ColorProfile     `json:"color_profile"`
	BadgeLocations  []BadgeFeature   `json:"badge_locations"`
	TrimDetails     []TrimFeature    `json:"trim_details"`
	SurfaceTexture  TextureSignature `json:"surface_texture"`
}

// InfraredFeatures - only available in infrared
type InfraredFeatures struct {
	ThermalSignature   []float64          `json:"thermal_signature"`
	ReflectiveElements []ReflectiveElement `json:"reflective_elements"`
	HeatPatterns       []HeatPattern      `json:"heat_patterns"`
	MaterialSignature  []float64          `json:"material_signature"`
	IRSignature        *IRSignature       `json:"ir_signature,omitempty"`
}

// Supporting types for features
type ColorProfile struct {
	DominantColors []Color `json:"dominant_colors"`
	Histogram      []int   `json:"histogram"`
}

type Color struct {
	R, G, B uint8   `json:"r,g,b"`
	Weight  float64 `json:"weight"`
}

type BadgeFeature struct {
	Position Point2D `json:"position"`
	Size     float64 `json:"size"`
	Shape    string  `json:"shape"`
}

type TrimFeature struct {
	Position Point2D `json:"position"`
	Type     string  `json:"type"`
	Texture  string  `json:"texture"`
}

type TextureSignature struct {
	Features []float64 `json:"features"`
	Type     string    `json:"type"`
}

type ReflectiveElement struct {
	Position   Point2D `json:"position"`
	Intensity  float64 `json:"intensity"`
	Size       float64 `json:"size"`
	Shape      string  `json:"shape"`
}

type HeatPattern struct {
	Region      Bounds    `json:"region"`
	Temperature float64   `json:"temperature"`
	Gradient    []float64 `json:"gradient"`
}