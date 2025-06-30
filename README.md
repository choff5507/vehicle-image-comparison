# Vehicle Image Comparison

A Go library for comparing vehicle images to detect license plate fraud using advanced computer vision and infrared signature analysis.

## Features

- **License Plate Fraud Detection**: Detects when the same license plate appears on different vehicles
- **IR Signature Analysis**: Analyzes infrared reflectivity patterns around license plates  
- **Multi-modal Support**: Works with daylight and infrared images
- **Robust Processing**: Quality assessment, geometric analysis, and light pattern matching
- **High Accuracy**: 86%+ similarity detection for identical vehicles

## Installation

### As a Library in Your Go Project

1. **Add to your project:**
```bash
go get github.com/choff5507/vehicle-image-comparison
```

2. **Import in your Go code:**
```go
import "github.com/choff5507/vehicle-image-comparison/pkg/vehiclecompare"
```

3. **Use the API:**
```go
service := vehiclecompare.NewVehicleComparisonService()
result, err := service.CompareVehicleImages("image1.jpg", "image2.jpg")
```

### Standalone Installation

```bash
# Clone and build
git clone https://github.com/choff5507/vehicle-image-comparison.git
cd vehicle-image-comparison
go build -o vehicle-compare cmd/main.go
```

## Requirements

- **Go 1.21+**
- **OpenCV 4.x** (via gocv)
- **Minimum 8GB RAM** recommended for image processing
- **Intel/AMD64 or ARM64** processor

### Installing OpenCV

**macOS:**
```bash
brew install opencv
```

**Ubuntu/Debian:**
```bash
sudo apt-get update
sudo apt-get install libopencv-dev pkg-config
```

**CentOS/RHEL:**
```bash
sudo yum install opencv-devel pkgconfig
```

**Windows:**
- Download OpenCV from https://opencv.org/releases/
- Follow gocv installation guide: https://gocv.io/getting-started/windows/

**Verify OpenCV Installation:**
```bash
pkg-config --modversion opencv4
# Should output version like: 4.8.0
```

## Integration Guide

### Step 1: Add to Your Project

Create a new Go module or add to existing project:

```bash
# New project
mkdir my-vehicle-app
cd my-vehicle-app
go mod init my-vehicle-app

# Add the library
go get github.com/choff5507/vehicle-image-comparison
```

### Step 2: Basic Integration

Create `main.go`:

```go
package main

import (
    "fmt"
    "log"
    "github.com/choff5507/vehicle-image-comparison/pkg/vehiclecompare"
)

func main() {
    // Create service (initialize once, reuse across requests)
    service := vehiclecompare.NewVehicleComparisonService()
    
    // Compare two vehicle images
    result, err := service.CompareVehicleImages("vehicle1.jpg", "vehicle2.jpg")
    if err != nil {
        log.Fatal(err)
    }
    
    // Check results
    fmt.Printf("Same vehicle: %v\n", result.IsSameVehicle)
    fmt.Printf("Similarity: %.3f\n", result.SimilarityScore)
    fmt.Printf("Processing time: %dms\n", result.ProcessingInfo.ProcessingTimeMs)
}
```

### Step 3: Advanced Integration

For production applications:

```go
package main

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "github.com/choff5507/vehicle-image-comparison/pkg/vehiclecompare"
    "github.com/choff5507/vehicle-image-comparison/internal/models"
)

type VehicleComparisonAPI struct {
    service *vehiclecompare.VehicleComparisonService
}

func NewAPI() *VehicleComparisonAPI {
    return &VehicleComparisonAPI{
        service: vehiclecompare.NewVehicleComparisonService(),
    }
}

// HTTP handler for vehicle comparison
func (api *VehicleComparisonAPI) compareHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    var request struct {
        Image1Base64 string `json:"image1_base64"`
        Image2Base64 string `json:"image2_base64"`
    }

    if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }

    // Perform comparison
    result, err := api.service.CompareVehicleImagesFromBase64(
        request.Image1Base64, 
        request.Image2Base64,
    )
    if err != nil {
        http.Error(w, fmt.Sprintf("Comparison failed: %v", err), http.StatusInternalServerError)
        return
    }

    // Return results
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(result)
}

// Fraud detection wrapper
func (api *VehicleComparisonAPI) DetectFraud(image1Path, image2Path string) (bool, float64, error) {
    result, err := api.service.CompareVehicleImages(image1Path, image2Path)
    if err != nil {
        return false, 0, err
    }

    // Consider it fraud if:
    // 1. Not same vehicle with high confidence
    // 2. Same view and lighting (rules out different perspectives)
    isFraud := !result.IsSameVehicle && 
               result.ConfidenceLevel == models.ConfidenceHigh &&
               result.ProcessingInfo.ViewConsistency &&
               result.ProcessingInfo.LightingConsistency

    return isFraud, result.SimilarityScore, nil
}

func main() {
    api := NewAPI()

    // Register HTTP handler
    http.HandleFunc("/compare", api.compareHandler)
    
    // Example fraud detection
    isFraud, similarity, err := api.DetectFraud("original.jpg", "suspect.jpg")
    if err != nil {
        log.Fatal(err)
    }
    
    if isFraud {
        fmt.Printf("🚨 FRAUD DETECTED! Similarity: %.3f\n", similarity)
    } else {
        fmt.Printf("✅ No fraud detected. Similarity: %.3f\n", similarity)
    }

    // Start server
    fmt.Println("Vehicle comparison API running on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

### Step 4: Build and Run

```bash
# Build your application
go build -o my-vehicle-app

# Run
./my-vehicle-app
```

## Quick Start Examples

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

- **Processing Time**: 300-600ms per comparison (depends on resolution)
- **Accuracy**: 86%+ similarity for identical vehicles
- **Memory Efficient**: Proper resource cleanup with defer statements
- **Robust**: Handles various image qualities and lighting conditions
- **Optimal Resolution**: 1280x960 to 2048x1568 for best accuracy/speed balance

### Resolution Guidelines

| Resolution | License Plate Size | Performance | Accuracy | Recommendation |
|------------|-------------------|-------------|----------|----------------|
| 640x480    | 40-60px width    | Very Fast   | Basic    | Minimum viable |
| 1280x960   | 80-120px width   | Fast        | Good     | **Recommended** |
| 2048x1568  | 120-200px width  | Moderate    | Excellent| High accuracy use cases |
| 4096x3072  | 240-400px width  | Slow        | Excellent| Overkill for most cases |

**Best Practice**: Use 1280x960 for real-time applications, 2048x1568 for forensic analysis.

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

# Test your integration
go run example/main.go path/to/image1.jpg path/to/image2.jpg
```

## Troubleshooting Integration

### Common Issues

**1. OpenCV not found:**
```
# Error: pkg-config: No package 'opencv4' found
# Solution: Install OpenCV development libraries
brew install opencv          # macOS
sudo apt install libopencv-dev  # Ubuntu
```

**2. CGO compilation errors:**
```bash
# Error: C compiler cannot create executables
# Solution: Install build tools
xcode-select --install       # macOS
sudo apt install build-essential  # Ubuntu
```

**3. Module import errors:**
```go
// Error: cannot find module
// Solution: Ensure correct import path
import "github.com/choff5507/vehicle-image-comparison/pkg/vehiclecompare"

// NOT: import "vehicle-image-comparison/pkg/vehiclecompare"
```

**4. Memory issues with large images:**
```go
// Solution: Check image resolution and available RAM
if result.ProcessingInfo.ProcessingTimeMs > 1000 {
    log.Printf("Slow processing detected - consider resizing images")
}
```

### Performance Optimization

```go
// For high-throughput applications, reuse the service
var globalService = vehiclecompare.NewVehicleComparisonService()

func compareImages(img1, img2 string) (*models.ComparisonResult, error) {
    return globalService.CompareVehicleImages(img1, img2)
}

// Process multiple comparisons concurrently
func compareBatch(pairs [][2]string) []Result {
    results := make([]Result, len(pairs))
    var wg sync.WaitGroup
    
    for i, pair := range pairs {
        wg.Add(1)
        go func(index int, images [2]string) {
            defer wg.Done()
            result, err := compareImages(images[0], images[1])
            results[index] = Result{result, err}
        }(i, pair)
    }
    
    wg.Wait()
    return results
}
```

### Docker Integration

Create `Dockerfile`:
```dockerfile
FROM golang:1.21-bullseye

# Install OpenCV
RUN apt-get update && apt-get install -y \
    libopencv-dev \
    pkg-config \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o vehicle-app main.go

EXPOSE 8080
CMD ["./vehicle-app"]
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