Here's a clear and expanded README.md for the Image Cropper project, with additional details on the image list manager and image rendering:

# Image Cropper

A versatile web application for batch processing images with perspective correction and cropping.

## Features

- Multiple image upload
- Intuitive quadrilateral selection for perspective correction
- Batch processing capabilities
- Real-time preview of selection
- One-click download of processed images

## How It Works

### Image Upload and Management

Users can upload multiple images at once using the file input. The application uses a File Manager component to handle these uploads:

```javascript
fileInput.addEventListener('change', (e) => {
    Array.from(e.target.files).forEach(file => {
        // Process each file...
    });
});
```

### Image List Manager

The Image List Manager displays thumbnails of all uploaded images and manages the active image state:

- Thumbnails are created and added to the `imageList` DOM element.
- Each thumbnail is clickable, allowing users to switch between images.
- The active image is tracked using the `activeImageIndex` variable.

```javascript
function addImageToList(img) {
    const imgThumb = document.createElement('img');
    imgThumb.src = img.src;
    imgThumb.addEventListener('click', () => loadImage(images.indexOf(img)));
    imageList.appendChild(imgThumb);
}
```

This component ensures efficient navigation through multiple images in a batch processing session.

### Image Rendering

When an image is selected, it's rendered on the canvas using the `loadImage` function:

```javascript
function loadImage(index) {
    activeImageIndex = index;
    const activeImage = images[activeImageIndex];
    clearCanvas();
    ctx.drawImage(activeImage.img, 0, 0, canvas.width, canvas.height);
    // ...
}
```

The image is scaled to fit the canvas dimensions while maintaining its aspect ratio. This is achieved by the `ctx.drawImage` method, which automatically handles the scaling when the destination dimensions (canvas width and height) differ from the source image dimensions.

### Perspective Correction

Users select four points on the image to define a quadrilateral. The application then uses these points to apply a perspective transform:

```javascript
function applyPerspectiveTransform(img, points) {
    // ... transformation logic ...
}
```

This function creates a new canvas with the corrected image, ready for download.

## Usage

1. Upload one or more images using the file input.
2. Click on an image thumbnail to select it for editing.
3. Click on the main canvas to place four points, defining the area for perspective correction.
4. Use the "Undo" or "Clear" buttons to adjust your selection if needed.
5. Click "Extract and Download" to process and download the corrected image.
6. Repeat steps 2-5 for each image in your batch.

## Future Enhancements

- Add support for keyboard navigation between images.
- Implement drag-and-drop functionality for image upload.
