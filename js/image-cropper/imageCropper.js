let images = [];
let activeImageIndex = -1;
let points = [];
const canvas = document.getElementById('imageCanvas');
const ctx = canvas.getContext('2d');
const imageList = document.getElementById('image-list');
const fileInput = document.getElementById('fileInput');
const extractBtn = document.getElementById('extractBtn');

// Handle file upload
fileInput.addEventListener('change', (e) => {
    console.log('File input change event triggered');
    Array.from(e.target.files).forEach((file, index) => {
        console.log(`Processing file ${index + 1}:`, file.name);
        const img = new Image();
        img.src = URL.createObjectURL(file);
        img.onload = () => {
            const newIndex = images.length;
            images.push({ file, img, points: [] });
            console.log(`Image ${newIndex} loaded:`, img.width, 'x', img.height);
            addImageToList(img, newIndex);
            if (activeImageIndex === -1) {
                console.log('Loading first image');
                loadImage(0);
            }
        };
    });
});

function addImageToList(img, index) {
    console.log('Adding image to list:', index, img);
    const imgThumb = document.createElement('img');
    imgThumb.src = img.src;
    imgThumb.dataset.index = index;
    imgThumb.addEventListener('click', () => {
        const clickedIndex = parseInt(imgThumb.dataset.index, 10);
        console.log('Thumbnail clicked, loading image at index:', clickedIndex);
        loadImage(clickedIndex);
    });
    imageList.appendChild(imgThumb);
}

function loadImage(index) {
    console.log('Loading image at index:', index);
    activeImageIndex = index;
    const activeImage = images[activeImageIndex];
    console.log('Active image:', activeImage);
    
    if (!activeImage || !activeImage.img) {
        console.error('Invalid active image or image data');
        return;
    }
    
    clearCanvas();
    
    // Calculate scaling to fit image in canvas while maintaining aspect ratio
    const scale = Math.min(canvas.width / activeImage.img.width, canvas.height / activeImage.img.height);
    const scaledWidth = activeImage.img.width * scale;
    const scaledHeight = activeImage.img.height * scale;
    const offsetX = (canvas.width - scaledWidth) / 2;
    const offsetY = (canvas.height - scaledHeight) / 2;
    
    console.log('Drawing image with dimensions:', scaledWidth, scaledHeight, 'at offset:', offsetX, offsetY);
    ctx.drawImage(activeImage.img, offsetX, offsetY, scaledWidth, scaledHeight);
    
    points = activeImage.points;
    drawPoints();
    updatePreviews();
}

function clearCanvas() {
    ctx.clearRect(0, 0, canvas.width, canvas.height);
    points = [];
    drawPoints();
    updatePreviews();
}

function drawPoints() {
    ctx.clearRect(0, 0, canvas.width, canvas.height);
    
    // Redraw the image
    const activeImage = images[activeImageIndex];
    if (activeImage && activeImage.img) {
        const scale = Math.min(canvas.width / activeImage.img.width, canvas.height / activeImage.img.height);
        const scaledWidth = activeImage.img.width * scale;
        const scaledHeight = activeImage.img.height * scale;
        const offsetX = (canvas.width - scaledWidth) / 2;
        const offsetY = (canvas.height - scaledHeight) / 2;
        ctx.drawImage(activeImage.img, offsetX, offsetY, scaledWidth, scaledHeight);
    }

    // Draw points and quadrilateral
    if (points.length === 4) {
        // Draw quadrilateral
        ctx.beginPath();
        ctx.moveTo(points[0].x, points[0].y);
        points.forEach((point, index) => {
            ctx.lineTo(point.x, point.y);
        });
        ctx.closePath();
        ctx.strokeStyle = 'yellow';
        ctx.stroke();
    }
    
    // Draw points
    points.forEach((point, index) => {
        ctx.fillStyle = 'red';
        ctx.beginPath();
        ctx.arc(point.x, point.y, 5, 0, 2 * Math.PI);
        ctx.fill();
    });

    extractBtn.disabled = points.length !== 4;
}

function updatePreviews() {
    console.log('Updating previews');
    const selectedPreview = document.getElementById('selectedPreview');
    const transformedPreview = document.getElementById('transformedPreview');

    if (!selectedPreview || !transformedPreview) {
        console.error('Preview canvases not found');
        return;
    }

    const activeImage = images[activeImageIndex];
    if (!activeImage || !activeImage.img) {
        console.error('Invalid active image for previews');
        return;
    }

    // Selected area preview
    const selectedCtx = selectedPreview.getContext('2d');
    selectedCtx.clearRect(0, 0, selectedPreview.width, selectedPreview.height);
    
    if (points.length === 4) {
        const minX = Math.min(...points.map(p => p.x));
        const minY = Math.min(...points.map(p => p.y));
        const maxX = Math.max(...points.map(p => p.x));
        const maxY = Math.max(...points.map(p => p.y));
        const width = maxX - minX;
        const height = maxY - minY;
        
        selectedCtx.drawImage(canvas, minX, minY, width, height, 0, 0, selectedPreview.width, selectedPreview.height);
        console.log('Drew selected area preview');
    }

    // Perspective transform preview
    if (points.length === 4) {
        const transformedCanvas = applyPerspectiveTransform(activeImage.img, points);
        const transformedCtx = transformedPreview.getContext('2d');
        transformedCtx.clearRect(0, 0, transformedPreview.width, transformedPreview.height);
        transformedCtx.drawImage(transformedCanvas, 0, 0, transformedPreview.width, transformedPreview.height);
        console.log('Drew perspective transform preview');
    }
}

// Handle canvas click to add points
canvas.addEventListener('click', (e) => {
    const rect = canvas.getBoundingClientRect();
    const x = e.clientX - rect.left;
    const y = e.clientY - rect.top;
    console.log('Canvas clicked at:', x, y);
    
    if (points.length < 4) {
        points.push({ x, y });
        ctx.fillStyle = 'red';
        ctx.beginPath();
        ctx.arc(x, y, 5, 0, 2 * Math.PI);
        ctx.fill();
        drawPoints();
        updatePreviews();
    }
});

// Undo button functionality
document.getElementById('undoBtn').addEventListener('click', () => {
    console.log('Undo button clicked');
    points.pop();
    loadImage(activeImageIndex);
    updatePreviews();
});

// Clear button functionality
document.getElementById('clearBtn').addEventListener('click', () => {
    console.log('Clear button clicked');
    clearCanvas();
});

function applyPerspectiveTransform(img, points) {
    console.log('Applying perspective transform');
    const srcPoints = points.flatMap(p => [p.x, p.y]);
    const dstPoints = [0, 0, img.width, 0, img.width, img.height, 0, img.height];
    
    const perspT = PerspT(srcPoints, dstPoints);
    
    const tempCanvas = document.createElement('canvas');
    const tempCtx = tempCanvas.getContext('2d');
    tempCanvas.width = img.width;
    tempCanvas.height = img.height;
    
    // Draw the original image onto the temporary canvas
    tempCtx.drawImage(img, 0, 0, img.width, img.height);
    
    // Create a new canvas for the transformed image
    const outputCanvas = document.createElement('canvas');
    const outputCtx = outputCanvas.getContext('2d');
    outputCanvas.width = img.width;
    outputCanvas.height = img.height;
    
    // Apply the perspective transform
    const imageData = tempCtx.getImageData(0, 0, img.width, img.height);
    const outputImageData = outputCtx.createImageData(img.width, img.height);
    
    for (let y = 0; y < img.height; y++) {
        for (let x = 0; x < img.width; x++) {
            const [srcX, srcY] = perspT.transformInverse(x, y);
            if (srcX >= 0 && srcX < img.width && srcY >= 0 && srcY < img.height) {
                const srcIndex = (Math.floor(srcY) * img.width + Math.floor(srcX)) * 4;
                const dstIndex = (y * img.width + x) * 4;
                outputImageData.data[dstIndex] = imageData.data[srcIndex];
                outputImageData.data[dstIndex + 1] = imageData.data[srcIndex + 1];
                outputImageData.data[dstIndex + 2] = imageData.data[srcIndex + 2];
                outputImageData.data[dstIndex + 3] = imageData.data[srcIndex + 3];
            }
        }
    }
    
    outputCtx.putImageData(outputImageData, 0, 0);
    return outputCanvas;
}

function downloadImage(canvas) {
    const link = document.createElement('a');
    link.href = canvas.toDataURL('image/jpeg');
    link.download = 'cropped-image.jpg';
    link.click();
}

// Update the extract button functionality
extractBtn.addEventListener('click', () => {
    if (points.length === 4) {
        const activeImage = images[activeImageIndex];
        const croppedCanvas = applyPerspectiveTransform(activeImage.img, points);
        downloadImage(croppedCanvas);
    }
});

// Add this at the end of the file
console.log('imageCropper.js loaded');
