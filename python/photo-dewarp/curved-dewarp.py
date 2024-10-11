import cv2
import numpy as np
import argparse
import sys
from scipy import interpolate

def debug_show(title, image):
    cv2.imshow(title, image)
    print(f"Showing: {title}")
    print("Press any key to continue, or 'q' to quit.")
    key = cv2.waitKey(0) & 0xFF
    if key == ord('q'):
        cv2.destroyAllWindows()
        sys.exit(0)

def load_and_preprocess(image_path, debug=False):
    image = cv2.imread(image_path)
    if image is None:
        print(f"Error: Unable to load image at path '{image_path}'.")
        sys.exit(1)
    
    gray = cv2.cvtColor(image, cv2.COLOR_BGR2GRAY)
    blurred = cv2.GaussianBlur(gray, (5, 5), 0)
    kernel = cv2.getStructuringElement(cv2.MORPH_RECT, (5, 5))
    morph = cv2.morphologyEx(blurred, cv2.MORPH_GRADIENT, kernel)
    
    if debug:
        debug_show("Original Image", image)
        debug_show("Morphological Gradient", morph)
    
    return image, morph

def detect_contours(morph, debug=False):
    edges = cv2.Canny(morph, 50, 150, apertureSize=3)
    contours, _ = cv2.findContours(edges, cv2.RETR_EXTERNAL, cv2.CHAIN_APPROX_SIMPLE)
    
    if debug:
        debug_show("Edge Detection", edges)
        contour_image = cv2.cvtColor(edges, cv2.COLOR_GRAY2BGR)
        cv2.drawContours(contour_image, contours, -1, (0, 255, 0), 2)
        debug_show("Detected Contours", contour_image)
    
    return contours

def find_largest_contour(contours, min_area=1000, debug=False):
    if not contours:
        print("No contours found.")
        return None
    
    contours = sorted(contours, key=cv2.contourArea, reverse=True)
    
    for i, contour in enumerate(contours):
        area = cv2.contourArea(contour)
        if area < min_area:
            continue
        
        peri = cv2.arcLength(contour, True)
        approx = cv2.approxPolyDP(contour, 0.02 * peri, True)
        
        if debug:
            print(f"Contour {i + 1}:")
            print(f"  Points: {len(approx)}")
            print(f"  Area: {area}")
        
        if len(approx) >= 4:
            if debug:
                print(f"Selected contour {i + 1} as the photo contour")
            return contour
    
    return None

def smooth_contour(contour, smoothing_factor=5, debug=False):
    contour = contour.squeeze()
    x, y = contour[:, 0], contour[:, 1]
    
    x = np.append(x, x[0])
    y = np.append(y, y[0])
    
    t = np.linspace(0, 1, len(x))
    spl_x = interpolate.UnivariateSpline(t, x, s=smoothing_factor)
    spl_y = interpolate.UnivariateSpline(t, y, s=smoothing_factor)
    
    t_new = np.linspace(0, 1, 1000)
    x_new = spl_x(t_new)
    y_new = spl_y(t_new)
    
    smoothed = np.vstack((x_new, y_new)).T
    smoothed = smoothed.astype(np.float32)
    
    if debug:
        debug_image = np.zeros((np.max(y).astype(int) + 10, np.max(x).astype(int) + 10, 3), dtype=np.uint8)
        cv2.drawContours(debug_image, [contour], -1, (0, 255, 0), 2)
        for point in smoothed.astype(int):
            cv2.circle(debug_image, tuple(point), 1, (0, 0, 255), -1)
        debug_show("Smoothed Contour", debug_image)
    
    return smoothed

def get_corners_from_smoothed_contour(smoothed_contour):
    top = smoothed_contour[np.argmin(smoothed_contour[:,1])]
    bottom = smoothed_contour[np.argmax(smoothed_contour[:,1])]
    left = smoothed_contour[np.argmin(smoothed_contour[:,0])]
    right = smoothed_contour[np.argmax(smoothed_contour[:,0])]
    
    pts = np.array([left, top, right, bottom], dtype="float32")
    return pts

def dewarp_image(image, pts, debug=False):
    def order_points(pts):
        rect = np.zeros((4, 2), dtype="float32")
        s = pts.sum(axis=1)
        rect[0] = pts[np.argmin(s)]      # Top-left
        rect[2] = pts[np.argmax(s)]      # Bottom-right
        diff = np.diff(pts, axis=1)
        rect[1] = pts[np.argmin(diff)]   # Top-right
        rect[3] = pts[np.argmax(diff)]   # Bottom-left
        return rect

    rect = order_points(pts)
    (tl, tr, br, bl) = rect

    widthA = np.linalg.norm(br - bl)
    widthB = np.linalg.norm(tr - tl)
    maxWidth = max(int(widthA), int(widthB))

    heightA = np.linalg.norm(tr - br)
    heightB = np.linalg.norm(tl - bl)
    maxHeight = max(int(heightA), int(heightB))

    dst = np.array([
        [0, 0],
        [maxWidth - 1, 0],
        [maxWidth - 1, maxHeight - 1],
        [0, maxHeight - 1]
    ], dtype="float32")

    M = cv2.getPerspectiveTransform(rect, dst)
    warped = cv2.warpPerspective(image, M, (maxWidth, maxHeight))

    if debug:
        debug_image = image.copy()
        for point in rect:
            cv2.circle(debug_image, tuple(point.astype(int)), 5, (0, 0, 255), -1)
        debug_show("Detected Corners", debug_image)
        debug_show("Dewarped Image", warped)

    return warped

def extract_and_dewarp_photo(image_path, output_path=None, display=True, debug=False):
    # Step 1: Load and preprocess
    image, morph = load_and_preprocess(image_path, debug)
    
    # Step 2: Detect contours
    contours = detect_contours(morph, debug)
    
    # Step 3: Find the largest contour
    photo_contour = find_largest_contour(contours, debug=debug)
    if photo_contour is None:
        print("Could not find a suitable contour for the photo.")
        sys.exit(1)
    
    # Step 4: Smooth the contour
    smoothed_contour = smooth_contour(photo_contour, debug=debug)
    
    # Step 5: Get corner points
    pts = get_corners_from_smoothed_contour(smoothed_contour)
    
    # Step 6: Dewarp the image
    warped = dewarp_image(image, pts, debug)
    
    # Step 7: Save or display the corrected photo
    if output_path:
        success = cv2.imwrite(output_path, warped)
        if success:
            print(f"Corrected photo saved to '{output_path}'.")
        else:
            print(f"Error: Could not save image to '{output_path}'.")
    
    if display:
        cv2.imshow("Dewarped Photo", warped)
        print("Press any key on the image window to exit.")
        cv2.waitKey(0)
        cv2.destroyAllWindows()

def main():
    parser = argparse.ArgumentParser(description="Extract and dewarp a curved photo from a scanned book image.")
    parser.add_argument("image_path", help="Path to the input scanned book image.")
    parser.add_argument("-o", "--output", help="Path to save the dewarped photo.", default=None)
    parser.add_argument("-d", "--display", action="store_true", help="Display the dewarped photo.")
    parser.add_argument("--debug", action="store_true", help="Enable debug mode to step through the process.")
    args = parser.parse_args()

    extract_and_dewarp_photo(args.image_path, args.output, args.display, args.debug)

if __name__ == "__main__":
    main()

