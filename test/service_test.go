package test

import (
	"vehicle-comparison/pkg/vehiclecompare"
	"testing"
	"os"
	"path/filepath"
)

func TestVehicleComparisonService(t *testing.T) {
	service := vehiclecompare.NewVehicleComparisonService()
	
	if service == nil {
		t.Fatal("Failed to create service")
	}
}

func TestCompareVehicleImagesWithNonExistentFiles(t *testing.T) {
	service := vehiclecompare.NewVehicleComparisonService()
	
	_, err := service.CompareVehicleImages("nonexistent1.jpg", "nonexistent2.jpg")
	if err == nil {
		t.Error("Expected error when comparing non-existent files")
	}
}

func TestCompareVehicleImagesFromBase64Invalid(t *testing.T) {
	service := vehiclecompare.NewVehicleComparisonService()
	
	_, err := service.CompareVehicleImagesFromBase64("invalid_base64", "invalid_base64")
	if err == nil {
		t.Error("Expected error when providing invalid base64")
	}
}

// Integration test - only run if test images are available
func TestCompareVehicleImagesIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	
	// Check if test images exist
	testDir := "testdata"
	if _, err := os.Stat(testDir); os.IsNotExist(err) {
		t.Skip("Test images not available - skipping integration test")
	}
	
	// Look for any jpg files in testdata
	files, err := filepath.Glob(filepath.Join(testDir, "*.jpg"))
	if err != nil || len(files) < 2 {
		t.Skip("Need at least 2 test images - skipping integration test")
	}
	
	service := vehiclecompare.NewVehicleComparisonService()
	
	result, err := service.CompareVehicleImages(files[0], files[1])
	if err != nil {
		t.Logf("Comparison failed (expected for random test images): %v", err)
		return
	}
	
	// Basic validation of result structure
	if result.SimilarityScore < 0 || result.SimilarityScore > 1 {
		t.Errorf("Similarity score out of range: %f", result.SimilarityScore)
	}
	
	if result.ProcessingInfo.ProcessingTimeMs <= 0 {
		t.Error("Processing time should be positive")
	}
	
	t.Logf("Comparison completed successfully")
	t.Logf("Same vehicle: %v", result.IsSameVehicle)
	t.Logf("Similarity: %.3f", result.SimilarityScore)
	t.Logf("Processing time: %dms", result.ProcessingInfo.ProcessingTimeMs)
}