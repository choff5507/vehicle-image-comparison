# Vehicle Image Comparison

A Go library for comparing vehicle images to detect license plate fraud using advanced computer vision and infrared signature analysis.

## Features

- **License Plate Fraud Detection**: Detects when the same license plate appears on different vehicles
- **IR Signature Analysis**: Analyzes infrared reflectivity patterns around license plates  
- **Multi-modal Support**: Works with daylight and infrared images
- **Robust Processing**: Quality assessment, geometric analysis, and light pattern matching
- **High Accuracy**: 86%+ similarity detection for identical vehicles

## Installation

```bash
go get github.com/choff5507/vehicle-image-comparison
```

## Requirements

- Go 1.21+
- OpenCV 4.x (via gocv)

### Installing OpenCV

```bash
# macOS
brew install opencv

# Ubuntu/Debian  
sudo apt-get install libopencv-dev

# See gocv documentation for other platforms
```

## Quick Start

```go
package main

import (
    "fmt"
    "log"
    "github.com/choff5507/vehicle-image-comparison/pkg/vehiclecompare"
)

func main() {
    // Create service
    service := vehiclecompare.NewVehicleComparisonService()
    
    // Compare two vehicle images
    result, err := service.CompareVehicleImages("vehicle1.jpg", "vehicle2.jpg")
    if err != nil {
        log.Fatal(err)
    }
    
    // Check results
    fmt.Printf("Same vehicle: %v\n", result.IsSameVehicle)
    fmt.Printf("Similarity: %.3f\n", result.SimilarityScore)
    fmt.Printf("Confidence: %v\n", result.ConfidenceLevel)
}
```

## API Usage

### File-based Comparison

```go
service := vehiclecompare.NewVehicleComparisonService()
result, err := service.CompareVehicleImages("image1.jpg", "image2.jpg")
```

### Base64 Comparison

```go
result, err := service.CompareVehicleImagesFromBase64(base64Image1, base64Image2)
```

### Result Structure

```go
type ComparisonResult struct {
    IsSameVehicle   bool            `json:"is_same_vehicle"`
    SimilarityScore float64         `json:"similarity_score"`
    ConfidenceLevel ConfidenceLevel `json:"confidence_level"`
    DetailedScores  DetailedScores  `json:"detailed_scores"`
    ProcessingInfo  ProcessingInfo  `json:"processing_info"`
}

type DetailedScores struct {
    GeometricSimilarity    float64 `json:"geometric_similarity"`
    LightPatternSimilarity float64 `json:"light_pattern_similarity"`
    BumperSimilarity       float64 `json:"bumper_similarity"`
    ColorSimilarity        float64 `json:"color_similarity,omitempty"`
    ThermalSimilarity      float64 `json:"thermal_similarity,omitempty"`
}
```

## Command Line Tool

```bash
# Build the CLI tool
go build -o vehicle-compare cmd/main.go

# Compare vehicle images
./vehicle-compare -image1 car1.jpg -image2 car2.jpg -verbose

# Base64 input
./vehicle-compare -image1-base64 <base64> -image2-base64 <base64>

# Save results to JSON
./vehicle-compare -image1 car1.jpg -image2 car2.jpg -output results.json
```

## Use Cases

### License Plate Fraud Detection

The primary use case is detecting when a license plate has been moved from one vehicle to another:

```go
// Same license plate on different vehicles will be detected
// due to different IR reflectivity signatures around the plate
result, _ := service.CompareVehicleImages("original_vehicle.jpg", "suspect_vehicle.jpg")

if !result.IsSameVehicle && result.ConfidenceLevel == models.ConfidenceHigh {
    fmt.Println("Potential license plate fraud detected!")
}
```

### Vehicle Authentication

Verify that sequential images show the same vehicle:

```go
// Authenticate vehicle across multiple captures
results := []bool{}
for i := 1; i < len(imageFiles); i++ {
    result, _ := service.CompareVehicleImages(imageFiles[0], imageFiles[i])
    results = append(results, result.IsSameVehicle)
}
```

## How It Works

### IR Signature Analysis

The system's breakthrough feature analyzes infrared reflectivity patterns around license plates:

1. **License Plate Detection**: Identifies bright retroreflective plates in IR images
2. **Surrounding Area Analysis**: Extracts 1.5x plate area for context
3. **Material Classification**: Distinguishes metal, plastic, rubber surfaces
4. **3D Structure Mapping**: Analyzes shadows and depth information
5. **Signature Comparison**: Compares unique vehicle "fingerprints"

### Multi-Factor Analysis

- **Geometric Features** (35%): Vehicle proportions and structure
- **Light Patterns** (35%): Headlight/taillight configurations  
- **IR Signatures** (10%): Material reflectivity around license plate
- **Bumper Features** (20%): Surface analysis and mounting patterns

## Performance

- **Processing Time**: 300-400ms per comparison
- **Accuracy**: 86%+ similarity for identical vehicles
- **Memory Efficient**: Proper resource cleanup
- **Robust**: Handles various image qualities and lighting conditions

## Error Handling

The library provides comprehensive error handling:

```go
result, err := service.CompareVehicleImages("image1.jpg", "image2.jpg")
if err != nil {
    switch {
    case strings.Contains(err.Error(), "quality too low"):
        fmt.Println("Image quality insufficient for analysis")
    case strings.Contains(err.Error(), "cannot compare different"):
        fmt.Println("Images have incompatible view/lighting conditions")
    default:
        fmt.Printf("Comparison failed: %v\n", err)
    }
}
```

## Testing

```bash
# Run unit tests
go test ./...

# Run with integration tests (requires test images)
go test -v ./test/
```

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Submit a pull request

## Support

For issues and questions:
- Open an issue on GitHub
- Check the project_overview.md for detailed technical documentation