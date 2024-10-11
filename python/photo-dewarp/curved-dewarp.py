# TODO: this doesn't really fully work, but it seems reasonably close.
# - [ ] The smoothed contours are way too smoothed
# - [ ] We need to compute the "most fitting" destination points and width/height instead of hard coding them
# - [ ] Add more code to debug / learn about TPS
# - [ ] Look at https://mzucker.github.io/2016/08/15/page-dewarping.html and https://safjan.com/tools-for-doc-deskewing-and-dewarping/

import cv2
import numpy as np
import argparse
import sys
from scipy import interpolate
from scipy.interpolate import Rbf
import os
import matplotlib.pyplot as plt

def debug_show(window_name, image):
    """Utility function to display debug images."""
    cv2.imshow(window_name, image)
    cv2.waitKey(0)
    cv2.destroyAllWindows()

def detect_rectangle_morphology(image, debug=False):
    gray = cv2.cvtColor(image, cv2.COLOR_BGR2GRAY)
    blurred = cv2.GaussianBlur(gray, (5,5), 0)
    thresh = cv2.adaptiveThreshold(blurred, 255,
                                   cv2.ADAPTIVE_THRESH_GAUSSIAN_C,
                                   cv2.THRESH_BINARY_INV, 11, 2)

    if debug:
        debug_show("Adaptive Threshold", thresh)

    kernel = cv2.getStructuringElement(cv2.MORPH_RECT, (5,5))
    closed = cv2.morphologyEx(thresh, cv2.MORPH_CLOSE, kernel, iterations=2)

    if debug:
        debug_show("Morphological Closing", closed)

    contours, _ = cv2.findContours(closed, cv2.RETR_EXTERNAL, cv2.CHAIN_APPROX_SIMPLE)
    if not contours:
        print("No contours found.")
        return None

    contours = sorted(contours, key=cv2.contourArea, reverse=True)

    if debug:
        contour_image = image.copy()
        cv2.drawContours(contour_image, contours[:5], -1, (0, 255, 0), 2)
        debug_show("Top 5 Contours", contour_image)

    detected_contours = []
    min_area = 1000  # Adjust this value based on your image size and requirements
    contours = sorted(contours, key=lambda x: cv2.arcLength(x, True), reverse=True)[:5]  # Keep top 5 by perimeter length
    for i, contour in enumerate(contours):
        peri = cv2.arcLength(contour, True)
        approx = cv2.approxPolyDP(contour, 0.02 * peri, True)

        if debug:
            print(f"Contour {i + 1}:")
            print(f"  Points: {len(approx)}")
            print(f"  Area: {cv2.contourArea(contour)}")

        if len(approx) >= 4:
            if debug:
                print(f"Selected contour {i + 1} as a photo contour")
            detected_contours.append(approx.reshape(-1, 2))

    return detected_contours if detected_contours else None

def order_points(pts):
    """
    Orders points in the order: top-left, top-right, bottom-right, bottom-left
    """
    rect = np.zeros((4, 2), dtype="float32")

    s = pts.sum(axis=1)
    diff = np.diff(pts, axis=1)

    rect[0] = pts[np.argmin(s)]
    rect[2] = pts[np.argmax(s)]
    rect[1] = pts[np.argmin(diff)]
    rect[3] = pts[np.argmax(diff)]

    return rect

def four_point_transform(image, pts):
    """
    Performs a perspective transform to obtain a top-down view of the image
    """
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

    return warped

def smooth_contour(image, contour, smoothing_factor=5, debug=False):
    print("Starting smooth_contour function")
    contour = contour.squeeze()
    x, y = contour[:, 0], contour[:, 1]
    
    print(f"Original contour shape: {contour.shape}")
    
    x = np.append(x, x[0])
    y = np.append(y, y[0])
    
    print(f"Appended x and y shapes: {x.shape}, {y.shape}")
    
    t = np.linspace(0, 1, len(x))
    spl_x = interpolate.UnivariateSpline(t, x, s=smoothing_factor)
    spl_y = interpolate.UnivariateSpline(t, y, s=smoothing_factor)
    
    print(f"Created splines with smoothing factor: {smoothing_factor}")
    
    t_new = np.linspace(0, 1, 1000)
    x_new = spl_x(t_new)
    y_new = spl_y(t_new)
    
    print(f"New x and y shapes after interpolation: {x_new.shape}, {y_new.shape}")
    
    smoothed = np.vstack((x_new, y_new)).T
    print(f"Final smoothed contour shape: {smoothed.shape}")
    
    if debug:
        fig, (ax1, ax2) = plt.subplots(1, 2, figsize=(20, 10))
        
        # Plot on the image
        ax1.imshow(cv2.cvtColor(image, cv2.COLOR_BGR2RGB))
        ax1.plot(x, y, 'r-', linewidth=2, label='Original')
        ax1.plot(x_new, y_new, 'b-', linewidth=2, label='Smoothed')
        ax1.legend()
        ax1.set_title('Contour Smoothing on Image')
        
        # Plot on a blank canvas
        ax2.plot(x, y, 'ro-', label='Original')
        ax2.plot(x_new, y_new, 'b-', label='Smoothed')
        ax2.legend()
        ax2.set_title('Contour Smoothing')
        
        plt.show()
    
    return smoothed.astype(np.float32)

def draw_debug_contours(image, contours, alpha=0.3):
    """Draw contours on a faded background image."""
    overlay = image.copy()
    output = image.copy()
    
    # Fade the background
    cv2.addWeighted(overlay, alpha, np.full_like(overlay, 255), 1 - alpha, 0, output)
    
    # Draw contours
    cv2.drawContours(output, contours, -1, (0, 255, 0), 2)
    
    return output

def get_key_points(image, smoothed, debug=False):
    top = smoothed[np.argmin(smoothed[:,1])]
    bottom = smoothed[np.argmax(smoothed[:,1])]
    left = smoothed[np.argmin(smoothed[:,0])]
    right = smoothed[np.argmax(smoothed[:,0])]
    
    # Midpoints
    top_mid = smoothed[np.argmin(smoothed[:,1]) + len(smoothed)//4]
    bottom_mid = smoothed[np.argmax(smoothed[:,1]) - len(smoothed)//4]
    left_mid = smoothed[np.argmin(smoothed[:,0]) + len(smoothed)//4]
    right_mid = smoothed[np.argmax(smoothed[:,0]) - len(smoothed)//4]
    
    key_points = np.array([left, top, right, bottom, top_mid, bottom_mid, left_mid, right_mid], dtype="float32")
    
    if debug:
        fig, (ax1, ax2) = plt.subplots(1, 2, figsize=(20, 10))
        
        # Plot on the image
        ax1.imshow(cv2.cvtColor(image, cv2.COLOR_BGR2RGB))
        ax1.plot(smoothed[:, 0], smoothed[:, 1], 'b-', linewidth=2, label='Smoothed Contour')
        ax1.plot(key_points[:, 0], key_points[:, 1], 'ro', markersize=10, label='Key Points')
        for i, point in enumerate(key_points):
            ax1.annotate(f'P{i}', (point[0], point[1]), xytext=(5, 5), textcoords='offset points')
        ax1.legend()
        ax1.set_title('Key Points on Image')
        
        # Plot on a blank canvas
        ax2.plot(smoothed[:, 0], smoothed[:, 1], 'b-', label='Smoothed Contour')
        ax2.plot(key_points[:, 0], key_points[:, 1], 'ro', label='Key Points')
        for i, point in enumerate(key_points):
            ax2.annotate(f'P{i}', (point[0], point[1]))
        ax2.legend()
        ax2.set_title('Key Points on Smoothed Contour')
        
        plt.show()
    
    return key_points

def thin_plate_spline_transform(image, src_pts, dst_pts, debug=False):
    """
    Performs Thin Plate Spline (TPS) transformation.
    """
    src_x, src_y = src_pts[:,0], src_pts[:,1]
    dst_x, dst_y = dst_pts[:,0], dst_pts[:,1]
    
    rbf_x = Rbf(src_x, src_y, dst_x, function='thin_plate')
    rbf_y = Rbf(src_x, src_y, dst_y, function='thin_plate')
    
    rows, cols = image.shape[:2]
    grid_x, grid_y = np.meshgrid(np.arange(cols), np.arange(rows))
    
    map_x = rbf_x(grid_x, grid_y)
    map_y = rbf_y(grid_x, grid_y)
    
    map_x = np.clip(map_x, 0, cols - 1).astype(np.float32)
    map_y = np.clip(map_y, 0, rows - 1).astype(np.float32)
    
    warped = cv2.remap(image, map_x, map_y, cv2.INTER_CUBIC)
    
    if debug:
        fig, (ax1, ax2) = plt.subplots(1, 2, figsize=(20, 10))
        ax1.imshow(cv2.cvtColor(image, cv2.COLOR_BGR2RGB))
        ax1.scatter(src_pts[:, 0], src_pts[:, 1], c='r', s=50)
        ax1.set_title('Original Image with Source Points')
        ax2.imshow(cv2.cvtColor(warped, cv2.COLOR_BGR2RGB))
        ax2.scatter(dst_pts[:, 0], dst_pts[:, 1], c='b', s=50)
        ax2.set_title('Warped Image with Destination Points')
        plt.show()
    
    return warped

def advanced_dewarp(image, contour, debug=False):
    """
    Applies advanced dewarping using TPS.
    """
    smoothed_contour = smooth_contour(image, contour, debug=debug)
    
    if debug:
        debug_image = draw_debug_contours(image, [smoothed_contour.astype(int)])
        for point in smoothed_contour.astype(int):
            cv2.circle(debug_image, tuple(point), 1, (0, 0, 255), -1)
        debug_show("Smoothed Contour", debug_image)
    
    src_pts = get_key_points(image, smoothed_contour, debug=debug)
    
    width = 800  # Desired width
    height = 600  # Desired height
    dst_pts = np.array([
        [0, 0],
        [width, 0],
        [width, height],
        [0, height],
        [width//2, 0],
        [width//2, height],
        [0, height//2],
        [width, height//2]
    ], dtype="float32")
    
    if debug:
        corr_image = draw_debug_contours(image, [smoothed_contour.astype(int)])
        for pt in src_pts:
            cv2.circle(corr_image, tuple(pt.astype(int)), 5, (0, 0, 255), -1)
        for pt in dst_pts:
            cv2.circle(corr_image, tuple(pt.astype(int)), 5, (255, 0, 0), -1)
        debug_show("Correspondence Points", corr_image)
    
    warped = thin_plate_spline_transform(image, src_pts, dst_pts, debug=debug)
    
    if debug:
        debug_show("Warped Image", warped)
    
    return warped

def apply_clahe(gray):
    clahe = cv2.createCLAHE(clipLimit=2.0, tileGridSize=(8,8))
    cl1 = clahe.apply(gray)
    return cl1

def extract_and_dewarp_photo(image_path, output_path=None, display=True, debug=False, advanced_dewarp_flag=False):
    """
    Extracts and dewarps a curved photo from the scanned book image.
    If output_path is provided, saves the dewarped photo to that path.
    If display is True, shows the dewarped photo.
    """
    image = cv2.imread(image_path)
    if image is None:
        print(f"Error: Unable to load image at path '{image_path}'.")
        sys.exit(1)

    orig = image.copy()

    # Apply CLAHE for contrast enhancement
    gray = cv2.cvtColor(image, cv2.COLOR_BGR2GRAY)
    enhanced_gray = apply_clahe(gray)
    image = cv2.cvtColor(enhanced_gray, cv2.COLOR_GRAY2BGR)

    photo_contours = detect_rectangle_morphology(image, debug=debug)

    if photo_contours is None:
        print("Error: Could not find a suitable photo contour in the image.")
        sys.exit(1)

    print(f"Found {len(photo_contours)} photo contours.")

    for i, contour in enumerate(photo_contours[:1]):
        if debug:
            contour_image = draw_debug_contours(image, [contour])
            debug_show(f"Detected Contour {i+1}", contour_image)

        if not advanced_dewarp_flag:
            warped = four_point_transform(image, contour)
        else:
            warped = advanced_dewarp(image, contour, debug=debug)

        if output_path:
            base, ext = os.path.splitext(output_path)
            current_output_path = f"{base}_{i+1}{ext}" if len(photo_contours) > 1 else output_path
            success = cv2.imwrite(current_output_path, warped)
            if success:
                print(f"Dewarped photo {i+1} saved to '{current_output_path}'.")
            else:
                print(f"Error: Could not save image to '{current_output_path}'.")

        if display or debug:
            cv2.imshow(f"Dewarped Photo {i+1}", warped)
            print("Press any key on the image window to continue.")
            cv2.waitKey(0)
            cv2.destroyAllWindows()

def main():
    parser = argparse.ArgumentParser(description="Extract and dewarp a curved photo from a scanned book image.")
    parser.add_argument("image_path", help="Path to the input scanned book image.")
    parser.add_argument("-o", "--output", help="Path to save the dewarped photo.", default=None)
    parser.add_argument("-d", "--display", action="store_true", help="Display the dewarped photo.")
    parser.add_argument("--debug", action="store_true", help="Enable debug mode with intermediate visualizations.")
    parser.add_argument("--advanced_dewarp", action="store_true", help="Enable advanced dewarping for curved borders.")
    args = parser.parse_args()

    extract_and_dewarp_photo(args.image_path, args.output, args.display, args.debug, args.advanced_dewarp)

if __name__ == "__main__":
    main()