# Image Cropper Lit

A modern web application for batch processing images with perspective correction and cropping, built using Lit components.

## Features

- Multiple image upload
- Intuitive quadrilateral selection for perspective correction
- Real-time preview of selection and transformed image (toggleable)
- One-click download of processed images
- Batch download of all processed images
- Display of image dimensions and point coordinates
- Keyboard navigation between images (Arrow Up/Down)
- Auto-advance option to move to the next image after completing a selection
- Auto-download option to save the processed image immediately after completing a selection
- Zoom controls to adjust the size of the image canvas
- Preferences (preview, auto-toggles, canvas zoom level) saved in local storage

## How It Works

### Components

1. **ImageCropperApp**: The main component that orchestrates the entire application.
2. **ImageList**: Displays thumbnails of uploaded images and manages image selection.
3. **ImageCanvas**: Renders the active image, handles point selection for perspective correction, and provides zoom controls.
4. **Controls**: Provides buttons for undoing, clearing, downloading, toggling previews, and batch downloading. Also includes toggles for auto-advance and auto-download features.
5. **Previews**: Shows real-time previews of the selected area and the transformed image when the quadrilateral is closed.

### Image Upload and Management

Users can upload multiple images using the file input. The `ImageCropperApp` component manages the list of uploaded images and the active image state.

### Image Rendering and Point Selection

The `ImageCanvas` component renders the active image and allows users to select four points to define the area for perspective correction. It also displays the image dimensions and the coordinates of the selected points.

### Perspective Correction and Preview

Once four points are selected, closing the quadrilateral, the application uses these points to apply a perspective transform. The `Previews` component shows a real-time preview of the selected area and the transformed image when toggled on.

### Preferences

User preferences such as preview visibility, auto-advance, auto-download, and zoom level are saved in the browser's local storage, allowing these settings to persist between sessions.

## Usage

1. Open the `index.html` file in a modern web browser.
2. Upload one or more images using the file input.
3. Click on an image thumbnail in the `ImageList` to select it for editing.
4. Click on the main canvas to place four points, defining the area for perspective correction.
5. Once the fourth point is placed, the previews will become available (toggle with the "Show/Hide Previews" button).
6. Use the "Undo" or "Clear" buttons to adjust your selection if needed.
7. Click "Download" to process and download the corrected image.
8. Use the "Download All" button to batch process and download all images with completed selections.
9. Use the Up and Down arrow keys to navigate between images.
10. Toggle the "Auto-advance" checkbox to automatically move to the next image after completing a selection.
11. Toggle the "Auto-download" checkbox to automatically save the processed image after completing a selection.
12. Use the "+" and "-" buttons to adjust the zoom level of the image canvas.

## Development

This project uses Lit components and ES modules. To modify the application:

1. Edit the relevant component files in the `components` directory.
2. Modify the `utils/perspectiveTransform.js` file to adjust the transformation logic.
3. Update the `index.html` file if you need to change the overall structure or add new scripts.

## Future Enhancements

- Implement drag-and-drop functionality for image upload.
- Add more image processing options, such as filters or adjustments.
- Improve mobile responsiveness and touch interactions.
- Add a progress indicator for batch processing.
