export function applyPerspectiveTransform(img, points) {
    const srcPoints = points.flatMap(p => [p.x, p.y]);
    const dstPoints = [0, 0, img.width, 0, img.width, img.height, 0, img.height];
    
    const perspT = PerspT(srcPoints, dstPoints);
    
    const tempCanvas = document.createElement('canvas');
    const tempCtx = tempCanvas.getContext('2d');
    tempCanvas.width = img.width;
    tempCanvas.height = img.height;
    
    tempCtx.drawImage(img, 0, 0, img.width, img.height);
    
    const outputCanvas = document.createElement('canvas');
    const outputCtx = outputCanvas.getContext('2d');
    outputCanvas.width = img.width;
    outputCanvas.height = img.height;
    
    const imageData = tempCtx.getImageData(0, 0, img.width, img.height);
    const outputImageData = outputCtx.createImageData(img.width, img.height);
    
}
