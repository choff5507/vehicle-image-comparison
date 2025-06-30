package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/choff5507/vehicle-image-comparison/internal/models"
	"github.com/choff5507/vehicle-image-comparison/pkg/vehiclecompare"
)

func main() {
	// Check if we have the required arguments
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run example/main.go <image1.jpg> <image2.jpg>")
		fmt.Println("Example: go run example/main.go testdata/car1.jpg testdata/car2.jpg")
		os.Exit(1)
	}

	image1Path := os.Args[1]
	image2Path := os.Args[2]

	// Create the vehicle comparison service
	service := vehiclecompare.NewVehicleComparisonService()
	
	fmt.Printf("Comparing vehicles:\n")
	fmt.Printf("  Image 1: %s\n", image1Path)
	fmt.Printf("  Image 2: %s\n", image2Path)
	fmt.Printf("  Processing...\n\n")

	// Compare the vehicle images
	result, err := service.CompareVehicleImages(image1Path, image2Path)
	if err != nil {
		log.Fatalf("Error comparing images: %v", err)
	}

	// Print basic results
	fmt.Printf("=== COMPARISON RESULTS ===\n")
	fmt.Printf("Same Vehicle: %v\n", result.IsSameVehicle)
	fmt.Printf("Similarity Score: %.3f\n", result.SimilarityScore)
	fmt.Printf("Confidence Level: %v\n", getConfidenceString(result.ConfidenceLevel))
	fmt.Printf("Processing Time: %dms\n\n", result.ProcessingInfo.ProcessingTimeMs)

	// Print detailed breakdown
	fmt.Printf("=== DETAILED SCORES ===\n")
	fmt.Printf("Geometric Similarity: %.3f\n", result.DetailedScores.GeometricSimilarity)
	fmt.Printf("Light Pattern Similarity: %.3f\n", result.DetailedScores.LightPatternSimilarity)
	fmt.Printf("Bumper Similarity: %.3f\n", result.DetailedScores.BumperSimilarity)
	
	if result.DetailedScores.ColorSimilarity > 0 {
		fmt.Printf("Color Similarity: %.3f\n", result.DetailedScores.ColorSimilarity)
	}
	
	if result.DetailedScores.ThermalSimilarity > 0 {
		fmt.Printf("Thermal/IR Similarity: %.3f\n", result.DetailedScores.ThermalSimilarity)
	}
	
	fmt.Printf("\n=== IMAGE QUALITY ===\n")
	fmt.Printf("Image 1 Quality: %.3f\n", result.ProcessingInfo.Image1Quality)
	fmt.Printf("Image 2 Quality: %.3f\n", result.ProcessingInfo.Image2Quality)
	fmt.Printf("View Consistency: %v\n", result.ProcessingInfo.ViewConsistency)
	fmt.Printf("Lighting Consistency: %v\n", result.ProcessingInfo.LightingConsistency)

	// Print fraud detection analysis
	fmt.Printf("\n=== FRAUD DETECTION ANALYSIS ===\n")
	if result.IsSameVehicle {
		fmt.Printf("✅ No fraud detected - Images appear to show the same vehicle\n")
	} else {
		if result.ProcessingInfo.ViewConsistency && result.ProcessingInfo.LightingConsistency {
			fmt.Printf("⚠️  POTENTIAL FRAUD: Different vehicles with same lighting/view conditions\n")
			fmt.Printf("   This could indicate license plate swapping between similar vehicles\n")
		} else {
			fmt.Printf("❓ Different vehicles detected (expected due to view/lighting differences)\n")
		}
	}

	// Demonstrate base64 functionality
	fmt.Printf("\n=== BASE64 EXAMPLE ===\n")
	demonstrateBase64Usage(service, image1Path, image2Path)

	// Export full JSON result
	fmt.Printf("\n=== FULL JSON RESULT ===\n")
	jsonResult, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		log.Printf("Error marshaling JSON: %v", err)
	} else {
		fmt.Printf("%s\n", jsonResult)
	}
}

func demonstrateBase64Usage(service *vehiclecompare.VehicleComparisonService, image1Path, image2Path string) {
	// Read and encode first image
	image1Data, err := os.ReadFile(image1Path)
	if err != nil {
		fmt.Printf("Could not read image1 for base64 demo: %v\n", err)
		return
	}

	image2Data, err := os.ReadFile(image2Path)
	if err != nil {
		fmt.Printf("Could not read image2 for base64 demo: %v\n", err)
		return
	}

	// Convert to base64
	base64Image1 := base64.StdEncoding.EncodeToString(image1Data)
	base64Image2 := base64.StdEncoding.EncodeToString(image2Data)

	fmt.Printf("Demonstrating base64 comparison...\n")
	
	// Compare using base64
	result, err := service.CompareVehicleImagesFromBase64(base64Image1, base64Image2)
	if err != nil {
		fmt.Printf("Base64 comparison failed: %v\n", err)
		return
	}

	fmt.Printf("Base64 comparison result: Same vehicle = %v, Similarity = %.3f\n", 
		result.IsSameVehicle, result.SimilarityScore)
}

func getConfidenceString(level models.ConfidenceLevel) string {
	switch level {
	case models.ConfidenceHigh:
		return "High"
	case models.ConfidenceMedium:
		return "Medium"
	case models.ConfidenceLow:
		return "Low"
	default:
		return "Unknown"
	}
}