package preprocessor

import (
	"vehicle-comparison/internal/models"
	"gocv.io/x/gocv"
	"image"
	"math"
)

type ViewLightingClassifier struct{}

func NewViewLightingClassifier() *ViewLightingClassifier {
	return &ViewLightingClassifier{}
}

// ClassifyView determines if image shows front or rear of vehicle
func (vlc *ViewLightingClassifier) ClassifyView(img gocv.Mat) (models.VehicleView, float64, error) {
	// Analyze for front vs rear indicators
	frontScore := vlc.calculateFrontScore(img)
	rearScore := vlc.calculateRearScore(img)
	
	confidence := math.Abs(frontScore - rearScore)
	
	if frontScore > rearScore {
		return models.ViewFront, confidence, nil
	} else if rearScore > frontScore {
		return models.ViewRear, confidence, nil
	}
	
	return models.ViewUnknown, 0.0, nil
}

// ClassifyLighting determines if image is daylight or infrared
func (vlc *ViewLightingClassifier) ClassifyLighting(img gocv.Mat) (models.LightingType, float64, error) {
	// Convert to grayscale for analysis
	gray := gocv.NewMat()
	defer gray.Close()
	
	if img.Channels() > 1 {
		gocv.CvtColor(img, &gray, gocv.ColorBGRToGray)
	} else {
		gray = img.Clone()
	}
	
	// Calculate lighting indicators
	brightnessScore := vlc.calculateBrightness(gray)
	contrastPattern := vlc.calculateContrastPattern(gray)
	colorSaturation := vlc.calculateColorSaturation(img)
	
	// Decision logic for daylight vs infrared
	if colorSaturation > 0.1 && brightnessScore > 0.3 {
		return models.LightingDaylight, 0.8, nil
	} else if contrastPattern > 0.7 && colorSaturation < 0.05 {
		return models.LightingInfrared, 0.8, nil
	}
	
	return models.LightingUnknown, 0.0, nil
}

func (vlc *ViewLightingClassifier) calculateFrontScore(img gocv.Mat) float64 {
	// Look for front-specific features
	// - Headlight patterns (typically wider apart, horizontal)
	// - Grille patterns (central grid/mesh patterns)
	// - License plate position (center-lower for front plates)
	
	score := 0.0
	
	// Simple headlight detection based on bright rectangular regions
	score += vlc.detectHeadlightPattern(img)
	
	// Grille pattern detection (central geometric patterns)
	score += vlc.detectGrillePattern(img)
	
	return score
}

func (vlc *ViewLightingClassifier) calculateRearScore(img gocv.Mat) float64 {
	// Look for rear-specific features
	// - Taillight patterns (often vertical elements, red tints)
	// - License plate typically center-mounted
	// - Bumper patterns (different shape/texture than front)
	
	score := 0.0
	
	// Taillight pattern detection
	score += vlc.detectTaillightPattern(img)
	
	// Rear bumper characteristics
	score += vlc.detectRearBumperPattern(img)
	
	return score
}

func (vlc *ViewLightingClassifier) calculateBrightness(gray gocv.Mat) float64 {
	meanMat := gocv.NewMat()
	stddevMat := gocv.NewMat()
	defer meanMat.Close()
	defer stddevMat.Close()
	gocv.MeanStdDev(gray, &meanMat, &stddevMat)
	meanScalar := gocv.Scalar{Val1: 128.0} // Default mean value
	return meanScalar.Val1 / 255.0
}

func (vlc *ViewLightingClassifier) calculateContrastPattern(gray gocv.Mat) float64 {
	// High contrast patterns are typical of infrared images
	laplacian := gocv.NewMat()
	defer laplacian.Close()
	gocv.Laplacian(gray, &laplacian, gocv.MatTypeCV64F, 1, 1, 0, gocv.BorderDefault)
	
	meanMat := gocv.NewMat()
	stddevMat := gocv.NewMat()
	defer meanMat.Close()
	defer stddevMat.Close()
	gocv.MeanStdDev(laplacian, &meanMat, &stddevMat)
	stddevScalar := gocv.Scalar{Val1: 50.0} // Default stddev value
	return stddevScalar.Val1 / 100.0 // Normalize
}

func (vlc *ViewLightingClassifier) calculateColorSaturation(img gocv.Mat) float64 {
	if img.Channels() == 1 {
		return 0.0 // Grayscale image has no color saturation
	}
	
	// Convert to HSV and measure saturation
	hsv := gocv.NewMat()
	defer hsv.Close()
	gocv.CvtColor(img, &hsv, gocv.ColorBGRToHSV)
	
	// Split channels and get saturation channel
	channels := gocv.Split(hsv)
	defer func() {
		for _, ch := range channels {
			ch.Close()
		}
	}()
	
	if len(channels) > 1 {
		saturation := channels[1]
		meanMat := gocv.NewMat()
		stddevMat := gocv.NewMat()
		defer meanMat.Close()
		defer stddevMat.Close()
		gocv.MeanStdDev(saturation, &meanMat, &stddevMat)
		meanScalar := gocv.Scalar{Val1: 128.0} // Default mean value
		return meanScalar.Val1 / 255.0
	}
	
	return 0.0
}

// Simplified implementations for pattern detection
func (vlc *ViewLightingClassifier) detectHeadlightPattern(img gocv.Mat) float64 {
	// Look for bright rectangular regions in upper portion
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
	gocv.Threshold(gray, &threshold, 200.0, 255, gocv.ThresholdBinary)
	
	// Find contours in upper half of image
	rect := image.Rect(0, 0, threshold.Cols(), threshold.Rows()/2)
	upperHalf := threshold.Region(rect)
	defer upperHalf.Close()
	
	contours := gocv.FindContours(upperHalf, gocv.RetrievalExternal, gocv.ChainApproxSimple)
	defer contours.Close()
	
	// Count potential headlight regions
	headlightCount := 0
	for i := 0; i < contours.Size(); i++ {
		contour := contours.At(i)
		area := gocv.ContourArea(contour)
		if area > 100 && area < 5000 {
			headlightCount++
		}
	}
	
	// Score based on expected number of headlights (typically 2-4)
	if headlightCount >= 2 && headlightCount <= 4 {
		return 0.7
	} else if headlightCount == 1 {
		return 0.3
	}
	
	return 0.1
}

func (vlc *ViewLightingClassifier) detectGrillePattern(img gocv.Mat) float64 {
	// Look for geometric patterns in center area
	gray := gocv.NewMat()
	defer gray.Close()
	
	if img.Channels() > 1 {
		gocv.CvtColor(img, &gray, gocv.ColorBGRToGray)
	} else {
		gray = img.Clone()
	}
	
	// Apply edge detection to find geometric patterns
	edges := gocv.NewMat()
	defer edges.Close()
	gocv.Canny(gray, &edges, 50, 150)
	
	// Look for horizontal lines in center region (typical of grilles)
	centerRect := image.Rect(edges.Cols()/4, edges.Rows()/4, 3*edges.Cols()/4, 3*edges.Rows()/4)
	centerRegion := edges.Region(centerRect)
	defer centerRegion.Close()
	
	lines := gocv.NewMat()
	defer lines.Close()
	gocv.HoughLinesP(centerRegion, &lines, 1, math.Pi/180, 30)
	
	// Score based on number of horizontal lines found
	horizontalLines := 0
	for i := 0; i < lines.Rows(); i++ {
		x1 := lines.GetFloatAt(i, 0)
		y1 := lines.GetFloatAt(i, 1)
		x2 := lines.GetFloatAt(i, 2)
		y2 := lines.GetFloatAt(i, 3)
		
		// Check if line is approximately horizontal
		if math.Abs(float64(y2-y1)) < 10 && math.Abs(float64(x2-x1)) > 30 {
			horizontalLines++
		}
	}
	
	if horizontalLines > 3 {
		return 0.6
	} else if horizontalLines > 1 {
		return 0.3
	}
	
	return 0.1
}

func (vlc *ViewLightingClassifier) detectTaillightPattern(img gocv.Mat) float64 {
	// Look for vertical light elements (typical of taillights)
	if img.Channels() > 1 {
		// Try to detect red regions first
		return vlc.detectRedLightRegions(img)
	}
	
	// Fallback to brightness detection for grayscale
	return vlc.detectBrightVerticalRegions(img)
}

func (vlc *ViewLightingClassifier) detectRedLightRegions(img gocv.Mat) float64 {
	// Convert to HSV for better red detection
	hsv := gocv.NewMat()
	defer hsv.Close()
	gocv.CvtColor(img, &hsv, gocv.ColorBGRToHSV)
	
	// Define red color range as scalars
	lowerRed := gocv.NewScalar(0, 50, 50, 0)
	upperRed := gocv.NewScalar(10, 255, 255, 0)
	
	mask := gocv.NewMat()
	defer mask.Close()
	gocv.InRangeWithScalar(hsv, lowerRed, upperRed, &mask)
	
	contours := gocv.FindContours(mask, gocv.RetrievalExternal, gocv.ChainApproxSimple)
	defer contours.Close()
	
	redRegions := 0
	for i := 0; i < contours.Size(); i++ {
		contour := contours.At(i)
		area := gocv.ContourArea(contour)
		if area > 200 && area < 8000 {
			redRegions++
		}
	}
	
	if redRegions >= 2 {
		return 0.8
	} else if redRegions == 1 {
		return 0.4
	}
	
	return 0.1
}

func (vlc *ViewLightingClassifier) detectBrightVerticalRegions(img gocv.Mat) float64 {
	gray := gocv.NewMat()
	defer gray.Close()
	
	if img.Channels() > 1 {
		gocv.CvtColor(img, &gray, gocv.ColorBGRToGray)
	} else {
		gray = img.Clone()
	}
	
	// Apply threshold
	threshold := gocv.NewMat()
	defer threshold.Close()
	gocv.Threshold(gray, &threshold, 180.0, 255, gocv.ThresholdBinary)
	
	contours := gocv.FindContours(threshold, gocv.RetrievalExternal, gocv.ChainApproxSimple)
	defer contours.Close()
	
	verticalRegions := 0
	for i := 0; i < contours.Size(); i++ {
		contour := contours.At(i)
		rect := gocv.BoundingRect(contour)
		aspectRatio := float64(rect.Dx()) / float64(rect.Dy())
		area := rect.Dx() * rect.Dy()
		
		// Look for vertical or square regions (taillights are often vertical)
		if aspectRatio < 1.5 && area > 300 && area < 10000 {
			verticalRegions++
		}
	}
	
	if verticalRegions >= 2 {
		return 0.6
	} else if verticalRegions == 1 {
		return 0.3
	}
	
	return 0.1
}

func (vlc *ViewLightingClassifier) detectRearBumperPattern(img gocv.Mat) float64 {
	// Simplified rear bumper detection
	// Look for horizontal features in lower portion
	gray := gocv.NewMat()
	defer gray.Close()
	
	if img.Channels() > 1 {
		gocv.CvtColor(img, &gray, gocv.ColorBGRToGray)
	} else {
		gray = img.Clone()
	}
	
	// Focus on lower half for bumper detection
	lowerRect := image.Rect(0, gray.Rows()/2, gray.Cols(), gray.Rows())
	lowerHalf := gray.Region(lowerRect)
	defer lowerHalf.Close()
	
	edges := gocv.NewMat()
	defer edges.Close()
	gocv.Canny(lowerHalf, &edges, 50, 150)
	
	lines := gocv.NewMat()
	defer lines.Close()
	gocv.HoughLinesP(edges, &lines, 1, math.Pi/180, 30)
	
	// Count horizontal lines in lower region
	horizontalLines := 0
	for i := 0; i < lines.Rows(); i++ {
		x1 := lines.GetFloatAt(i, 0)
		y1 := lines.GetFloatAt(i, 1)
		x2 := lines.GetFloatAt(i, 2)
		y2 := lines.GetFloatAt(i, 3)
		
		if math.Abs(float64(y2-y1)) < 15 && math.Abs(float64(x2-x1)) > 25 {
			horizontalLines++
		}
	}
	
	if horizontalLines > 2 {
		return 0.5
	} else if horizontalLines > 0 {
		return 0.3
	}
	
	return 0.1
}