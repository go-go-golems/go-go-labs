export function applyPerspectiveTransform(img, points) {
    console.log('PerspectiveTransform: Applying transform', {
        imgWidth: img.width,
        imgHeight: img.height,
        points: points
    });

    const srcPoints = points.flatMap(p => [p.x, p.y]);
    const dstPoints = [0, 0, img.width, 0, img.width, img.height, 0, img.height];
    
    console.log('PerspectiveTransform: Source and destination points', {
        srcPoints: srcPoints,
        dstPoints: dstPoints
    });

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

    console.log('PerspectiveTransform: Transform applied');
    
    return outputCanvas;
}
