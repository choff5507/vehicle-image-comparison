package vehiclecompare

import (
	"vehicle-comparison/internal/models"
	"vehicle-comparison/internal/preprocessor"
	"vehicle-comparison/internal/extractor"
	"vehicle-comparison/internal/comparator"
	"gocv.io/x/gocv"
	"encoding/base64"
	"fmt"
	"time"
)

type VehicleComparisonService struct {
	qualityAssessor        *preprocessor.QualityAssessor
	viewLightingClassifier *preprocessor.ViewLightingClassifier
	geometricExtractor     *extractor.GeometricExtractor
	lightPatternExtractor  *extractor.LightPatternExtractor
	comparisonEngine       *comparator.ComparisonEngine
}

func NewVehicleComparisonService() *VehicleComparisonService {
	return &VehicleComparisonService{
		qualityAssessor:        preprocessor.NewQualityAssessor(),
		viewLightingClassifier: preprocessor.NewViewLightingClassifier(),
		geometricExtractor:     extractor.NewGeometricExtractor(),
		lightPatternExtractor:  extractor.NewLightPatternExtractor(),
		comparisonEngine:       comparator.NewComparisonEngine(),
	}
}

// CompareVehicleImages is the main entry point for vehicle comparison from file paths
func (vcs *VehicleComparisonService) CompareVehicleImages(image1Path, image2Path string) (*models.ComparisonResult, error) {
	startTime := time.Now()
	
	// Load images
	img1 := gocv.IMRead(image1Path, gocv.IMReadColor)
	img2 := gocv.IMRead(image2Path, gocv.IMReadColor)
	defer img1.Close()
	defer img2.Close()
	
	if img1.Empty() || img2.Empty() {
		return nil, fmt.Errorf("failed to load one or both images")
	}
	
	return vcs.compareImages(img1, img2, startTime)
}

// CompareVehicleImagesFromBase64 compares images from base64 encoded strings
func (vcs *VehicleComparisonService) CompareVehicleImagesFromBase64(image1Base64, image2Base64 string) (*models.ComparisonResult, error) {
	startTime := time.Now()
	
	// Decode base64 images
	img1Data, err := base64.StdEncoding.DecodeString(image1Base64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image1 base64: %v", err)
	}
	
	img2Data, err := base64.StdEncoding.DecodeString(image2Base64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image2 base64: %v", err)
	}
	
	// Create Mat from image data
	img1, err := gocv.IMDecode(img1Data, gocv.IMReadColor)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image1: %v", err)
	}
	img2, err := gocv.IMDecode(img2Data, gocv.IMReadColor)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image2: %v", err)
	}
	defer img1.Close()
	defer img2.Close()
	
	if img1.Empty() || img2.Empty() {
		return nil, fmt.Errorf("failed to decode one or both images")
	}
	
	return vcs.compareImages(img1, img2, startTime)
}

func (vcs *VehicleComparisonService) compareImages(img1, img2 gocv.Mat, startTime time.Time) (*models.ComparisonResult, error) {
	// Process both images
	vehicleImg1, err := vcs.processImage(img1)
	if err != nil {
		return nil, fmt.Errorf("failed to process image 1: %v", err)
	}
	defer vehicleImg1.Image.Close()
	
	vehicleImg2, err := vcs.processImage(img2)
	if err != nil {
		return nil, fmt.Errorf("failed to process image 2: %v", err)
	}
	defer vehicleImg2.Image.Close()
	
	// Validate consistency
	if err := vcs.validateImageConsistency(vehicleImg1, vehicleImg2); err != nil {
		return nil, err
	}
	
	// Extract features
	features1, err := vcs.extractFeatures(vehicleImg1)
	if err != nil {
		return nil, fmt.Errorf("failed to extract features from image 1: %v", err)
	}
	
	features2, err := vcs.extractFeatures(vehicleImg2)
	if err != nil {
		return nil, fmt.Errorf("failed to extract features from image 2: %v", err)
	}
	
	// Compare features
	result, err := vcs.comparisonEngine.CompareVehicles(features1, features2)
	if err != nil {
		return nil, fmt.Errorf("failed to compare vehicles: %v", err)
	}
	
	// Add processing information
	result.ProcessingInfo = models.ProcessingInfo{
		ProcessingTimeMs:    time.Since(startTime).Milliseconds(),
		Image1Quality:       vehicleImg1.QualityScore,
		Image2Quality:       vehicleImg2.QualityScore,
		ViewConsistency:     vehicleImg1.View == vehicleImg2.View,
		LightingConsistency: vehicleImg1.Lighting == vehicleImg2.Lighting,
	}
	
	return result, nil
}

func (vcs *VehicleComparisonService) processImage(img gocv.Mat) (*models.VehicleImage, error) {
	// Assess image quality
	quality, err := vcs.qualityAssessor.AssessImageQuality(img)
	if err != nil {
		return nil, err
	}
	
	if quality < 0.3 {
		return nil, fmt.Errorf("image quality too low: %f", quality)
	}
	
	// Classify view and lighting
	view, viewConfidence, err := vcs.viewLightingClassifier.ClassifyView(img)
	if err != nil {
		return nil, err
	}
	
	if viewConfidence < 0.5 {
		return nil, fmt.Errorf("unable to determine vehicle view with sufficient confidence: %f", viewConfidence)
	}
	
	lighting, lightingConfidence, err := vcs.viewLightingClassifier.ClassifyLighting(img)
	if err != nil {
		return nil, err
	}
	
	if lightingConfidence < 0.5 {
		return nil, fmt.Errorf("unable to determine lighting conditions with sufficient confidence: %f", lightingConfidence)
	}
	
	// For this implementation, we'll use the full image as the vehicle region
	// In a full implementation, you would use vehicle detection here
	croppedVehicle := img.Clone()
	bounds := models.Bounds{
		X: 0, Y: 0, 
		Width: img.Cols(), 
		Height: img.Rows(),
	}
	
	return &models.VehicleImage{
		Image:        croppedVehicle,
		View:         view,
		Lighting:     lighting,
		QualityScore: quality,
		ProcessingMeta: models.ProcessingMetadata{
			OriginalWidth:    img.Cols(),
			OriginalHeight:   img.Rows(),
			VehicleBounds:    bounds,
			NormalizedWidth:  croppedVehicle.Cols(),
			NormalizedHeight: croppedVehicle.Rows(),
		},
	}, nil
}

func (vcs *VehicleComparisonService) validateImageConsistency(img1, img2 *models.VehicleImage) error {
	if img1.View != img2.View {
		return fmt.Errorf("vehicle views do not match: %v vs %v", img1.View, img2.View)
	}
	
	if img1.Lighting != img2.Lighting {
		return fmt.Errorf("lighting conditions do not match: %v vs %v", img1.Lighting, img2.Lighting)
	}
	
	if img1.QualityScore < 0.5 || img2.QualityScore < 0.5 {
		return fmt.Errorf("one or both images have insufficient quality")
	}
	
	return nil
}

func (vcs *VehicleComparisonService) extractFeatures(vehicleImg *models.VehicleImage) (models.VehicleFeatures, error) {
	features := models.VehicleFeatures{
		View:     vehicleImg.View,
		Lighting: vehicleImg.Lighting,
	}
	
	// Extract geometric features (universal)
	geometricFeatures, err := vcs.geometricExtractor.ExtractGeometricFeatures(vehicleImg.Image, vehicleImg.View)
	if err != nil {
		return features, err
	}
	features.GeometricFeatures = geometricFeatures
	
	// Extract light patterns
	lightPatterns, err := vcs.lightPatternExtractor.ExtractLightPatterns(vehicleImg.Image, vehicleImg.View, vehicleImg.Lighting)
	if err != nil {
		return features, err
	}
	features.LightPatterns = lightPatterns
	
	// Extract bumper features (simplified implementation)
	features.BumperFeatures = vcs.extractBumperFeatures(vehicleImg.Image)
	
	// Extract lighting-specific features
	if vehicleImg.Lighting == models.LightingDaylight {
		// Extract daylight-specific features (simplified)
		features.DaylightFeatures = vcs.extractDaylightFeatures(vehicleImg.Image)
	} else if vehicleImg.Lighting == models.LightingInfrared {
		// Extract infrared-specific features (simplified)
		features.InfraredFeatures = vcs.extractInfraredFeatures(vehicleImg.Image)
	}
	
	// Calculate extraction quality
	features.ExtractionQuality = vcs.calculateExtractionQuality(features)
	
	return features, nil
}

func (vcs *VehicleComparisonService) extractBumperFeatures(img gocv.Mat) models.BumperFeatures {
	// Simplified bumper feature extraction
	return models.BumperFeatures{
		ContourSignature: []models.Point2D{},
		TextureFeatures:  []float64{0.5, 0.3, 0.7}, // Placeholder
		MountingPoints:   []models.Point2D{},
		LicensePlateArea: models.Bounds{},
	}
}

func (vcs *VehicleComparisonService) extractDaylightFeatures(img gocv.Mat) *models.DaylightFeatures {
	// Simplified daylight feature extraction
	return &models.DaylightFeatures{
		ColorProfile: models.ColorProfile{
			DominantColors: []models.Color{
				{R: 128, G: 128, B: 128, Weight: 0.5},
				{R: 64, G: 64, B: 64, Weight: 0.3},
			},
			Histogram: make([]int, 256),
		},
		BadgeLocations:  []models.BadgeFeature{},
		TrimDetails:     []models.TrimFeature{},
		SurfaceTexture: models.TextureSignature{
			Features: []float64{0.4, 0.6, 0.2},
			Type:     "basic",
		},
	}
}

func (vcs *VehicleComparisonService) extractInfraredFeatures(img gocv.Mat) *models.InfraredFeatures {
	// Simplified infrared feature extraction
	return &models.InfraredFeatures{
		ThermalSignature:   []float64{0.3, 0.7, 0.5},
		ReflectiveElements: []models.ReflectiveElement{},
		HeatPatterns:       []models.HeatPattern{},
		MaterialSignature:  []float64{0.6, 0.4, 0.8},
	}
}

func (vcs *VehicleComparisonService) calculateExtractionQuality(features models.VehicleFeatures) float64 {
	quality := 0.0
	components := 0
	
	// Geometric features quality
	if len(features.GeometricFeatures.ReferencePoints) > 0 {
		quality += 0.8
	}
	components++
	
	// Light pattern quality
	if len(features.LightPatterns.LightElements) > 0 {
		quality += 0.9
	}
	components++
	
	// Lighting-specific features quality
	if features.DaylightFeatures != nil || features.InfraredFeatures != nil {
		quality += 0.7
	}
	components++
	
	if components == 0 {
		return 0.0
	}
	
	return quality / float64(components)
}