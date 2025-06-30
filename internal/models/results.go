package models

import "math"

// ComparisonResult holds the final comparison results
type ComparisonResult struct {
	IsSameVehicle   bool            `json:"is_same_vehicle"`
	SimilarityScore float64         `json:"similarity_score"`
	ConfidenceLevel ConfidenceLevel `json:"confidence_level"`
	DetailedScores  DetailedScores  `json:"detailed_scores"`
	ProcessingInfo  ProcessingInfo  `json:"processing_info"`
}

type ConfidenceLevel int
const (
	ConfidenceHigh ConfidenceLevel = iota
	ConfidenceMedium
	ConfidenceLow
)

// DetailedScores breaks down similarity by feature type
type DetailedScores struct {
	GeometricSimilarity    float64 `json:"geometric_similarity"`
	LightPatternSimilarity float64 `json:"light_pattern_similarity"`
	BumperSimilarity       float64 `json:"bumper_similarity"`
	ColorSimilarity        float64 `json:"color_similarity,omitempty"`
	ThermalSimilarity      float64 `json:"thermal_similarity,omitempty"`
}

// ProcessingInfo holds processing metadata
type ProcessingInfo struct {
	ProcessingTimeMs    int64   `json:"processing_time_ms"`
	Image1Quality       float64 `json:"image1_quality"`
	Image2Quality       float64 `json:"image2_quality"`
	AlignmentQuality    float64 `json:"alignment_quality"`
	ViewConsistency     bool    `json:"view_consistency"`
	LightingConsistency bool    `json:"lighting_consistency"`
}

// ValidateAndSanitize ensures all float values in the result are valid for JSON marshaling
func (cr *ComparisonResult) ValidateAndSanitize() {
	cr.SimilarityScore = sanitizeFloat64(cr.SimilarityScore, 0.0)
	
	cr.DetailedScores.GeometricSimilarity = sanitizeFloat64(cr.DetailedScores.GeometricSimilarity, 0.0)
	cr.DetailedScores.LightPatternSimilarity = sanitizeFloat64(cr.DetailedScores.LightPatternSimilarity, 0.0)
	cr.DetailedScores.BumperSimilarity = sanitizeFloat64(cr.DetailedScores.BumperSimilarity, 0.0)
	cr.DetailedScores.ColorSimilarity = sanitizeFloat64(cr.DetailedScores.ColorSimilarity, 0.0)
	cr.DetailedScores.ThermalSimilarity = sanitizeFloat64(cr.DetailedScores.ThermalSimilarity, 0.0)
	
	cr.ProcessingInfo.Image1Quality = sanitizeFloat64(cr.ProcessingInfo.Image1Quality, 0.0)
	cr.ProcessingInfo.Image2Quality = sanitizeFloat64(cr.ProcessingInfo.Image2Quality, 0.0)
	cr.ProcessingInfo.AlignmentQuality = sanitizeFloat64(cr.ProcessingInfo.AlignmentQuality, 0.0)
}

func sanitizeFloat64(value float64, defaultValue float64) float64 {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return defaultValue
	}
	return value
}