import PerspT from '../thirdparty/perspective-transform.js';

export function applyPerspectiveTransform(img, points) {
    console.log('PerspectiveTransform: Applying transform', {
        imgWidth: img.width,
        imgHeight: img.height,
        points: points
    });

    // Calculate side lengths
    const sideLengths = calculateSideLengths(points);
    
    // Determine target width and height
    const targetWidth = Math.max(sideLengths[0], sideLengths[2]);
    const targetHeight = Math.max(sideLengths[1], sideLengths[3]);

    console.log('PerspectiveTransform: Side lengths', {
        sideLengths: sideLengths,
        targetWidth: targetWidth,
        targetHeight: targetHeight
    });

    const srcPoints = points.flatMap(p => [p.x, p.y]);
    const dstPoints = [0, 0, targetWidth, 0, targetWidth, targetHeight, 0, targetHeight];
    
    console.log('PerspectiveTransform: Source and destination points', {
        srcPoints: srcPoints,
        dstPoints: dstPoints,
        targetWidth: targetWidth,
        targetHeight: targetHeight
    });

    const perspT = PerspT(srcPoints, dstPoints);
    const matrix = perspT.getMatrix();
    
    console.log('PerspectiveTransform: Transformation matrix', matrix);

    // Debug: Transform input points
    console.log('PerspectiveTransform: Input points transformation');
    points.forEach((point, index) => {
        const transformed = perspT.transform(point.x, point.y);
        const inverseTransformed = perspT.transformInverse(transformed[0], transformed[1]);
        console.log(`Point ${index}:`, {
            original: point,
            transformed: {x: transformed[0], y: transformed[1]},
            inverseTransformed: {x: inverseTransformed[0], y: inverseTransformed[1]}
        });
    });

    const canvas = document.createElement('canvas');
    canvas.width = targetWidth;
    canvas.height = targetHeight;
    const ctx = canvas.getContext('2d');

    ctx.save();
    ctx.transform(
        matrix[0], matrix[3], matrix[1],
        matrix[4], matrix[2], matrix[5]
    );
    ctx.transform(1, 0, 0, 1, matrix[6], matrix[7]);
    ctx.drawImage(img, 0, 0);
    ctx.restore();

    console.log('PerspectiveTransform: Transform applied', {
        canvasWidth: canvas.width,
        canvasHeight: canvas.height
    });
    
    return canvas;
}

function calculateSideLengths(points) {
    const lengths = [];
    for (let i = 0; i < 4; i++) {
        const p1 = points[i];
        const p2 = points[(i + 1) % 4];
        const dx = p2.x - p1.x;
        const dy = p2.y - p1.y;
        lengths.push(Math.sqrt(dx * dx + dy * dy));
    }
    return lengths;
}
