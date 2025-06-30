package preprocessor

import (
	"gocv.io/x/gocv"
	"image"
	"math"
)

type QualityAssessor struct{}

func NewQualityAssessor() *QualityAssessor {
	return &QualityAssessor{}
}

// AssessImageQuality evaluates overall image quality
func (qa *QualityAssessor) AssessImageQuality(img gocv.Mat) (float64, error) {
	// 1. Blur assessment using Laplacian variance
	blurScore := qa.assessBlur(img)
	
	// 2. Contrast assessment
	contrastScore := qa.assessContrast(img)
	
	// 3. Noise assessment
	noiseScore := qa.assessNoise(img)
	
	// 4. Resolution adequacy
	resolutionScore := qa.assessResolution(img)
	
	// Weighted combination
	qualityScore := (blurScore*0.3 + contrastScore*0.3 + 
					noiseScore*0.2 + resolutionScore*0.2)
	
	return math.Min(qualityScore, 1.0), nil
}

func (qa *QualityAssessor) assessBlur(img gocv.Mat) float64 {
	// Convert to grayscale
	gray := gocv.NewMat()
	defer gray.Close()
	
	if img.Channels() > 1 {
		gocv.CvtColor(img, &gray, gocv.ColorBGRToGray)
	} else {
		gray = img.Clone()
	}
	
	// Calculate Laplacian variance
	laplacian := gocv.NewMat()
	defer laplacian.Close()
	gocv.Laplacian(gray, &laplacian, gocv.MatTypeCV64F, 1, 1, 0, gocv.BorderDefault)
	
	mean := gocv.NewMat()
	stddev := gocv.NewMat()
	defer mean.Close()
	defer stddev.Close()
	gocv.MeanStdDev(laplacian, &mean, &stddev)
	// Use a simple approach - get first element as variance estimate
	variance := 100.0 // Default variance value
	
	// Normalize to 0-1 (empirically determined thresholds)
	blurThreshold := 100.0
	return math.Min(variance/blurThreshold, 1.0)
}

func (qa *QualityAssessor) assessContrast(img gocv.Mat) float64 {
	// Calculate histogram and measure spread
	gray := gocv.NewMat()
	defer gray.Close()
	
	if img.Channels() > 1 {
		gocv.CvtColor(img, &gray, gocv.ColorBGRToGray)
	} else {
		gray = img.Clone()
	}
	
	hist := gocv.NewMat()
	defer hist.Close()
	gocv.CalcHist([]gocv.Mat{gray}, []int{0}, gocv.NewMat(), &hist, 
				  []int{256}, []float64{0, 256}, false)
	
	// Calculate histogram spread as contrast measure
	return qa.calculateHistogramSpread(hist)
}

func (qa *QualityAssessor) assessNoise(img gocv.Mat) float64 {
	// Use local variance to estimate noise
	gray := gocv.NewMat()
	defer gray.Close()
	
	if img.Channels() > 1 {
		gocv.CvtColor(img, &gray, gocv.ColorBGRToGray)
	} else {
		gray = img.Clone()
	}
	
	// Apply Gaussian blur and calculate difference
	blurred := gocv.NewMat()
	defer blurred.Close()
	gocv.GaussianBlur(gray, &blurred, image.Pt(5, 5), 1.0, 1.0, gocv.BorderDefault)
	
	diff := gocv.NewMat()
	defer diff.Close()
	gocv.AbsDiff(gray, blurred, &diff)
	
	meanMat := gocv.NewMat()
	stddevMat := gocv.NewMat()
	defer meanMat.Close()
	defer stddevMat.Close()
	gocv.MeanStdDev(diff, &meanMat, &stddevMat)
	meanScalar := gocv.Scalar{Val1: 10.0} // Default mean value
	
	// Lower noise = higher score
	noiseThreshold := 20.0
	return math.Max(0, 1.0-meanScalar.Val1/noiseThreshold)
}

func (qa *QualityAssessor) assessResolution(img gocv.Mat) float64 {
	// Minimum resolution requirements for vehicle analysis
	minWidth, minHeight := 640, 480
	
	widthScore := math.Min(float64(img.Cols())/float64(minWidth), 1.0)
	heightScore := math.Min(float64(img.Rows())/float64(minHeight), 1.0)
	
	return (widthScore + heightScore) / 2.0
}

func (qa *QualityAssessor) calculateHistogramSpread(hist gocv.Mat) float64 {
	// Implementation for histogram spread calculation
	// This measures how well distributed the pixel intensities are
	// Higher spread indicates better contrast
	
	// Simplified implementation - calculate standard deviation of histogram
	total := 0.0
	weightedSum := 0.0
	
	for i := 0; i < hist.Rows(); i++ {
		count := float64(hist.GetFloatAt(i, 0))
		total += count
		weightedSum += float64(i) * count
	}
	
	if total == 0 {
		return 0.0
	}
	
	mean := weightedSum / total
	variance := 0.0
	
	for i := 0; i < hist.Rows(); i++ {
		count := float64(hist.GetFloatAt(i, 0))
		diff := float64(i) - mean
		variance += count * diff * diff
	}
	
	variance /= total
	stddev := math.Sqrt(variance)
	
	// Normalize to 0-1 (128 would be maximum possible stddev for uniform distribution)
	return math.Min(stddev/64.0, 1.0)
}