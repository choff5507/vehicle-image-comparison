package extractor

import (
	"vehicle-comparison/internal/models"
	"gocv.io/x/gocv"
	"image"
	"math"
)

type LicensePlateExtractor struct {
	minPlateWidth  int
	maxPlateWidth  int
	minPlateHeight int
	maxPlateHeight int
	aspectRatioMin float64
	aspectRatioMax float64
}

func NewLicensePlateExtractor() *LicensePlateExtractor {
	return &LicensePlateExtractor{
		minPlateWidth:  80,   // Minimum plate width in pixels
		maxPlateWidth:  400,  // Maximum plate width in pixels
		minPlateHeight: 20,   // Minimum plate height in pixels
		maxPlateHeight: 120,  // Maximum plate height in pixels
		aspectRatioMin: 2.0,  // US plates are roughly 2:1 to 4:1 ratio
		aspectRatioMax: 4.5,
	}
}

// DetectLicensePlate finds the license plate region in IR images
func (lpe *LicensePlateExtractor) DetectLicensePlate(img gocv.Mat) (*models.LicensePlateRegion, error) {
	// Convert to grayscale if not already
	gray := gocv.NewMat()
	defer gray.Close()
	
	if img.Channels() == 3 {
		gocv.CvtColor(img, &gray, gocv.ColorBGRToGray)
	} else {
		gray = img.Clone()
	}
	
	// For IR images, license plates are typically the brightest regions
	// Apply threshold to find bright regions
	thresh := gocv.NewMat()
	defer thresh.Close()
	
	// Use adaptive threshold to handle varying illumination
	gocv.AdaptiveThreshold(gray, &thresh, 255, gocv.AdaptiveThresholdMean, gocv.ThresholdBinary, 15, -2)
	
	// Also try simple threshold for very bright regions (retroreflective plates)
	brightThresh := gocv.NewMat()
	defer brightThresh.Close()
	gocv.Threshold(gray, &brightThresh, 200, 255, gocv.ThresholdBinary)
	
	// Combine both thresholding approaches
	combined := gocv.NewMat()
	defer combined.Close()
	gocv.BitwiseOr(thresh, brightThresh, &combined)
	
	// Find contours
	contours := gocv.FindContours(combined, gocv.RetrievalExternal, gocv.ChainApproxSimple)
	
	var bestCandidate *models.LicensePlateRegion
	var bestScore float64
	
	for i := 0; i < contours.Size(); i++ {
		contour := contours.At(i)
		
		// Get bounding rectangle
		rect := gocv.BoundingRect(contour)
		
		// Check if dimensions match typical license plate proportions
		if lpe.isValidPlateSize(rect) {
			score := lpe.scorePlateCandidate(gray, rect, contour)
			
			if score > bestScore {
				bestScore = score
				bestCandidate = &models.LicensePlateRegion{
					Bounds: models.Bounds{
						X:      rect.Min.X,
						Y:      rect.Min.Y,
						Width:  rect.Dx(),
						Height: rect.Dy(),
					},
					Confidence:    score,
					AvgBrightness: lpe.calculateAverageBrightness(gray, rect),
					IsReflective:  lpe.isReflectiveRegion(gray, rect),
				}
			}
		}
	}
	
	if bestCandidate == nil {
		// Fallback: find brightest rectangular region
		bestCandidate = lpe.findBrightestRectangularRegion(gray)
	}
	
	return bestCandidate, nil
}

func (lpe *LicensePlateExtractor) isValidPlateSize(rect image.Rectangle) bool {
	width := rect.Dx()
	height := rect.Dy()
	
	// Check size constraints
	if width < lpe.minPlateWidth || width > lpe.maxPlateWidth {
		return false
	}
	if height < lpe.minPlateHeight || height > lpe.maxPlateHeight {
		return false
	}
	
	// Check aspect ratio
	aspectRatio := float64(width) / float64(height)
	if aspectRatio < lpe.aspectRatioMin || aspectRatio > lpe.aspectRatioMax {
		return false
	}
	
	return true
}

func (lpe *LicensePlateExtractor) scorePlateCandidate(gray gocv.Mat, rect image.Rectangle, contour gocv.PointVector) float64 {
	score := 0.0
	
	// Score based on brightness (higher is better for IR retroreflective plates)
	avgBrightness := lpe.calculateAverageBrightness(gray, rect)
	brightnessScore := math.Min(avgBrightness/255.0, 1.0)
	score += brightnessScore * 0.4
	
	// Score based on rectangularity (how close to rectangle shape)
	area := float64(rect.Dx() * rect.Dy())
	contourArea := gocv.ContourArea(contour)
	rectangularityScore := contourArea / area
	score += rectangularityScore * 0.3
	
	// Score based on aspect ratio (closer to typical plate ratio is better)
	aspectRatio := float64(rect.Dx()) / float64(rect.Dy())
	idealRatio := 3.0 // Typical US plate ratio
	ratioScore := 1.0 - math.Abs(aspectRatio-idealRatio)/idealRatio
	score += ratioScore * 0.2
	
	// Score based on position (plates typically in lower portion of image)
	positionScore := 1.0
	if rect.Min.Y < gray.Rows()/3 {
		positionScore = 0.5 // Penalize plates in upper portion
	}
	score += positionScore * 0.1
	
	return math.Max(0.0, math.Min(1.0, score))
}

func (lpe *LicensePlateExtractor) calculateAverageBrightness(gray gocv.Mat, rect image.Rectangle) float64 {
	// Extract region of interest
	roi := gray.Region(rect)
	defer roi.Close()
	
	// Calculate mean brightness
	mean := gocv.NewMat()
	defer mean.Close()
	gocv.MeanStdDev(roi, &mean, nil)
	
	// Get the mean value from the result
	return mean.GetDoubleAt(0, 0)
}

func (lpe *LicensePlateExtractor) isReflectiveRegion(gray gocv.Mat, rect image.Rectangle) bool {
	avgBrightness := lpe.calculateAverageBrightness(gray, rect)
	// Consider reflective if average brightness is above threshold
	return avgBrightness > 180
}

func (lpe *LicensePlateExtractor) findBrightestRectangularRegion(gray gocv.Mat) *models.LicensePlateRegion {
	// Fallback method: find brightest region that could be a license plate
	rows := gray.Rows()
	cols := gray.Cols()
	
	var bestRegion *models.LicensePlateRegion
	var maxBrightness float64
	
	// Search in the lower 2/3 of the image where plates are typically located
	startY := rows / 3
	
	// Try different rectangular regions with plate-like proportions
	for y := startY; y < rows-lpe.minPlateHeight; y += 10 {
		for x := 0; x < cols-lpe.minPlateWidth; x += 10 {
			for width := lpe.minPlateWidth; width <= lpe.maxPlateWidth && x+width < cols; width += 20 {
				height := int(float64(width) / 3.0) // Approximate 3:1 ratio
				if height < lpe.minPlateHeight || height > lpe.maxPlateHeight || y+height >= rows {
					continue
				}
				
				rect := image.Rect(x, y, x+width, y+height)
				brightness := lpe.calculateAverageBrightness(gray, rect)
				
				if brightness > maxBrightness {
					maxBrightness = brightness
					bestRegion = &models.LicensePlateRegion{
						Bounds: models.Bounds{
							X:      x,
							Y:      y,
							Width:  width,
							Height: height,
						},
						Confidence:    brightness / 255.0,
						AvgBrightness: brightness,
						IsReflective:  brightness > 180,
					}
				}
			}
		}
	}
	
	if bestRegion == nil {
		// Ultimate fallback: use bottom center of image
		width := lpe.minPlateWidth * 2
		height := int(float64(width) / 3.0)
		x := (cols - width) / 2
		y := rows - height - 20
		
		bestRegion = &models.LicensePlateRegion{
			Bounds: models.Bounds{
				X:      x,
				Y:      y,
				Width:  width,
				Height: height,
			},
			Confidence:    0.1,
			AvgBrightness: 0,
			IsReflective:  false,
		}
	}
	
	return bestRegion
}