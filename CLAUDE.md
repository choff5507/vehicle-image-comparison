# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and Development Commands

### Build the main CLI tool
```bash
make build                    # Standard build
make build-prod              # Production build with optimizations
go build -o vehicle-compare cmd/main.go  # Direct build
```

### Test commands
```bash
make test                    # Run all tests
make test-verbose           # Run tests with verbose output
make test-coverage          # Run tests with coverage report
go test ./internal/comparator -v  # Test specific package
go test -run TestCompareVehicles ./internal/comparator  # Run specific test
```

### Code quality
```bash
make fmt                     # Format code
make vet                     # Run go vet
go fmt ./...                # Direct format
```

### Example usage
```bash
./vehicle-compare -image1 car1.jpg -image2 car2.jpg -verbose
./vehicle-compare -image1-base64 <base64> -image2-base64 <base64> -output result.json
go run example/main.go image1.jpg image2.jpg  # Run example
```

## Architecture Overview

### Core Processing Pipeline

The system implements a sophisticated vehicle comparison pipeline specifically designed to detect license plate fraud through IR signature analysis:

1. **Image Input** → Quality Assessment → View/Lighting Classification
2. **Feature Extraction** → Geometric + Light Patterns + IR Signatures
3. **Comparison Engine** → Weighted Scoring → Fraud Detection Result

### Key Architectural Decisions

**IR Signature Innovation**: The system's breakthrough is analyzing reflectivity patterns AROUND license plates rather than the plate content. This detects when the same plate is moved between vehicles by comparing the unique "fingerprint" of surrounding materials and structure.

**Modular Feature Extraction**: Each feature extractor (`internal/extractor/`) is independent, allowing parallel processing and easy extension. The comparison engine (`internal/comparator/engine.go`) uses adaptive weighting based on image conditions.

**NaN/Inf Protection**: All float calculations use `safeFloat64()` wrapper to prevent JSON marshaling failures. This is critical for robust API responses.

**OpenCV Compatibility**: The code handles multiple gocv API versions through defensive programming patterns (e.g., matrix operations, contour handling).

### Critical Code Paths

**License Plate Detection** (`internal/extractor/plate.go`):
- Detects bright retroreflective regions in IR images
- Falls back to brightest rectangular region if detection fails
- Returns `LicensePlateRegion` with confidence scoring

**IR Signature Extraction** (`internal/extractor/ir_signature.go`):
- Extracts 1.5x plate area for surrounding context
- Creates 8x8 reflectivity map excluding plate area
- Analyzes material signatures, shadows, and texture
- Returns `IRSignature` for comparison

**Comparison Engine** (`internal/comparator/engine.go`):
- `CompareVehicles()` orchestrates all feature comparisons
- `compareIRSignatures()` handles the fraud detection logic
- Adaptive weighting: Daylight (30/30/20/20) vs IR (35/35/20/10)
- Threshold: 75% similarity for daylight, 70% for infrared

### Service Layer Design

The public API (`pkg/vehiclecompare/service.go`) provides two entry points:
- `CompareVehicleImages(path1, path2)` - File-based comparison
- `CompareVehicleImagesFromBase64(b64_1, b64_2)` - Base64 input

Both methods follow: Validate → Process → Extract → Compare → Sanitize → Return

### Performance Considerations

- Image processing is CPU-bound; OpenCV operations are not thread-safe within a single image
- Service instance should be reused (`NewVehicleComparisonService()` once, compare many times)
- Resolution sweet spot: 1280x960 for speed, 2048x1568 for accuracy
- Memory usage: ~25MB per 2048x1568 image during processing

### Error Handling Patterns

Quality gates throughout pipeline:
- Image quality must exceed 0.3 score
- View/lighting classification needs 0.5+ confidence  
- Both images must have matching view and lighting
- All numeric results sanitized before JSON marshaling

## Testing Approach

- Unit tests focus on individual extractors with mock images
- Integration tests (`test/service_test.go`) require actual image files
- Example program (`example/main.go`) demonstrates full API usage
- Test with both identical images (expect 85%+ similarity) and different vehicles with same plate (expect <70% similarity with high confidence)

## Key Dependencies

- **gocv v0.34.0**: OpenCV Go bindings (requires OpenCV 4.x system library)
- **Go 1.21+**: For modern error handling and performance
- **pkg-config**: Required for OpenCV library detection

Verify OpenCV installation: `pkg-config --modversion opencv4`