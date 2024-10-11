# Photo Dewarping Tool

This Python script is designed to extract and correct warped rectangular photos from scanned book images. It implements multiple algorithms to detect and correct the perspective distortion of photos, making them appear as if they were taken from a top-down view.

## Features

- Multiple rectangle detection methods:
  - Contour detection
  - Hough Line Transform
  - Morphological operations
  - Feature detection (placeholder)
- Perspective correction
- Debug mode for visualizing intermediate steps
- Command-line interface for easy use

## Requirements

- Python 3.x
- OpenCV (cv2)
- NumPy

## Usage

```
python photo-dewarp.py <image_path> [-o OUTPUT] [-d] [--debug] [-m {contour,hough,morphology,features}]
```

Arguments:
- `image_path`: Path to the input scanned book image
- `-o`, `--output`: Path to save the corrected photo (optional)
- `-d`, `--display`: Display the corrected photo
- `--debug`: Enable debug mode to step through the process
- `-m`, `--method`: Method to use for rectangle detection (default: contour)

## Algorithms

### 1. Contour Detection (Default)

This method uses edge detection and contour finding to identify the rectangular photo:

1. Convert the image to grayscale
2. Apply Gaussian blur to reduce noise
3. Use Canny edge detection to find edges
4. Find contours in the edge image
5. Identify the largest contour with four sides (approximated)

### 2. Hough Line Transform

This method detects straight lines in the image to identify the photo's edges:

1. Convert to grayscale and apply Gaussian blur
2. Use Canny edge detection
3. Apply Hough Line Transform to detect lines
4. Separate lines into horizontal and vertical
5. Find intersections of extreme lines to get corners

### 3. Morphological Operations

This method uses morphological operations to enhance edges before contour detection:

1. Convert to grayscale and apply Gaussian blur
2. Apply adaptive thresholding
3. Perform morphological closing to fill gaps in edges
4. Find and filter contours to identify the photo

### 4. Feature Detection (Placeholder)

This method uses feature detection to identify keypoints in the image. The current implementation is a placeholder and would require further development to be functional.

## Perspective Correction

Once the corners of the photo are detected, the script applies a perspective transform:

1. Order the detected points (top-left, top-right, bottom-right, bottom-left)
2. Calculate the dimensions of the output image
3. Apply a perspective transform to obtain a top-down view

## Debug Mode

When run with the `--debug` flag, the script displays intermediate results and waits for user input at each step. This is useful for understanding the process and diagnosing issues with particular images.

## Limitations and Future Improvements

- The feature detection method is currently a placeholder and needs implementation
- The algorithms may struggle with highly distorted or poorly lit images
- Future versions could implement machine learning-based approaches for more robust detection

## Contributing

Contributions to improve the algorithms, add new detection methods, or enhance the user interface are welcome. Please submit pull requests or open issues to discuss proposed changes.
