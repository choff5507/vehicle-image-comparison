package main

import (
	"github.com/choff5507/vehicle-image-comparison/pkg/vehiclecompare"
	"github.com/choff5507/vehicle-image-comparison/internal/models"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	var (
		image1Path   = flag.String("image1", "", "Path to first vehicle image")
		image2Path   = flag.String("image2", "", "Path to second vehicle image")
		image1Base64 = flag.String("image1-base64", "", "Base64 encoded first vehicle image")
		image2Base64 = flag.String("image2-base64", "", "Base64 encoded second vehicle image")
		outputPath   = flag.String("output", "", "Path to output JSON file (optional)")
		verbose      = flag.Bool("verbose", false, "Enable verbose output")
	)
	flag.Parse()
	
	// Validate input parameters
	hasFilePaths := *image1Path != "" && *image2Path != ""
	hasBase64 := *image1Base64 != "" && *image2Base64 != ""
	
	if !hasFilePaths && !hasBase64 {
		fmt.Println("Usage:")
		fmt.Println("  For file inputs:")
		fmt.Println("    ./vehicle-compare -image1 <path> -image2 <path> [-output <path>] [-verbose]")
		fmt.Println("  For base64 inputs:")
		fmt.Println("    ./vehicle-compare -image1-base64 <base64> -image2-base64 <base64> [-output <path>] [-verbose]")
		os.Exit(1)
	}
	
	if hasFilePaths && hasBase64 {
		log.Fatal("Cannot specify both file paths and base64 inputs")
	}
	
	// Initialize the vehicle comparison service
	service := vehiclecompare.NewVehicleComparisonService()
	
	// Compare the vehicles
	var result *models.ComparisonResult
	var err error
	
	if hasFilePaths {
		if *verbose {
			fmt.Printf("Comparing images: %s vs %s\n", *image1Path, *image2Path)
		}
		result, err = service.CompareVehicleImages(*image1Path, *image2Path)
	} else {
		if *verbose {
			fmt.Println("Comparing base64 encoded images")
		}
		result, err = service.CompareVehicleImagesFromBase64(*image1Base64, *image2Base64)
	}
	
	if err != nil {
		log.Fatalf("Comparison failed: %v", err)
	}
	
	// Validate and sanitize result to prevent NaN/Inf JSON marshaling errors
	result.ValidateAndSanitize()
	
	// Output results
	resultJSON, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal result: %v", err)
	}
	
	if *outputPath != "" {
		err = os.WriteFile(*outputPath, resultJSON, 0644)
		if err != nil {
			log.Fatalf("Failed to write output file: %v", err)
		}
		fmt.Printf("Results written to %s\n", *outputPath)
	}
	
	// Print summary to console
	fmt.Printf("Vehicle Comparison Results:\n")
	fmt.Printf("==========================\n")
	fmt.Printf("Same Vehicle: %v\n", result.IsSameVehicle)
	fmt.Printf("Similarity Score: %.3f\n", result.SimilarityScore)
	fmt.Printf("Confidence: %v\n", getConfidenceString(result.ConfidenceLevel))
	fmt.Printf("Processing Time: %dms\n", result.ProcessingInfo.ProcessingTimeMs)
	
	if *verbose {
		fmt.Printf("\nDetailed Scores:\n")
		fmt.Printf("  Geometric: %.3f\n", result.DetailedScores.GeometricSimilarity)
		fmt.Printf("  Light Pattern: %.3f\n", result.DetailedScores.LightPatternSimilarity)
		fmt.Printf("  Bumper: %.3f\n", result.DetailedScores.BumperSimilarity)
		
		if result.DetailedScores.ColorSimilarity > 0 {
			fmt.Printf("  Color: %.3f\n", result.DetailedScores.ColorSimilarity)
		}
		
		if result.DetailedScores.ThermalSimilarity > 0 {
			fmt.Printf("  Thermal: %.3f\n", result.DetailedScores.ThermalSimilarity)
		}
		
		fmt.Printf("\nProcessing Info:\n")
		fmt.Printf("  Image 1 Quality: %.3f\n", result.ProcessingInfo.Image1Quality)
		fmt.Printf("  Image 2 Quality: %.3f\n", result.ProcessingInfo.Image2Quality)
		fmt.Printf("  View Consistency: %v\n", result.ProcessingInfo.ViewConsistency)
		fmt.Printf("  Lighting Consistency: %v\n", result.ProcessingInfo.LightingConsistency)
	}
	
	if *outputPath == "" && *verbose {
		fmt.Printf("\nFull JSON Result:\n%s\n", string(resultJSON))
	}
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