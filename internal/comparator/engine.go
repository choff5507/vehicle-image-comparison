package comparator

import (
	"github.com/choff5507/vehicle-image-comparison/internal/models"
	"fmt"
	"math"
)

// safeFloat64 ensures a float64 value is valid (not NaN or Inf) and within bounds
func safeFloat64(value float64, defaultValue float64) float64 {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return defaultValue
	}
	return math.Max(0.0, math.Min(1.0, value))
}

type ComparisonEngine struct {
	geometricWeight    float64
	lightPatternWeight float64
	bumperWeight       float64
	colorWeight        float64
	thermalWeight      float64
}

func NewComparisonEngine() *ComparisonEngine {
	return &ComparisonEngine{
		geometricWeight:    0.35,
		lightPatternWeight: 0.30,
		bumperWeight:       0.20,
		colorWeight:        0.10,
		thermalWeight:      0.05,
	}
}

// CompareVehicles performs comprehensive vehicle comparison
func (ce *ComparisonEngine) CompareVehicles(features1, features2 models.VehicleFeatures) (*models.ComparisonResult, error) {
	// Validate that views and lighting are consistent
	if features1.View != features2.View {
		return nil, fmt.Errorf("cannot compare different vehicle views")
	}
	
	if features1.Lighting != features2.Lighting {
		return nil, fmt.Errorf("cannot compare different lighting conditions")
	}
	
	// Calculate individual similarity scores
	detailedScores := models.DetailedScores{
		GeometricSimilarity:    ce.compareGeometricFeatures(features1.GeometricFeatures, features2.GeometricFeatures),
		LightPatternSimilarity: ce.compareLightPatterns(features1.LightPatterns, features2.LightPatterns),
		BumperSimilarity:       ce.compareBumperFeatures(features1.BumperFeatures, features2.BumperFeatures),
	}
	
	// Add lighting-specific comparisons
	if features1.DaylightFeatures != nil && features2.DaylightFeatures != nil {
		detailedScores.ColorSimilarity = ce.compareDaylightFeatures(*features1.DaylightFeatures, *features2.DaylightFeatures)
	}
	
	if features1.InfraredFeatures != nil && features2.InfraredFeatures != nil {
		detailedScores.ThermalSimilarity = ce.compareInfraredFeatures(*features1.InfraredFeatures, *features2.InfraredFeatures)
	}
	
	// Calculate weighted overall similarity
	overallSimilarity := ce.calculateWeightedSimilarity(detailedScores, features1.Lighting)
	
	// Determine if same vehicle
	isSameVehicle := overallSimilarity > ce.getSimilarityThreshold(features1.Lighting)
	
	// Calculate confidence level
	confidenceLevel := ce.calculateConfidenceLevel(overallSimilarity, features1, features2)
	
	return &models.ComparisonResult{
		IsSameVehicle:   isSameVehicle,
		SimilarityScore: overallSimilarity,
		ConfidenceLevel: confidenceLevel,
		DetailedScores:  detailedScores,
	}, nil
}

func (ce *ComparisonEngine) compareGeometricFeatures(geo1, geo2 models.GeometricFeatures) float64 {
	// Compare vehicle proportions
	proportionSimilarity := ce.compareVehicleProportions(geo1.VehicleProportions, geo2.VehicleProportions)
	
	// Compare structural elements
	structuralSimilarity := ce.compareStructuralElements(geo1.StructuralElements, geo2.StructuralElements)
	
	// Compare reference points alignment
	alignmentSimilarity := ce.compareReferencePoints(geo1.ReferencePoints, geo2.ReferencePoints)
	
	// Weighted combination
	result := (proportionSimilarity*0.4 + structuralSimilarity*0.4 + alignmentSimilarity*0.2)
	return safeFloat64(result, 0.5)
}

func (ce *ComparisonEngine) compareVehicleProportions(prop1, prop2 models.VehicleProportions) float64 {
	// Compare width/height ratio
	widthHeightSim := 0.5 // Default similarity
	if maxRatio := math.Max(prop1.WidthHeightRatio, prop2.WidthHeightRatio); maxRatio > 0 {
		widthHeightSim = 1.0 - math.Abs(prop1.WidthHeightRatio-prop2.WidthHeightRatio)/maxRatio
	}
	
	// Compare upper/lower ratio
	upperLowerSim := 0.5 // Default similarity
	if maxRatio := math.Max(prop1.UpperLowerRatio, prop2.UpperLowerRatio); maxRatio > 0 {
		upperLowerSim = 1.0 - math.Abs(prop1.UpperLowerRatio-prop2.UpperLowerRatio)/maxRatio
	}
	
	// Compare license plate ratio (if available)
	licensePlateSim := 1.0
	if prop1.LicensePlateRatio > 0 && prop2.LicensePlateRatio > 0 {
		if maxRatio := math.Max(prop1.LicensePlateRatio, prop2.LicensePlateRatio); maxRatio > 0 {
			licensePlateSim = 1.0 - math.Abs(prop1.LicensePlateRatio-prop2.LicensePlateRatio)/maxRatio
		}
	}
	
	result := (widthHeightSim*0.4 + upperLowerSim*0.4 + licensePlateSim*0.2)
	
	// Ensure result is valid
	if math.IsNaN(result) || math.IsInf(result, 0) {
		return 0.5
	}
	return math.Max(0.0, math.Min(1.0, result))
}

func (ce *ComparisonEngine) compareStructuralElements(elements1, elements2 []models.StructuralElement) float64 {
	if len(elements1) == 0 && len(elements2) == 0 {
		return 1.0
	}
	
	if len(elements1) == 0 || len(elements2) == 0 {
		return 0.0
	}
	
	// Find best matching between structural elements
	totalSimilarity := 0.0
	matchCount := 0
	
	for _, e1 := range elements1 {
		bestSimilarity := 0.0
		for _, e2 := range elements2 {
			if e1.Type == e2.Type {
				// Calculate position similarity
				dx := e1.Position.X - e2.Position.X
				dy := e1.Position.Y - e2.Position.Y
				distance := math.Sqrt(dx*dx + dy*dy)
				positionSim := math.Exp(-distance / 50.0)
				
				// Calculate size similarity
				sizeSim := 0.5 // Default
				if maxSize := math.Max(e1.Size, e2.Size); maxSize > 0 {
					sizeSim = 1.0 - math.Abs(e1.Size-e2.Size)/maxSize
				}
				
				// Combine similarities
				similarity := safeFloat64(positionSim*0.7 + sizeSim*0.3, 0.0)
				if similarity > bestSimilarity {
					bestSimilarity = similarity
				}
			}
		}
		
		if bestSimilarity > 0.3 {
			totalSimilarity += bestSimilarity
			matchCount++
		}
	}
	
	if matchCount == 0 {
		return 0.0
	}
	
	result := totalSimilarity / float64(matchCount)
	return safeFloat64(result, 0.5)
}

func (ce *ComparisonEngine) compareReferencePoints(points1, points2 []models.Point2D) float64 {
	if len(points1) == 0 || len(points2) == 0 {
		return 0.5 // Neutral score if no reference points
	}
	
	// Find best matching between point sets
	totalDistance := 0.0
	matchCount := 0
	
	for _, p1 := range points1 {
		minDistance := math.Inf(1)
		for _, p2 := range points2 {
			distance := math.Sqrt(math.Pow(p1.X-p2.X, 2) + math.Pow(p1.Y-p2.Y, 2))
			if distance < minDistance {
				minDistance = distance
			}
		}
		
		if minDistance < 50.0 && !math.IsInf(minDistance, 0) { // Threshold for matching
			totalDistance += minDistance
			matchCount++
		}
	}
	
	if matchCount == 0 {
		return 0.0
	}
	
	avgDistance := totalDistance / float64(matchCount)
	
	// Convert distance to similarity (closer = more similar)
	result := math.Exp(-avgDistance / 20.0)
	return safeFloat64(result, 0.5)
}

func (ce *ComparisonEngine) compareLightPatterns(pattern1, pattern2 models.LightPatternFeatures) float64 {
	// Compare pattern signatures
	signatureSimilarity := ce.compareSignatures(pattern1.PatternSignature, pattern2.PatternSignature)
	
	// Compare individual light elements
	elementSimilarity := ce.compareLightElements(pattern1.LightElements, pattern2.LightElements)
	
	// Compare light configuration
	configSimilarity := ce.compareLightConfiguration(pattern1.LightConfiguration, pattern2.LightConfiguration)
	
	result := (signatureSimilarity*0.4 + elementSimilarity*0.4 + configSimilarity*0.2)
	return safeFloat64(result, 0.5)
}

func (ce *ComparisonEngine) compareSignatures(sig1, sig2 []float64) float64 {
	if len(sig1) != len(sig2) {
		return 0.0
	}
	
	if len(sig1) == 0 {
		return 1.0 // Empty signatures are identical
	}
	
	// Calculate cosine similarity
	dotProduct := 0.0
	norm1 := 0.0
	norm2 := 0.0
	
	for i := range sig1 {
		dotProduct += sig1[i] * sig2[i]
		norm1 += sig1[i] * sig1[i]
		norm2 += sig2[i] * sig2[i]
	}
	
	if norm1 == 0 || norm2 == 0 {
		return 0.0
	}
	
	denominator := math.Sqrt(norm1) * math.Sqrt(norm2)
	if denominator == 0 {
		return 0.0
	}
	
	result := dotProduct / denominator
	return safeFloat64(result, 0.0)
}

func (ce *ComparisonEngine) compareLightElements(elements1, elements2 []models.LightElement) float64 {
	if len(elements1) == 0 && len(elements2) == 0 {
		return 1.0
	}
	
	if len(elements1) == 0 || len(elements2) == 0 {
		return 0.0
	}
	
	// Find best matching between light elements
	totalSimilarity := 0.0
	matchCount := 0
	
	for _, e1 := range elements1 {
		bestSimilarity := 0.0
		for _, e2 := range elements2 {
			similarity := ce.compareSingleLightElement(e1, e2)
			if similarity > bestSimilarity {
				bestSimilarity = similarity
			}
		}
		
		if bestSimilarity > 0.3 { // Threshold for matching
			totalSimilarity += bestSimilarity
			matchCount++
		}
	}
	
	if matchCount == 0 {
		return 0.0
	}
	
	return totalSimilarity / float64(matchCount)
}

func (ce *ComparisonEngine) compareSingleLightElement(e1, e2 models.LightElement) float64 {
	// Elements must be same type
	if e1.Type != e2.Type {
		return 0.0
	}
	
	// Compare position (normalized)
	positionDistance := math.Sqrt(math.Pow(e1.Position.X-e2.Position.X, 2) + math.Pow(e1.Position.Y-e2.Position.Y, 2))
	positionSim := math.Exp(-positionDistance / 30.0)
	
	// Compare shape
	shapeSim := 0.0
	if e1.Shape == e2.Shape {
		shapeSim = 1.0
	}
	
	// Compare size
	sizeSim := 0.5 // Default
	if maxSize := math.Max(e1.Size, e2.Size); maxSize > 0 {
		sizeSim = 1.0 - math.Abs(e1.Size-e2.Size)/maxSize
	}
	
	// Compare intensity
	intensitySim := 1.0 - math.Abs(e1.Intensity-e2.Intensity)
	
	result := (positionSim*0.4 + shapeSim*0.2 + sizeSim*0.2 + intensitySim*0.2)
	return safeFloat64(result, 0.0)
}

func (ce *ComparisonEngine) compareLightConfiguration(config1, config2 models.LightConfiguration) float64 {
	// Compare number of elements
	numElementsSim := 0.5 // Default
	if maxElements := math.Max(float64(config1.NumElements), float64(config2.NumElements)); maxElements > 0 {
		numElementsSim = 1.0 - math.Abs(float64(config1.NumElements-config2.NumElements))/maxElements
	}
	
	// Compare symmetry
	symmetrySim := 1.0 - math.Abs(config1.Symmetry-config2.Symmetry)
	
	// Compare spacing
	spacingSim := 1.0
	if config1.Spacing > 0 && config2.Spacing > 0 {
		if maxSpacing := math.Max(config1.Spacing, config2.Spacing); maxSpacing > 0 {
			spacingSim = 1.0 - math.Abs(config1.Spacing-config2.Spacing)/maxSpacing
		}
	}
	
	result := (numElementsSim*0.4 + symmetrySim*0.3 + spacingSim*0.3)
	return safeFloat64(result, 0.5)
}

func (ce *ComparisonEngine) compareBumperFeatures(bumper1, bumper2 models.BumperFeatures) float64 {
	// Compare contour signatures
	contourSimilarity := ce.compareContours(bumper1.ContourSignature, bumper2.ContourSignature)
	
	// Compare texture features
	textureSimilarity := ce.compareSignatures(bumper1.TextureFeatures, bumper2.TextureFeatures)
	
	// Compare mounting points
	mountingSimilarity := ce.compareReferencePoints(bumper1.MountingPoints, bumper2.MountingPoints)
	
	// Compare license plate area
	plateAreaSimilarity := ce.comparePlateAreas(bumper1.LicensePlateArea, bumper2.LicensePlateArea)
	
	result := (contourSimilarity*0.3 + textureSimilarity*0.3 + mountingSimilarity*0.2 + plateAreaSimilarity*0.2)
	return safeFloat64(result, 0.5)
}

func (ce *ComparisonEngine) compareContours(contour1, contour2 []models.Point2D) float64 {
	if len(contour1) == 0 && len(contour2) == 0 {
		return 1.0
	}
	
	if len(contour1) == 0 || len(contour2) == 0 {
		return 0.0
	}
	
	// Simplified contour comparison using point matching
	totalDistance := 0.0
	matchCount := 0
	
	for _, p1 := range contour1 {
		minDistance := math.Inf(1)
		for _, p2 := range contour2 {
			distance := math.Sqrt(math.Pow(p1.X-p2.X, 2) + math.Pow(p1.Y-p2.Y, 2))
			if distance < minDistance {
				minDistance = distance
			}
		}
		
		if minDistance < 30.0 {
			totalDistance += minDistance
			matchCount++
		}
	}
	
	if matchCount == 0 {
		return 0.0
	}
	
	avgDistance := totalDistance / float64(matchCount)
	result := math.Exp(-avgDistance / 15.0)
	return safeFloat64(result, 0.5)
}

func (ce *ComparisonEngine) comparePlateAreas(area1, area2 models.Bounds) float64 {
	// Compare position
	centerX1 := float64(area1.X + area1.Width/2)
	centerY1 := float64(area1.Y + area1.Height/2)
	centerX2 := float64(area2.X + area2.Width/2)
	centerY2 := float64(area2.Y + area2.Height/2)
	
	centerDistance := math.Sqrt(math.Pow(centerX1-centerX2, 2) + math.Pow(centerY1-centerY2, 2))
	positionSim := math.Exp(-centerDistance / 20.0)
	
	// Compare size
	area1Size := float64(area1.Width * area1.Height)
	area2Size := float64(area2.Width * area2.Height)
	sizeSim := 0.5 // Default
	if maxArea := math.Max(area1Size, area2Size); maxArea > 0 {
		sizeSim = 1.0 - math.Abs(area1Size-area2Size)/maxArea
	}
	
	result := (positionSim*0.6 + sizeSim*0.4)
	return safeFloat64(result, 0.5)
}

func (ce *ComparisonEngine) compareDaylightFeatures(day1, day2 models.DaylightFeatures) float64 {
	// Compare color profiles
	colorSimilarity := ce.compareColorProfiles(day1.ColorProfile, day2.ColorProfile)
	
	// Compare badge locations
	badgeSimilarity := ce.compareBadgeFeatures(day1.BadgeLocations, day2.BadgeLocations)
	
	// Compare trim details
	trimSimilarity := ce.compareTrimFeatures(day1.TrimDetails, day2.TrimDetails)
	
	// Compare surface texture
	textureSimilarity := ce.compareTextureSignatures(day1.SurfaceTexture, day2.SurfaceTexture)
	
	result := (colorSimilarity*0.4 + badgeSimilarity*0.2 + trimSimilarity*0.2 + textureSimilarity*0.2)
	return safeFloat64(result, 0.5)
}

func (ce *ComparisonEngine) compareInfraredFeatures(ir1, ir2 models.InfraredFeatures) float64 {
	// If both have IR signatures, use them for comparison
	if ir1.IRSignature != nil && ir2.IRSignature != nil {
		return ce.compareIRSignatures(*ir1.IRSignature, *ir2.IRSignature)
	}
	
	// Fallback to basic thermal comparison
	// Compare thermal signatures
	thermalSimilarity := ce.compareSignatures(ir1.ThermalSignature, ir2.ThermalSignature)
	
	// Compare reflective elements
	reflectiveSimilarity := ce.compareReflectiveElements(ir1.ReflectiveElements, ir2.ReflectiveElements)
	
	// Compare heat patterns
	heatSimilarity := ce.compareHeatPatterns(ir1.HeatPatterns, ir2.HeatPatterns)
	
	// Compare material signatures
	materialSimilarity := ce.compareSignatures(ir1.MaterialSignature, ir2.MaterialSignature)
	
	result := (thermalSimilarity*0.3 + reflectiveSimilarity*0.3 + heatSimilarity*0.2 + materialSimilarity*0.2)
	return safeFloat64(result, 0.5)
}

// Placeholder methods for missing feature comparisons
func (ce *ComparisonEngine) compareColorProfiles(color1, color2 models.ColorProfile) float64 {
	// Compare dominant colors
	if len(color1.DominantColors) == 0 && len(color2.DominantColors) == 0 {
		return 1.0
	}
	
	if len(color1.DominantColors) == 0 || len(color2.DominantColors) == 0 {
		return 0.0
	}
	
	// Simple color comparison
	totalSimilarity := 0.0
	matchCount := 0
	
	for _, c1 := range color1.DominantColors {
		bestSimilarity := 0.0
		for _, c2 := range color2.DominantColors {
			// Calculate color distance in RGB space
			rDiff := float64(c1.R) - float64(c2.R)
			gDiff := float64(c1.G) - float64(c2.G)
			bDiff := float64(c1.B) - float64(c2.B)
			colorDistance := math.Sqrt(rDiff*rDiff + gDiff*gDiff + bDiff*bDiff)
			
			similarity := 1.0 - colorDistance/441.67 // 441.67 = sqrt(255^2 + 255^2 + 255^2)
			if similarity > bestSimilarity {
				bestSimilarity = similarity
			}
		}
		
		if bestSimilarity > 0.3 {
			totalSimilarity += bestSimilarity
			matchCount++
		}
	}
	
	if matchCount == 0 {
		return 0.0
	}
	
	return totalSimilarity / float64(matchCount)
}

func (ce *ComparisonEngine) compareBadgeFeatures(badges1, badges2 []models.BadgeFeature) float64 {
	return 0.5 // Placeholder implementation
}

func (ce *ComparisonEngine) compareTrimFeatures(trim1, trim2 []models.TrimFeature) float64 {
	return 0.5 // Placeholder implementation
}

func (ce *ComparisonEngine) compareTextureSignatures(texture1, texture2 models.TextureSignature) float64 {
	return ce.compareSignatures(texture1.Features, texture2.Features)
}

func (ce *ComparisonEngine) compareReflectiveElements(elements1, elements2 []models.ReflectiveElement) float64 {
	return 0.5 // Placeholder implementation
}

func (ce *ComparisonEngine) compareHeatPatterns(patterns1, patterns2 []models.HeatPattern) float64 {
	return 0.5 // Placeholder implementation
}

// compareIRSignatures compares IR signatures around license plates
func (ce *ComparisonEngine) compareIRSignatures(sig1, sig2 models.IRSignature) float64 {
	// Compare different components of the IR signature
	reflectivitySimilarity := ce.compareReflectivityMaps(sig1.ReflectivityMap, sig2.ReflectivityMap)
	materialSimilarity := ce.compareSignatures(sig1.MaterialSignature, sig2.MaterialSignature)
	illuminationSimilarity := ce.compareSignatures(sig1.IlluminationGradient, sig2.IlluminationGradient)
	shadowSimilarity := ce.compareShadowPatterns(sig1.ShadowPatterns, sig2.ShadowPatterns)
	textureSimilarity := ce.compareSignatures(sig1.TextureFeatures, sig2.TextureFeatures)
	
	// Weight the different components
	// Reflectivity map and material signature are most important for distinguishing vehicles
	// Shadow patterns and illumination gradients help with 3D structure
	// Texture provides additional discriminative power
	result := (reflectivitySimilarity*0.35 + 
			  materialSimilarity*0.30 + 
			  shadowSimilarity*0.15 + 
			  illuminationSimilarity*0.10 + 
			  textureSimilarity*0.10)
	
	return safeFloat64(result, 0.5)
}

func (ce *ComparisonEngine) compareReflectivityMaps(map1, map2 [][]float64) float64 {
	if len(map1) == 0 && len(map2) == 0 {
		return 1.0
	}
	
	if len(map1) == 0 || len(map2) == 0 {
		return 0.0
	}
	
	// Ensure maps are same size
	if len(map1) != len(map2) {
		return 0.0
	}
	
	var totalDifference float64
	var cellCount int
	
	for i := 0; i < len(map1); i++ {
		if len(map1[i]) != len(map2[i]) {
			continue
		}
		
		for j := 0; j < len(map1[i]); j++ {
			// Calculate absolute difference between reflectivity values
			diff := math.Abs(map1[i][j] - map2[i][j])
			totalDifference += diff
			cellCount++
		}
	}
	
	if cellCount == 0 {
		return 0.0
	}
	
	// Convert average difference to similarity (lower difference = higher similarity)
	avgDifference := totalDifference / float64(cellCount)
	similarity := 1.0 - avgDifference
	
	return safeFloat64(similarity, 0.5)
}

func (ce *ComparisonEngine) compareShadowPatterns(shadows1, shadows2 []models.Point2D) float64 {
	if len(shadows1) == 0 && len(shadows2) == 0 {
		return 1.0
	}
	
	if len(shadows1) == 0 || len(shadows2) == 0 {
		return 0.0
	}
	
	// Find best matching between shadow points
	totalSimilarity := 0.0
	matchCount := 0
	
	for _, s1 := range shadows1 {
		bestSimilarity := 0.0
		for _, s2 := range shadows2 {
			// Calculate distance between shadow points
			distance := math.Sqrt(math.Pow(s1.X-s2.X, 2) + math.Pow(s1.Y-s2.Y, 2))
			
			// Convert distance to similarity (closer = more similar)
			similarity := math.Exp(-distance / 50.0) // 50 pixel tolerance
			if similarity > bestSimilarity {
				bestSimilarity = similarity
			}
		}
		
		if bestSimilarity > 0.3 { // Minimum threshold for matching
			totalSimilarity += bestSimilarity
			matchCount++
		}
	}
	
	if matchCount == 0 {
		return 0.0
	}
	
	// Normalize by the number of shadows in the smaller set
	minShadows := math.Min(float64(len(shadows1)), float64(len(shadows2)))
	normalizedSimilarity := totalSimilarity / minShadows
	
	return safeFloat64(normalizedSimilarity, 0.5)
}

func (ce *ComparisonEngine) calculateWeightedSimilarity(scores models.DetailedScores, lighting models.LightingType) float64 {
	// Adjust weights based on lighting conditions
	var weights struct {
		geometric, lightPattern, bumper, color, thermal float64
	}
	
	if lighting == models.LightingDaylight {
		weights.geometric = 0.30
		weights.lightPattern = 0.30
		weights.bumper = 0.20
		weights.color = 0.20
		weights.thermal = 0.0
	} else { // Infrared
		weights.geometric = 0.35
		weights.lightPattern = 0.35
		weights.bumper = 0.20
		weights.color = 0.0
		weights.thermal = 0.10
	}
	
	result := (safeFloat64(scores.GeometricSimilarity, 0.5)*weights.geometric +
			safeFloat64(scores.LightPatternSimilarity, 0.5)*weights.lightPattern +
			safeFloat64(scores.BumperSimilarity, 0.5)*weights.bumper +
			safeFloat64(scores.ColorSimilarity, 0.5)*weights.color +
			safeFloat64(scores.ThermalSimilarity, 0.5)*weights.thermal)
	return safeFloat64(result, 0.5)
}

func (ce *ComparisonEngine) getSimilarityThreshold(lighting models.LightingType) float64 {
	if lighting == models.LightingDaylight {
		return 0.75 // Higher threshold for daylight (more features available)
	}
	return 0.70 // Slightly lower threshold for infrared
}

func (ce *ComparisonEngine) calculateConfidenceLevel(similarity float64, features1, features2 models.VehicleFeatures) models.ConfidenceLevel {
	// Calculate confidence based on:
	// - Feature extraction quality
	// - Lighting conditions
	// - Number of features available
	
	avgQuality := (features1.ExtractionQuality + features2.ExtractionQuality) / 2.0
	
	if avgQuality > 0.8 && similarity > 0.9 {
		return models.ConfidenceHigh
	} else if avgQuality > 0.6 && similarity > 0.8 {
		return models.ConfidenceHigh
	} else if avgQuality > 0.4 {
		return models.ConfidenceMedium
	}
	
	return models.ConfidenceLow
}