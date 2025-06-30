package extractor

import (
	"github.com/choff5507/vehicle-image-comparison/internal/models"
	"gocv.io/x/gocv"
	"image"
	"math"
)

type LightPatternExtractor struct{}

func NewLightPatternExtractor() *LightPatternExtractor {
	return &LightPatternExtractor{}
}

// ExtractLightPatterns extracts headlight/taillight patterns
func (lpe *LightPatternExtractor) ExtractLightPatterns(img gocv.Mat, view models.VehicleView, lighting models.LightingType) (models.LightPatternFeatures, error) {
	features := models.LightPatternFeatures{}
	
	switch view {
	case models.ViewFront:
		features = lpe.extractHeadlightPatterns(img, lighting)
	case models.ViewRear:
		features = lpe.extractTaillightPatterns(img, lighting)
	}
	
	return features, nil
}

func (lpe *LightPatternExtractor) extractHeadlightPatterns(img gocv.Mat, lighting models.LightingType) models.LightPatternFeatures {
	features := models.LightPatternFeatures{}
	
	// Find bright regions that could be headlights
	lightRegions := lpe.findBrightRegions(img, lighting)
	
	// Filter for headlight-like characteristics
	headlights := lpe.filterHeadlightCandidates(lightRegions, img)
	
	// Extract individual light elements
	features.LightElements = lpe.analyzeLightElements(headlights, models.TypeHeadlight)
	
	// Generate pattern signature
	features.PatternSignature = lpe.generatePatternSignature(features.LightElements)
	
	// Determine light configuration
	features.LightConfiguration = lpe.classifyLightConfiguration(features.LightElements)
	
	// Clean up regions
	for _, region := range lightRegions {
		region.Close()
	}
	for _, region := range headlights {
		region.Close()
	}
	
	return features
}

func (lpe *LightPatternExtractor) extractTaillightPatterns(img gocv.Mat, lighting models.LightingType) models.LightPatternFeatures {
	features := models.LightPatternFeatures{}
	
	// Find light regions (different approach for taillights)
	lightRegions := lpe.findTaillightRegions(img, lighting)
	
	// Filter for taillight characteristics
	taillights := lpe.filterTaillightCandidates(lightRegions, img)
	
	// Extract individual light elements
	features.LightElements = lpe.analyzeLightElements(taillights, models.TypeTaillight)
	
	// Generate pattern signature
	features.PatternSignature = lpe.generatePatternSignature(features.LightElements)
	
	// Determine light configuration
	features.LightConfiguration = lpe.classifyLightConfiguration(features.LightElements)
	
	// Clean up regions
	for _, region := range lightRegions {
		region.Close()
	}
	for _, region := range taillights {
		region.Close()
	}
	
	return features
}

func (lpe *LightPatternExtractor) findBrightRegions(img gocv.Mat, lighting models.LightingType) []gocv.Mat {
	gray := gocv.NewMat()
	defer gray.Close()
	
	if img.Channels() > 1 {
		gocv.CvtColor(img, &gray, gocv.ColorBGRToGray)
	} else {
		gray = img.Clone()
	}
	
	// Apply threshold to find bright regions
	threshold := gocv.NewMat()
	defer threshold.Close()
	
	var thresholdValue float64
	if lighting == models.LightingInfrared {
		thresholdValue = 200.0 // Higher threshold for IR
	} else {
		thresholdValue = 180.0 // Lower threshold for daylight
	}
	
	gocv.Threshold(gray, &threshold, float32(thresholdValue), 255, gocv.ThresholdBinary)
	
	// Apply morphological operations to clean up
	kernel := gocv.GetStructuringElement(gocv.MorphEllipse, image.Pt(5, 5))
	defer kernel.Close()
	
	cleaned := gocv.NewMat()
	defer cleaned.Close()
	gocv.MorphologyEx(threshold, &cleaned, gocv.MorphOpen, kernel)
	
	// Find contours of bright regions
	contours := gocv.FindContours(cleaned, gocv.RetrievalExternal, gocv.ChainApproxSimple)
	defer contours.Close()
	
	regions := []gocv.Mat{}
	for i := 0; i < contours.Size(); i++ {
		contour := contours.At(i)
		rect := gocv.BoundingRect(contour)
		
		// Filter by size
		area := rect.Dx() * rect.Dy()
		if area > 100 && area < 10000 {
			roi := img.Region(rect)
			regions = append(regions, roi.Clone())
			roi.Close()
		}
	}
	
	return regions
}

func (lpe *LightPatternExtractor) findTaillightRegions(img gocv.Mat, lighting models.LightingType) []gocv.Mat {
	// For taillights, we might look for different characteristics
	// - Red color in daylight images
	// - Specific shapes (often vertical or L-shaped)
	
	if lighting == models.LightingDaylight {
		return lpe.findRedRegions(img)
	} else {
		return lpe.findBrightRegions(img, lighting)
	}
}

func (lpe *LightPatternExtractor) findRedRegions(img gocv.Mat) []gocv.Mat {
	if img.Channels() == 1 {
		// Grayscale image, fall back to brightness detection
		return lpe.findBrightRegions(img, models.LightingDaylight)
	}
	
	// Convert to HSV for better color detection
	hsv := gocv.NewMat()
	defer hsv.Close()
	gocv.CvtColor(img, &hsv, gocv.ColorBGRToHSV)
	
	// Define red color ranges in HSV as scalars
	lowerRed1 := gocv.NewScalar(0, 50, 50, 0)
	upperRed1 := gocv.NewScalar(10, 255, 255, 0)
	lowerRed2 := gocv.NewScalar(170, 50, 50, 0)
	upperRed2 := gocv.NewScalar(180, 255, 255, 0)
	
	// Create masks for red regions
	mask1 := gocv.NewMat()
	mask2 := gocv.NewMat()
	redMask := gocv.NewMat()
	defer mask1.Close()
	defer mask2.Close()
	defer redMask.Close()
	
	gocv.InRangeWithScalar(hsv, lowerRed1, upperRed1, &mask1)
	gocv.InRangeWithScalar(hsv, lowerRed2, upperRed2, &mask2)
	gocv.BitwiseOr(mask1, mask2, &redMask)
	
	// Find contours of red regions
	contours := gocv.FindContours(redMask, gocv.RetrievalExternal, gocv.ChainApproxSimple)
	defer contours.Close()
	
	regions := []gocv.Mat{}
	for i := 0; i < contours.Size(); i++ {
		contour := contours.At(i)
		rect := gocv.BoundingRect(contour)
		
		// Filter by size and aspect ratio
		area := rect.Dx() * rect.Dy()
		aspectRatio := float64(rect.Dx()) / float64(rect.Dy())
		
		if area > 200 && area < 8000 && aspectRatio > 0.3 && aspectRatio < 3.0 {
			roi := img.Region(rect)
			regions = append(regions, roi.Clone())
			roi.Close()
		}
	}
	
	return regions
}

func (lpe *LightPatternExtractor) filterHeadlightCandidates(regions []gocv.Mat, fullImage gocv.Mat) []gocv.Mat {
	// Filter regions based on headlight characteristics
	// - Position (typically in upper portion of vehicle)
	// - Size (reasonable for headlights)
	// - Pair detection (headlights usually come in pairs)
	
	filtered := []gocv.Mat{}
	
	for _, region := range regions {
		// Check if region is reasonable size for headlights
		if region.Cols() > 20 && region.Rows() > 15 && region.Cols() < 200 && region.Rows() < 150 {
			filtered = append(filtered, region.Clone())
		}
	}
	
	return filtered
}

func (lpe *LightPatternExtractor) filterTaillightCandidates(regions []gocv.Mat, fullImage gocv.Mat) []gocv.Mat {
	// Filter regions based on taillight characteristics
	// - Position (various positions depending on vehicle type)
	// - Shape (often vertical or complex shapes)
	// - Pair detection
	
	filtered := []gocv.Mat{}
	
	for _, region := range regions {
		// Basic size filtering for taillights
		if region.Cols() > 15 && region.Rows() > 20 && region.Cols() < 150 && region.Rows() < 200 {
			filtered = append(filtered, region.Clone())
		}
	}
	
	return filtered
}

func (lpe *LightPatternExtractor) analyzeLightElements(lightRegions []gocv.Mat, lightType models.LightType) []models.LightElement {
	elements := []models.LightElement{}
	
	for i, region := range lightRegions {
		element := models.LightElement{
			Type:      lightType,
			Position:  models.Point2D{X: float64(i * 100), Y: 50}, // Placeholder position
			Shape:     lpe.classifyLightShape(region),
			Size:      lpe.calculateLightSize(region),
			Intensity: lpe.calculateLightIntensity(region),
		}
		
		elements = append(elements, element)
	}
	
	return elements
}

func (lpe *LightPatternExtractor) classifyLightShape(region gocv.Mat) models.LightShape {
	// Analyze region shape to classify light type
	aspectRatio := float64(region.Cols()) / float64(region.Rows())
	
	if aspectRatio > 1.5 {
		return models.ShapeRectangular
	} else if aspectRatio < 0.7 {
		return models.ShapeRectangular // Vertical rectangular
	} else {
		// Analyze contour to determine if round or angular
		gray := gocv.NewMat()
		defer gray.Close()
		
		if region.Channels() > 1 {
			gocv.CvtColor(region, &gray, gocv.ColorBGRToGray)
		} else {
			gray = region.Clone()
		}
		
		// Apply threshold
		threshold := gocv.NewMat()
		defer threshold.Close()
		gocv.Threshold(gray, &threshold, 127, 255, gocv.ThresholdBinary)
		
		// Find contours
		contours := gocv.FindContours(threshold, gocv.RetrievalExternal, gocv.ChainApproxSimple)
		defer contours.Close()
		
		if contours.Size() > 0 {
			// Analyze largest contour
			largestArea := 0.0
			largestContour := 0
			
			for i := 0; i < contours.Size(); i++ {
				contour := contours.At(i)
				area := gocv.ContourArea(contour)
				if area > largestArea {
					largestArea = area
					largestContour = i
				}
			}
			
			// Check circularity
			largestContourMat := contours.At(largestContour)
			perimeter := gocv.ArcLength(largestContourMat, true)
			circularity := 4 * math.Pi * largestArea / (perimeter * perimeter)
			
			if circularity > 0.7 {
				return models.ShapeRound
			} else {
				return models.ShapeAngular
			}
		}
		
		return models.ShapeCustom
	}
}

func (lpe *LightPatternExtractor) calculateLightSize(region gocv.Mat) float64 {
	return float64(region.Cols() * region.Rows())
}

func (lpe *LightPatternExtractor) calculateLightIntensity(region gocv.Mat) float64 {
	gray := gocv.NewMat()
	defer gray.Close()
	
	if region.Channels() > 1 {
		gocv.CvtColor(region, &gray, gocv.ColorBGRToGray)
	} else {
		gray = region.Clone()
	}
	
	meanMat := gocv.NewMat()
	stddevMat := gocv.NewMat()
	defer meanMat.Close()
	defer stddevMat.Close()
	gocv.MeanStdDev(gray, &meanMat, &stddevMat)
	meanScalar := gocv.Scalar{Val1: 128.0} // Default mean value
	return meanScalar.Val1 / 255.0
}

func (lpe *LightPatternExtractor) generatePatternSignature(elements []models.LightElement) []float64 {
	// Generate a feature vector representing the light pattern
	signature := make([]float64, 10) // Fixed-size signature
	
	for i, element := range elements {
		if i >= len(signature) {
			break
		}
		
		// Encode element characteristics into signature
		signature[i] = element.Size + element.Intensity + float64(element.Shape)
	}
	
	// Add pattern-level features
	if len(elements) > 0 {
		// Average intensity
		avgIntensity := 0.0
		for _, e := range elements {
			avgIntensity += e.Intensity
		}
		avgIntensity /= float64(len(elements))
		
		// Add to signature if space available
		if len(signature) > len(elements) {
			signature[len(elements)] = avgIntensity
		}
		
		// Symmetry measure (simplified)
		if len(elements) >= 2 && len(signature) > len(elements)+1 {
			leftElements := 0
			rightElements := 0
			centerX := 0.0
			
			// Calculate center
			for _, e := range elements {
				centerX += e.Position.X
			}
			centerX /= float64(len(elements))
			
			// Count elements on each side
			for _, e := range elements {
				if e.Position.X < centerX {
					leftElements++
				} else {
					rightElements++
				}
			}
			
			symmetry := 1.0 - math.Abs(float64(leftElements-rightElements))/float64(len(elements))
			signature[len(elements)+1] = symmetry
		}
	}
	
	return signature
}

func (lpe *LightPatternExtractor) classifyLightConfiguration(elements []models.LightElement) models.LightConfiguration {
	config := models.LightConfiguration{
		NumElements: len(elements),
		Symmetry:    0.5, // Default
		Spacing:     0.0,
	}
	
	if len(elements) < 2 {
		return config
	}
	
	// Calculate symmetry
	leftElements := 0
	rightElements := 0
	centerX := 0.0
	
	// Calculate center
	for _, e := range elements {
		centerX += e.Position.X
	}
	centerX /= float64(len(elements))
	
	// Count elements on each side
	for _, e := range elements {
		if e.Position.X < centerX {
			leftElements++
		} else {
			rightElements++
		}
	}
	
	config.Symmetry = 1.0 - math.Abs(float64(leftElements-rightElements))/float64(len(elements))
	
	// Calculate average spacing
	if len(elements) > 1 {
		totalSpacing := 0.0
		for i := 0; i < len(elements)-1; i++ {
			for j := i + 1; j < len(elements); j++ {
				dx := elements[i].Position.X - elements[j].Position.X
				dy := elements[i].Position.Y - elements[j].Position.Y
				distance := math.Sqrt(dx*dx + dy*dy)
				totalSpacing += distance
			}
		}
		
		numPairs := len(elements) * (len(elements) - 1) / 2
		config.Spacing = totalSpacing / float64(numPairs)
	}
	
	return config
}