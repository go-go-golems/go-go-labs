export function downloadImage(canvas, filename = 'image.jpg') {
    const link = document.createElement('a');
    link.download = filename;
    // Change the image format to JPEG and set quality to 0.9 (90%)
    link.href = canvas.toDataURL('image/jpeg', 0.9);
    link.click();
}
