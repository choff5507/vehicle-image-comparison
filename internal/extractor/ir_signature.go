package extractor

import (
	"vehicle-comparison/internal/models"
	"gocv.io/x/gocv"
	"image"
	"math"
)

type IRSignatureExtractor struct {
	plateExtractor *LicensePlateExtractor
}

func NewIRSignatureExtractor() *IRSignatureExtractor {
	return &IRSignatureExtractor{
		plateExtractor: NewLicensePlateExtractor(),
	}
}

// ExtractIRSignature extracts IR reflectivity signature around license plate
func (irse *IRSignatureExtractor) ExtractIRSignature(img gocv.Mat) (*models.IRSignature, error) {
	// First detect the license plate
	plateRegion, err := irse.plateExtractor.DetectLicensePlate(img)
	if err != nil {
		return nil, err
	}
	
	// Convert to grayscale if needed
	gray := gocv.NewMat()
	defer gray.Close()
	
	if img.Channels() == 3 {
		gocv.CvtColor(img, &gray, gocv.ColorBGRToGray)
	} else {
		gray = img.Clone()
	}
	
	// Define surrounding area around license plate
	surroundingRegion := irse.calculateSurroundingRegion(plateRegion.Bounds, gray.Rows(), gray.Cols())
	
	// Extract various signature components
	signature := &models.IRSignature{
		PlateRegion:        *plateRegion,
		SurroundingRegion:  surroundingRegion,
		ReflectivityMap:    irse.extractReflectivityMap(gray, surroundingRegion, plateRegion.Bounds),
		MaterialSignature:  irse.extractMaterialSignature(gray, surroundingRegion, plateRegion.Bounds),
		IlluminationGradient: irse.extractIlluminationGradient(gray, surroundingRegion, plateRegion.Bounds),
		ShadowPatterns:     irse.extractShadowPatterns(gray, surroundingRegion, plateRegion.Bounds),
		TextureFeatures:    irse.extractTextureFeatures(gray, surroundingRegion, plateRegion.Bounds),
	}
	
	return signature, nil
}

func (irse *IRSignatureExtractor) calculateSurroundingRegion(plateBounds models.Bounds, imgHeight, imgWidth int) models.Bounds {
	// Create region 1.5x the plate size in each direction
	expandX := int(float64(plateBounds.Width) * 0.75)
	expandY := int(float64(plateBounds.Height) * 0.75)
	
	x := math.Max(0, float64(plateBounds.X-expandX))
	y := math.Max(0, float64(plateBounds.Y-expandY))
	width := math.Min(float64(imgWidth-int(x)), float64(plateBounds.Width+2*expandX))
	height := math.Min(float64(imgHeight-int(y)), float64(plateBounds.Height+2*expandY))
	
	return models.Bounds{
		X:      int(x),
		Y:      int(y),
		Width:  int(width),
		Height: int(height),
	}
}

func (irse *IRSignatureExtractor) extractReflectivityMap(gray gocv.Mat, surroundingRegion models.Bounds, plateBounds models.Bounds) [][]float64 {
	// Extract surrounding region
	surroundingRect := image.Rect(surroundingRegion.X, surroundingRegion.Y, 
		surroundingRegion.X+surroundingRegion.Width, surroundingRegion.Y+surroundingRegion.Height)
	roi := gray.Region(surroundingRect)
	defer roi.Close()
	
	// Create mask to exclude license plate area
	mask := gocv.NewMatWithSize(roi.Rows(), roi.Cols(), gocv.MatTypeCV8UC1)
	defer mask.Close()
	mask.SetTo(gocv.NewScalar(255, 255, 255, 255)) // White (include)
	
	// Set license plate area to black (exclude)
	plateX := plateBounds.X - surroundingRegion.X
	plateY := plateBounds.Y - surroundingRegion.Y
	if plateX >= 0 && plateY >= 0 && plateX+plateBounds.Width <= roi.Cols() && plateY+plateBounds.Height <= roi.Rows() {
		plateRect := image.Rect(plateX, plateY, plateX+plateBounds.Width, plateY+plateBounds.Height)
		plateROI := mask.Region(plateRect)
		plateROI.SetTo(gocv.NewScalar(0, 0, 0, 0)) // Black (exclude)
		plateROI.Close()
	}
	
	// Divide into grid and calculate average reflectivity for each cell
	gridSize := 8
	cellWidth := roi.Cols() / gridSize
	cellHeight := roi.Rows() / gridSize
	
	reflectivityMap := make([][]float64, gridSize)
	for i := range reflectivityMap {
		reflectivityMap[i] = make([]float64, gridSize)
	}
	
	for i := 0; i < gridSize; i++ {
		for j := 0; j < gridSize; j++ {
			cellX := j * cellWidth
			cellY := i * cellHeight
			cellW := cellWidth
			cellH := cellHeight
			
			// Ensure cell is within bounds
			if cellX+cellW > roi.Cols() {
				cellW = roi.Cols() - cellX
			}
			if cellY+cellH > roi.Rows() {
				cellH = roi.Rows() - cellY
			}
			
			if cellW > 0 && cellH > 0 {
				cellRect := image.Rect(cellX, cellY, cellX+cellW, cellY+cellH)
				cellROI := roi.Region(cellRect)
				cellMask := mask.Region(cellRect)
				
				// Calculate mean reflectivity for this cell (excluding plate area)
				mean := gocv.NewMat()
				defer mean.Close()
				gocv.MeanStdDev(cellROI, &mean, nil)
				reflectivityMap[i][j] = mean.GetDoubleAt(0, 0) / 255.0
				
				cellROI.Close()
				cellMask.Close()
			}
		}
	}
	
	return reflectivityMap
}

func (irse *IRSignatureExtractor) extractMaterialSignature(gray gocv.Mat, surroundingRegion models.Bounds, plateBounds models.Bounds) []float64 {
	// Extract different material response signatures
	signature := make([]float64, 6)
	
	surroundingRect := image.Rect(surroundingRegion.X, surroundingRegion.Y, 
		surroundingRegion.X+surroundingRegion.Width, surroundingRegion.Y+surroundingRegion.Height)
	roi := gray.Region(surroundingRect)
	defer roi.Close()
	
	// Apply different filters to detect material properties
	
	// 1. High reflectivity areas (metal, chrome)
	highReflThresh := gocv.NewMat()
	defer highReflThresh.Close()
	gocv.Threshold(roi, &highReflThresh, 200, 255, gocv.ThresholdBinary)
	signature[0] = float64(gocv.CountNonZero(highReflThresh)) / float64(roi.Rows()*roi.Cols())
	
	// 2. Medium reflectivity areas (painted surfaces)
	medReflThresh := gocv.NewMat()
	defer medReflThresh.Close()
	gocv.Threshold(roi, &medReflThresh, 100, 200, gocv.ThresholdBinaryInv)
	gocv.Threshold(medReflThresh, &medReflThresh, 100, 255, gocv.ThresholdBinary)
	signature[1] = float64(gocv.CountNonZero(medReflThresh)) / float64(roi.Rows()*roi.Cols())
	
	// 3. Low reflectivity areas (plastic, rubber)
	lowReflThresh := gocv.NewMat()
	defer lowReflThresh.Close()
	gocv.Threshold(roi, &lowReflThresh, 80, 255, gocv.ThresholdBinaryInv)
	signature[2] = float64(gocv.CountNonZero(lowReflThresh)) / float64(roi.Rows()*roi.Cols())
	
	// 4. Edge density (material transitions)
	edges := gocv.NewMat()
	defer edges.Close()
	gocv.Canny(roi, &edges, 50, 150)
	signature[3] = float64(gocv.CountNonZero(edges)) / float64(roi.Rows()*roi.Cols())
	
	// 5. Texture variance (surface roughness)
	laplacian := gocv.NewMat()
	defer laplacian.Close()
	gocv.Laplacian(roi, &laplacian, gocv.MatTypeCV64F, 1, 1, 0, gocv.BorderDefault)
	
	meanMat := gocv.NewMat()
	stddevMat2 := gocv.NewMat()
	defer meanMat.Close()
	defer stddevMat2.Close()
	gocv.MeanStdDev(laplacian, &meanMat, &stddevMat2)
	signature[4] = stddevMat2.GetDoubleAt(0, 0) / 255.0
	
	// 6. Overall brightness variation
	stddevMat3 := gocv.NewMat()
	meanMat2 := gocv.NewMat()
	defer stddevMat3.Close()
	defer meanMat2.Close()
	gocv.MeanStdDev(roi, &meanMat2, &stddevMat3)
	signature[5] = stddevMat3.GetDoubleAt(0, 0) / 255.0
	
	return signature
}

func (irse *IRSignatureExtractor) extractIlluminationGradient(gray gocv.Mat, surroundingRegion models.Bounds, plateBounds models.Bounds) []float64 {
	// Analyze how illumination falls off from IR source
	surroundingRect := image.Rect(surroundingRegion.X, surroundingRegion.Y, 
		surroundingRegion.X+surroundingRegion.Width, surroundingRegion.Y+surroundingRegion.Height)
	roi := gray.Region(surroundingRect)
	defer roi.Close()
	
	gradient := make([]float64, 4) // Top, Right, Bottom, Left gradients from plate
	
	// Calculate plate center relative to surrounding region
	plateCenterX := (plateBounds.X + plateBounds.Width/2) - surroundingRegion.X
	plateCenterY := (plateBounds.Y + plateBounds.Height/2) - surroundingRegion.Y
	
	// Sample brightness at different distances and directions from plate
	distances := []int{10, 20, 30}
	
	for dir := 0; dir < 4; dir++ {
		var totalGradient float64
		var samples int
		
		for _, dist := range distances {
			var x, y int
			switch dir {
			case 0: // Top
				x, y = plateCenterX, plateCenterY-dist
			case 1: // Right
				x, y = plateCenterX+dist, plateCenterY
			case 2: // Bottom
				x, y = plateCenterX, plateCenterY+dist
			case 3: // Left
				x, y = plateCenterX-dist, plateCenterY
			}
			
			if x >= 0 && x < roi.Cols() && y >= 0 && y < roi.Rows() {
				brightness := roi.GetUCharAt(y, x)
				totalGradient += float64(brightness) / 255.0
				samples++
			}
		}
		
		if samples > 0 {
			gradient[dir] = totalGradient / float64(samples)
		}
	}
	
	return gradient
}

func (irse *IRSignatureExtractor) extractShadowPatterns(gray gocv.Mat, surroundingRegion models.Bounds, plateBounds models.Bounds) []models.Point2D {
	// Find shadow/dark regions that indicate 3D structure
	surroundingRect := image.Rect(surroundingRegion.X, surroundingRegion.Y, 
		surroundingRegion.X+surroundingRegion.Width, surroundingRegion.Y+surroundingRegion.Height)
	roi := gray.Region(surroundingRect)
	defer roi.Close()
	
	// Find dark regions (shadows)
	shadows := gocv.NewMat()
	defer shadows.Close()
	gocv.Threshold(roi, &shadows, 60, 255, gocv.ThresholdBinaryInv)
	
	// Find contours of shadow regions
	contours := gocv.FindContours(shadows, gocv.RetrievalExternal, gocv.ChainApproxSimple)
	
	var shadowCenters []models.Point2D
	for i := 0; i < contours.Size(); i++ {
		contour := contours.At(i)
		area := gocv.ContourArea(contour)
		
		// Only consider significant shadow regions
		if area > 50 {
			// Calculate centroid using bounding rect as approximation
			rect := gocv.BoundingRect(contour)
			cx := float64(rect.Min.X + rect.Dx()/2)
			cy := float64(rect.Min.Y + rect.Dy()/2)
			
			// Convert back to image coordinates
			shadowCenters = append(shadowCenters, models.Point2D{
				X: cx + float64(surroundingRegion.X),
				Y: cy + float64(surroundingRegion.Y),
			})
		}
	}
	
	return shadowCenters
}

func (irse *IRSignatureExtractor) extractTextureFeatures(gray gocv.Mat, surroundingRegion models.Bounds, plateBounds models.Bounds) []float64 {
	// Extract texture features using Local Binary Patterns concept
	surroundingRect := image.Rect(surroundingRegion.X, surroundingRegion.Y, 
		surroundingRegion.X+surroundingRegion.Width, surroundingRegion.Y+surroundingRegion.Height)
	roi := gray.Region(surroundingRect)
	defer roi.Close()
	
	features := make([]float64, 4)
	
	// 1. Local variance (texture roughness)
	blurred := gocv.NewMat()
	defer blurred.Close()
	gocv.GaussianBlur(roi, &blurred, image.Pt(5, 5), 0, 0, gocv.BorderDefault)
	
	diff := gocv.NewMat()
	defer diff.Close()
	gocv.AbsDiff(roi, blurred, &diff)
	
	meanMat := gocv.NewMat()
	defer meanMat.Close()
	gocv.MeanStdDev(diff, &meanMat, nil)
	features[0] = meanMat.GetDoubleAt(0, 0) / 255.0
	
	// 2. Gradient magnitude (edge density)
	gradX := gocv.NewMat()
	gradY := gocv.NewMat()
	defer gradX.Close()
	defer gradY.Close()
	
	gocv.Sobel(roi, &gradX, gocv.MatTypeCV64F, 1, 0, 3, 1, 0, gocv.BorderDefault)
	gocv.Sobel(roi, &gradY, gocv.MatTypeCV64F, 0, 1, 3, 1, 0, gocv.BorderDefault)
	
	gradMag := gocv.NewMat()
	defer gradMag.Close()
	gocv.Magnitude(gradX, gradY, &gradMag)
	
	meanMat2 := gocv.NewMat()
	defer meanMat2.Close()
	gocv.MeanStdDev(gradMag, &meanMat2, nil)
	features[1] = meanMat2.GetDoubleAt(0, 0) / 255.0
	
	// 3. Directional texture (horizontal vs vertical patterns)
	meanMatX := gocv.NewMat()
	meanMatY := gocv.NewMat()
	defer meanMatX.Close()
	defer meanMatY.Close()
	gocv.MeanStdDev(gradX, &meanMatX, nil)
	horizontalGrad := math.Abs(meanMatX.GetDoubleAt(0, 0))
	gocv.MeanStdDev(gradY, &meanMatY, nil)
	verticalGrad := math.Abs(meanMatY.GetDoubleAt(0, 0))
	
	if horizontalGrad+verticalGrad > 0 {
		features[2] = horizontalGrad / (horizontalGrad + verticalGrad)
	}
	
	// 4. Pattern regularity (how uniform the texture is)
	hist := gocv.NewMat()
	defer hist.Close()
	mask := gocv.NewMat()
	defer mask.Close()
	
	channels := []int{0}
	histSize := []int{256}
	ranges := []float64{0, 256}
	
	gocv.CalcHist([]gocv.Mat{roi}, channels, mask, &hist, histSize, ranges, false)
	
	// Calculate histogram entropy as measure of regularity
	var entropy float64
	total := float64(roi.Rows() * roi.Cols())
	
	for i := 0; i < 256; i++ {
		count := float64(hist.GetFloatAt(i, 0))
		if count > 0 {
			p := count / total
			entropy -= p * math.Log2(p)
		}
	}
	
	features[3] = entropy / 8.0 // Normalize by max entropy
	
	return features
}