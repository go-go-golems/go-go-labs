# Image Cropper Lit

A modern web application for batch processing images with perspective correction and cropping, built using Lit components.

## Features

- Multiple image upload
- Intuitive quadrilateral selection for perspective correction
- Real-time preview of selection and transformed image
- One-click download of processed images
- Display of image dimensions and point coordinates

## How It Works

### Components

1. **ImageCropperApp**: The main component that orchestrates the entire application.
2. **ImageList**: Displays thumbnails of uploaded images and manages image selection.
3. **ImageCanvas**: Renders the active image and handles point selection for perspective correction.
4. **Controls**: Provides buttons for undoing, clearing, and downloading the image.
5. **Previews**: Shows real-time previews of the selected area and the transformed image when the quadrilateral is closed.

### Image Upload and Management

Users can upload multiple images using the file input. The `ImageCropperApp` component manages the list of uploaded images and the active image state.

### Image Rendering and Point Selection

The `ImageCanvas` component renders the active image and allows users to select four points to define the area for perspective correction. It also displays the image dimensions and the coordinates of the selected points.

### Perspective Correction and Preview

Once four points are selected, closing the quadrilateral, the application uses these points to apply a perspective transform. The `Previews` component shows a real-time preview of the selected area and the transformed image.

## Usage

1. Open the `index.html` file in a modern web browser.
2. Upload one or more images using the file input.
3. Click on an image thumbnail in the `ImageList` to select it for editing.
4. Click on the main canvas to place four points, defining the area for perspective correction.
5. Once the fourth point is placed, the previews will become visible.
6. Use the "Undo" or "Clear" buttons to adjust your selection if needed.
7. Click "Download" to process and download the corrected image.
8. Repeat steps 3-7 for each image in your batch.

## Development

This project uses Lit components and ES modules. To modify the application:

1. Edit the relevant component files in the `components` directory.
2. Modify the `utils/perspectiveTransform.js` file to adjust the transformation logic.
3. Update the `index.html` file if you need to change the overall structure or add new scripts.

## Future Enhancements

- Add support for keyboard navigation between images.
- Implement drag-and-drop functionality for image upload.
- Add more image processing options, such as filters or adjustments.
