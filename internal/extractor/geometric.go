package extractor

import (
	"vehicle-comparison/internal/models"
	"gocv.io/x/gocv"
	"image"
	"math"
)

type GeometricExtractor struct{}

func NewGeometricExtractor() *GeometricExtractor {
	return &GeometricExtractor{}
}

// ExtractGeometricFeatures extracts view-consistent geometric features
func (ge *GeometricExtractor) ExtractGeometricFeatures(img gocv.Mat, view models.VehicleView) (models.GeometricFeatures, error) {
	features := models.GeometricFeatures{}
	
	// Extract vehicle proportions
	features.VehicleProportions = ge.extractVehicleProportions(img)
	
	// Extract structural elements based on view
	features.StructuralElements = ge.extractStructuralElements(img, view)
	
	// Extract reference points for alignment
	features.ReferencePoints = ge.extractReferencePoints(img, view)
	
	return features, nil
}

func (ge *GeometricExtractor) extractVehicleProportions(img gocv.Mat) models.VehicleProportions {
	height := float64(img.Rows())
	width := float64(img.Cols())
	
	// Calculate basic proportional relationships
	widthHeightRatio := width / height
	
	// Estimate upper/lower vehicle proportions
	upperLowerRatio := ge.calculateUpperLowerRatio(img)
	
	// License plate proportion (if detectable)
	licensePlateRatio := ge.estimateLicensePlateRatio(img)
	
	return models.VehicleProportions{
		WidthHeightRatio:  widthHeightRatio,
		UpperLowerRatio:   upperLowerRatio,
		LicensePlateRatio: licensePlateRatio,
	}
}

func (ge *GeometricExtractor) calculateUpperLowerRatio(img gocv.Mat) float64 {
	// Find horizontal edge that divides upper/lower vehicle
	gray := gocv.NewMat()
	defer gray.Close()
	
	if img.Channels() > 1 {
		gocv.CvtColor(img, &gray, gocv.ColorBGRToGray)
	} else {
		gray = img.Clone()
	}
	
	// Apply edge detection
	edges := gocv.NewMat()
	defer edges.Close()
	gocv.Canny(gray, &edges, 50, 150)
	
	// Find horizontal lines using Hough transform
	lines := gocv.NewMat()
	defer lines.Close()
	gocv.HoughLinesP(edges, &lines, 1, math.Pi/180, 50)
	
	// Analyze horizontal lines to find vehicle division
	dividingLine := ge.findVehicleDividingLine(lines, img.Rows())
	
	if dividingLine > 0 {
		upperHeight := float64(dividingLine)
		lowerHeight := float64(img.Rows() - dividingLine)
		return upperHeight / lowerHeight
	}
	
	return 1.0 // Default ratio if no clear division found
}

func (ge *GeometricExtractor) findVehicleDividingLine(lines gocv.Mat, imageHeight int) int {
	// Analyze detected lines to find the main horizontal division
	// This would typically be around the bumper line
	
	horizontalLines := []int{}
	
	for i := 0; i < lines.Rows(); i++ {
		x1 := int(lines.GetFloatAt(i, 0))
		y1 := int(lines.GetFloatAt(i, 1))
		x2 := int(lines.GetFloatAt(i, 2))
		y2 := int(lines.GetFloatAt(i, 3))
		
		// Check if line is approximately horizontal
		if math.Abs(float64(y2-y1)) < 10 && math.Abs(float64(x2-x1)) > 50 {
			horizontalLines = append(horizontalLines, (y1+y2)/2)
		}
	}
	
	if len(horizontalLines) == 0 {
		return 0
	}
	
	// Find the most prominent horizontal line (simple mode calculation)
	return ge.findMostCommonValue(horizontalLines)
}

func (ge *GeometricExtractor) estimateLicensePlateRatio(img gocv.Mat) float64 {
	// Detect rectangular regions that could be license plates
	gray := gocv.NewMat()
	defer gray.Close()
	
	if img.Channels() > 1 {
		gocv.CvtColor(img, &gray, gocv.ColorBGRToGray)
	} else {
		gray = img.Clone()
	}
	
	// Apply edge detection and morphological operations
	edges := gocv.NewMat()
	defer edges.Close()
	gocv.Canny(gray, &edges, 50, 150)
	
	// Find contours
	contours := gocv.FindContours(edges, gocv.RetrievalExternal, gocv.ChainApproxSimple)
	defer contours.Close()
	
	// Look for rectangular contours with license plate aspect ratio
	for i := 0; i < contours.Size(); i++ {
		contour := contours.At(i)
		rect := gocv.BoundingRect(contour)
		aspectRatio := float64(rect.Dx()) / float64(rect.Dy())
		
		// License plates are typically 2:1 ratio
		if aspectRatio > 1.5 && aspectRatio < 2.5 {
			// Check if size is reasonable for a license plate
			area := rect.Dx() * rect.Dy()
			imageArea := img.Cols() * img.Rows()
			
			if float64(area)/float64(imageArea) > 0.01 && float64(area)/float64(imageArea) < 0.15 {
				vehicleWidth := float64(img.Cols())
				plateWidth := float64(rect.Dx())
				return plateWidth / vehicleWidth
			}
		}
	}
	
	return 0.0 // No license plate detected
}

func (ge *GeometricExtractor) extractStructuralElements(img gocv.Mat, view models.VehicleView) []models.StructuralElement {
	elements := []models.StructuralElement{}
	
	switch view {
	case models.ViewFront:
		elements = append(elements, ge.extractFrontStructuralElements(img)...)
	case models.ViewRear:
		elements = append(elements, ge.extractRearStructuralElements(img)...)
	}
	
	return elements
}

func (ge *GeometricExtractor) extractFrontStructuralElements(img gocv.Mat) []models.StructuralElement {
	// Extract front-specific structural elements
	elements := []models.StructuralElement{}
	
	// Detect headlight regions
	headlights := ge.detectHeadlightRegions(img)
	for _, hl := range headlights {
		elements = append(elements, models.StructuralElement{
			Type:     "headlight",
			Position: hl,
			Size:     100.0, // Placeholder
		})
	}
	
	// Detect grille area
	grilleCenter := ge.detectGrilleCenter(img)
	if grilleCenter.X > 0 && grilleCenter.Y > 0 {
		elements = append(elements, models.StructuralElement{
			Type:     "grille",
			Position: grilleCenter,
			Size:     200.0, // Placeholder
		})
	}
	
	return elements
}

func (ge *GeometricExtractor) extractRearStructuralElements(img gocv.Mat) []models.StructuralElement {
	// Extract rear-specific structural elements
	elements := []models.StructuralElement{}
	
	// Detect taillight regions
	taillights := ge.detectTaillightRegions(img)
	for _, tl := range taillights {
		elements = append(elements, models.StructuralElement{
			Type:     "taillight",
			Position: tl,
			Size:     80.0, // Placeholder
		})
	}
	
	// Detect rear bumper line
	bumperLine := ge.detectRearBumperLine(img)
	if bumperLine.X > 0 && bumperLine.Y > 0 {
		elements = append(elements, models.StructuralElement{
			Type:     "bumper_line",
			Position: bumperLine,
			Size:     float64(img.Cols()), // Width of the line
		})
	}
	
	return elements
}

func (ge *GeometricExtractor) detectHeadlightRegions(img gocv.Mat) []models.Point2D {
	points := []models.Point2D{}
	
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
	
	// Focus on upper half of image
	upperRect := image.Rect(0, 0, threshold.Cols(), threshold.Rows()/2)
	upperHalf := threshold.Region(upperRect)
	defer upperHalf.Close()
	
	contours := gocv.FindContours(upperHalf, gocv.RetrievalExternal, gocv.ChainApproxSimple)
	defer contours.Close()
	
	for i := 0; i < contours.Size(); i++ {
		contour := contours.At(i)
		area := gocv.ContourArea(contour)
		if area > 100 && area < 5000 {
			rect := gocv.BoundingRect(contour)
			centerX := float64(rect.Min.X + rect.Dx()/2)
			centerY := float64(rect.Min.Y + rect.Dy()/2)
			points = append(points, models.Point2D{X: centerX, Y: centerY})
		}
	}
	
	return points
}

func (ge *GeometricExtractor) detectTaillightRegions(img gocv.Mat) []models.Point2D {
	points := []models.Point2D{}
	
	// Try to detect red regions if color image
	if img.Channels() > 1 {
		points = ge.detectRedRegionCenters(img)
	}
	
	// Fallback to brightness detection
	if len(points) == 0 {
		points = ge.detectBrightRegionCenters(img)
	}
	
	return points
}

func (ge *GeometricExtractor) detectRedRegionCenters(img gocv.Mat) []models.Point2D {
	points := []models.Point2D{}
	
	hsv := gocv.NewMat()
	defer hsv.Close()
	gocv.CvtColor(img, &hsv, gocv.ColorBGRToHSV)
	
	// Define red color ranges as scalars
	lowerRed1 := gocv.NewScalar(0, 50, 50, 0)
	upperRed1 := gocv.NewScalar(10, 255, 255, 0)
	lowerRed2 := gocv.NewScalar(170, 50, 50, 0)
	upperRed2 := gocv.NewScalar(180, 255, 255, 0)
	
	mask1 := gocv.NewMat()
	mask2 := gocv.NewMat()
	redMask := gocv.NewMat()
	defer mask1.Close()
	defer mask2.Close()
	defer redMask.Close()
	
	gocv.InRangeWithScalar(hsv, lowerRed1, upperRed1, &mask1)
	gocv.InRangeWithScalar(hsv, lowerRed2, upperRed2, &mask2)
	gocv.BitwiseOr(mask1, mask2, &redMask)
	
	contours := gocv.FindContours(redMask, gocv.RetrievalExternal, gocv.ChainApproxSimple)
	defer contours.Close()
	
	for i := 0; i < contours.Size(); i++ {
		contour := contours.At(i)
		area := gocv.ContourArea(contour)
		if area > 200 && area < 8000 {
			rect := gocv.BoundingRect(contour)
			centerX := float64(rect.Min.X + rect.Dx()/2)
			centerY := float64(rect.Min.Y + rect.Dy()/2)
			points = append(points, models.Point2D{X: centerX, Y: centerY})
		}
	}
	
	return points
}

func (ge *GeometricExtractor) detectBrightRegionCenters(img gocv.Mat) []models.Point2D {
	points := []models.Point2D{}
	
	gray := gocv.NewMat()
	defer gray.Close()
	
	if img.Channels() > 1 {
		gocv.CvtColor(img, &gray, gocv.ColorBGRToGray)
	} else {
		gray = img.Clone()
	}
	
	threshold := gocv.NewMat()
	defer threshold.Close()
	gocv.Threshold(gray, &threshold, 180.0, 255, gocv.ThresholdBinary)
	
	contours := gocv.FindContours(threshold, gocv.RetrievalExternal, gocv.ChainApproxSimple)
	defer contours.Close()
	
	for i := 0; i < contours.Size(); i++ {
		contour := contours.At(i)
		area := gocv.ContourArea(contour)
		if area > 300 && area < 10000 {
			rect := gocv.BoundingRect(contour)
			centerX := float64(rect.Min.X + rect.Dx()/2)
			centerY := float64(rect.Min.Y + rect.Dy()/2)
			points = append(points, models.Point2D{X: centerX, Y: centerY})
		}
	}
	
	return points
}

func (ge *GeometricExtractor) detectGrilleCenter(img gocv.Mat) models.Point2D {
	gray := gocv.NewMat()
	defer gray.Close()
	
	if img.Channels() > 1 {
		gocv.CvtColor(img, &gray, gocv.ColorBGRToGray)
	} else {
		gray = img.Clone()
	}
	
	// Focus on center region for grille detection
	centerRect := image.Rect(gray.Cols()/4, gray.Rows()/4, 3*gray.Cols()/4, 3*gray.Rows()/4)
	centerRegion := gray.Region(centerRect)
	defer centerRegion.Close()
	
	edges := gocv.NewMat()
	defer edges.Close()
	gocv.Canny(centerRegion, &edges, 50, 150)
	
	// Find center of mass of edge points
	var sumX, sumY, count float64
	for y := 0; y < edges.Rows(); y++ {
		for x := 0; x < edges.Cols(); x++ {
			if edges.GetUCharAt(y, x) > 0 {
				sumX += float64(x + gray.Cols()/4) // Adjust for region offset
				sumY += float64(y + gray.Rows()/4)
				count++
			}
		}
	}
	
	if count > 100 { // Minimum edge points for valid grille
		return models.Point2D{X: sumX / count, Y: sumY / count}
	}
	
	return models.Point2D{X: 0, Y: 0}
}

func (ge *GeometricExtractor) detectRearBumperLine(img gocv.Mat) models.Point2D {
	gray := gocv.NewMat()
	defer gray.Close()
	
	if img.Channels() > 1 {
		gocv.CvtColor(img, &gray, gocv.ColorBGRToGray)
	} else {
		gray = img.Clone()
	}
	
	// Focus on lower half for bumper line
	lowerRect := image.Rect(0, gray.Rows()/2, gray.Cols(), gray.Rows())
	lowerHalf := gray.Region(lowerRect)
	defer lowerHalf.Close()
	
	edges := gocv.NewMat()
	defer edges.Close()
	gocv.Canny(lowerHalf, &edges, 50, 150)
	
	lines := gocv.NewMat()
	defer lines.Close()
	gocv.HoughLinesP(edges, &lines, 1, math.Pi/180, 40)
	
	// Find the longest horizontal line
	var bestLine models.Point2D
	maxLength := 0.0
	
	for i := 0; i < lines.Rows(); i++ {
		x1 := lines.GetFloatAt(i, 0)
		y1 := lines.GetFloatAt(i, 1)
		x2 := lines.GetFloatAt(i, 2)
		y2 := lines.GetFloatAt(i, 3)
		
		// Check if line is approximately horizontal
		if math.Abs(float64(y2-y1)) < 15 {
			length := math.Abs(float64(x2 - x1))
			if length > maxLength {
				maxLength = length
				centerX := (x1 + x2) / 2
				centerY := (y1 + y2) / 2 + float32(gray.Rows()/2) // Adjust for region offset
				bestLine = models.Point2D{X: float64(centerX), Y: float64(centerY)}
			}
		}
	}
	
	return bestLine
}

func (ge *GeometricExtractor) extractReferencePoints(img gocv.Mat, view models.VehicleView) []models.Point2D {
	points := []models.Point2D{}
	
	// Extract key reference points for image alignment
	switch view {
	case models.ViewFront:
		// Front reference points: headlight centers, grille center, bumper corners
		points = append(points, ge.findFrontReferencePoints(img)...)
	case models.ViewRear:
		// Rear reference points: taillight centers, license plate center, bumper corners
		points = append(points, ge.findRearReferencePoints(img)...)
	}
	
	return points
}

func (ge *GeometricExtractor) findFrontReferencePoints(img gocv.Mat) []models.Point2D {
	points := []models.Point2D{}
	
	// Add headlight centers
	headlights := ge.detectHeadlightRegions(img)
	points = append(points, headlights...)
	
	// Add grille center
	grilleCenter := ge.detectGrilleCenter(img)
	if grilleCenter.X > 0 && grilleCenter.Y > 0 {
		points = append(points, grilleCenter)
	}
	
	// Add image corners as reference
	points = append(points, models.Point2D{X: 0, Y: 0})
	points = append(points, models.Point2D{X: float64(img.Cols()), Y: 0})
	points = append(points, models.Point2D{X: 0, Y: float64(img.Rows())})
	points = append(points, models.Point2D{X: float64(img.Cols()), Y: float64(img.Rows())})
	
	return points
}

func (ge *GeometricExtractor) findRearReferencePoints(img gocv.Mat) []models.Point2D {
	points := []models.Point2D{}
	
	// Add taillight centers
	taillights := ge.detectTaillightRegions(img)
	points = append(points, taillights...)
	
	// Add bumper line center
	bumperLine := ge.detectRearBumperLine(img)
	if bumperLine.X > 0 && bumperLine.Y > 0 {
		points = append(points, bumperLine)
	}
	
	// Add image corners as reference
	points = append(points, models.Point2D{X: 0, Y: 0})
	points = append(points, models.Point2D{X: float64(img.Cols()), Y: 0})
	points = append(points, models.Point2D{X: 0, Y: float64(img.Rows())})
	points = append(points, models.Point2D{X: float64(img.Cols()), Y: float64(img.Rows())})
	
	return points
}

func (ge *GeometricExtractor) findMostCommonValue(values []int) int {
	if len(values) == 0 {
		return 0
	}
	
	frequency := make(map[int]int)
	for _, value := range values {
		frequency[value]++
	}
	
	maxCount := 0
	mostCommon := values[0]
	
	for value, count := range frequency {
		if count > maxCount {
			maxCount = count
			mostCommon = value
		}
	}
	
	return mostCommon
}