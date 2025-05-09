import cv2
import numpy as np
import argparse
import sys
import random

def order_points(pts):
    """
    Orders points in the order: top-left, top-right, bottom-right, bottom-left
    """
    rect = np.zeros((4, 2), dtype="float32")

    # Sum and difference to find corners
    s = pts.sum(axis=1)
    diff = np.diff(pts, axis=1)

    # Top-left point has the smallest sum
    rect[0] = pts[np.argmin(s)]
    # Bottom-right point has the largest sum
    rect[2] = pts[np.argmax(s)]
    # Top-right has the smallest difference
    rect[1] = pts[np.argmin(diff)]
    # Bottom-left has the largest difference
    rect[3] = pts[np.argmax(diff)]

    return rect

def four_point_transform(image, pts):
    """
    Performs a perspective transform to obtain a top-down view of the image
    """
    rect = order_points(pts)
    (tl, tr, br, bl) = rect

    # Compute width of the new image
    widthA = np.linalg.norm(br - bl)
    widthB = np.linalg.norm(tr - tl)
    maxWidth = max(int(widthA), int(widthB))

    # Compute height of the new image
    heightA = np.linalg.norm(tr - br)
    heightB = np.linalg.norm(tl - bl)
    maxHeight = max(int(heightA), int(heightB))

    # Destination points for the transform
    dst = np.array([
        [0, 0],
        [maxWidth - 1, 0],
        [maxWidth - 1, maxHeight -1],
        [0, maxHeight - 1]
    ], dtype="float32")

    # Compute the perspective transform matrix and apply it
    M = cv2.getPerspectiveTransform(rect, dst)
    warped = cv2.warpPerspective(image, M, (maxWidth, maxHeight))

    return warped

def find_photo_contour(image, edged, debug=False):
    """
    Finds the contour that most likely represents the photo
    """
    contours, _ = cv2.findContours(edged.copy(), cv2.RETR_EXTERNAL, cv2.CHAIN_APPROX_SIMPLE)
    contours = sorted(contours, key=cv2.contourArea, reverse=True)[:5]  # Get top 5 largest contours

    if debug:
        print(f"Number of contours found: {len(contours)}")
        # Create a color version of the edge image to draw colored contours
        edge_with_contours = cv2.cvtColor(edged, cv2.COLOR_GRAY2BGR)
        cv2.drawContours(edge_with_contours, contours, -1, (0, 255, 0), 2)
        debug_show("Top 5 Contours on Edge Detection", edge_with_contours)

    for i, contour in enumerate(contours):
        # Approximate the contour
        peri = cv2.arcLength(contour, True)
        approx = cv2.approxPolyDP(contour, 0.02 * peri, True)

        if debug:
            print(f"Contour {i + 1}:")
            print(f"  Points: {len(approx)}")
            print(f"  Area: {cv2.contourArea(contour)}")
            x, y, w, h = cv2.boundingRect(contour)
            print(f"  Dimensions: {w}x{h}")

        # If the contour has four points or is close to a rectangle, consider it
        if len(approx) == 4 or (len(approx) >= 4 and cv2.isContourConvex(approx)):
            if debug:
                print(f"Selected contour {i + 1} as the photo contour")
            return approx

    # If no suitable contour found, try to find the largest rectangular area
    if debug:
        print("No suitable contour found, attempting to find largest rectangular area")

    max_area = 0
    max_rect = None
    for contour in contours:
        x, y, w, h = cv2.boundingRect(contour)
        area = w * h
        if area > max_area:
            max_area = area
            max_rect = np.array([[x, y], [x+w, y], [x+w, y+h], [x, y+h]], dtype=np.float32)

    if max_rect is not None and debug:
        print(f"Found largest rectangular area: {max_area}")
        edge_with_rect = cv2.cvtColor(edged, cv2.COLOR_GRAY2BGR)
        cv2.drawContours(edge_with_rect, [max_rect.astype(int)], -1, (0, 255, 0), 2)
        debug_show("Largest Rectangular Area on Edge Detection", edge_with_rect)

    return max_rect

def debug_show(title, image):
    cv2.imshow(title, image)
    print(f"Showing: {title}")
    print("Press any key to continue, or 'q' to quit.")
    key = cv2.waitKey(0) & 0xFF
    if key == ord('q'):
        cv2.destroyAllWindows()
        sys.exit(0)

def detect_rectangle_hough(image, debug=False):
    gray = cv2.cvtColor(image, cv2.COLOR_BGR2GRAY)
    blurred = cv2.GaussianBlur(gray, (5, 5), 0)
    edges = cv2.Canny(blurred, 50, 150, apertureSize=3)

    if debug:
        debug_show("Edge Detection for Hough", edges)

    # Detect lines using Hough Transform
    lines = cv2.HoughLinesP(edges, 1, np.pi/180, threshold=100, minLineLength=100, maxLineGap=10)
    if lines is None:
        print("No lines detected.")
        return None

    if debug:
        line_image = image.copy()
        for line in lines:
            x1, y1, x2, y2 = line[0]
            cv2.line(line_image, (x1, y1), (x2, y2), (0, 255, 0), 2)
        debug_show("Detected Lines", line_image)

    # Separate lines into horizontal and vertical based on their angles
    horizontals = []
    verticals = []
    for line in lines:
        x1, y1, x2, y2 = line[0]
        if abs(y2 - y1) < 10:
            horizontals.append(line[0])
        elif abs(x2 - x1) < 10:
            verticals.append(line[0])

    if len(horizontals) < 2 or len(verticals) < 2:
        print("Not enough horizontal or vertical lines detected.")
        return None

    if debug:
        print(f"Horizontal lines: {len(horizontals)}")
        print(f"Vertical lines: {len(verticals)}")

    # Find the extreme lines
    horizontals = sorted(horizontals, key=lambda line: line[1])
    verticals = sorted(verticals, key=lambda line: line[0])

    top_line = horizontals[0]
    bottom_line = horizontals[-1]
    left_line = verticals[0]
    right_line = verticals[-1]

    # Find intersections
    def line_intersection(line1, line2):
        x1, y1, x2, y2 = line1
        x3, y3, x4, y4 = line2

        denom = (x1 - x2)*(y3 - y4) - (y1 - y2)*(x3 - x4)
        if denom == 0:
            return None
        Px = ((x1*y2 - y1*x2)*(x3 - x4) - (x1 - x2)*(x3*y4 - y3*x4)) / denom
        Py = ((x1*y2 - y1*x2)*(y3 - y4) - (y1 - y2)*(x3*y4 - y3*x4)) / denom
        return [Px, Py]

    top_left = line_intersection(top_line, left_line)
    top_right = line_intersection(top_line, right_line)
    bottom_right = line_intersection(bottom_line, right_line)
    bottom_left = line_intersection(bottom_line, left_line)

    if None in [top_left, top_right, bottom_right, bottom_left]:
        print("Could not find all intersection points.")
        return None

    if debug:
        corner_image = image.copy()
        for point in [top_left, top_right, bottom_right, bottom_left]:
            cv2.circle(corner_image, tuple(map(int, point)), 5, (0, 0, 255), -1)
        debug_show("Detected Corners", corner_image)

    pts = np.array([top_left, top_right, bottom_right, bottom_left], dtype="float32")
    return pts

def detect_rectangle_morphology(image, debug=False, interactive=False):
    gray = cv2.cvtColor(image, cv2.COLOR_BGR2GRAY)
    blurred = cv2.GaussianBlur(gray, (5,5), 0)

    # Initial parameters
    block_size = 11
    C = 2

    def update_threshold(block_size, C):
        thresh = cv2.adaptiveThreshold(blurred, 255,
                                       cv2.ADAPTIVE_THRESH_GAUSSIAN_C,
                                       cv2.THRESH_BINARY_INV, block_size, C)
        kernel = cv2.getStructuringElement(cv2.MORPH_RECT, (5,5))
        closed = cv2.morphologyEx(thresh, cv2.MORPH_CLOSE, kernel, iterations=2)
        return thresh, closed

    thresh, closed = update_threshold(block_size, C)

    if interactive:
        print(f"Initial Block Size: {block_size}")
        print(f"Initial C: {C}")
        print("Use the following keys to adjust parameters:")
        print("'b' to increase Block Size, 'B' to decrease")
        print("'c' to increase C, 'C' to decrease")
        print("'q' to quit")

        while True:
            cv2.imshow('Adaptive Threshold', thresh)
            cv2.imshow('Morphological Closing', closed)
            key = cv2.waitKey(0) & 0xFF
            
            if key == ord('b'):
                block_size += 2
            elif key == ord('B'):
                block_size = max(3, block_size - 2)
            elif key == ord('c'):
                C += 1
            elif key == ord('C'):
                C = max(0, C - 1)
            elif key == ord('q'):
                break
            
            block_size = block_size if block_size % 2 == 1 else block_size + 1
            thresh, closed = update_threshold(block_size, C)
            print(f"Block Size: {block_size}, C: {C}")

        cv2.destroyAllWindows()
    if debug:
        debug_show("Adaptive Threshold", thresh)
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

    for i, contour in enumerate(contours):
        peri = cv2.arcLength(contour, True)
        approx = cv2.approxPolyDP(contour, 0.02 * peri, True)

        if debug:
            print(f"Contour {i + 1}:")
            print(f"  Points: {len(approx)}")
            print(f"  Area: {cv2.contourArea(contour)}")

        if len(approx) == 4:
            if debug:
                print(f"Selected contour {i + 1} as the photo contour")
            return approx.reshape(4, 2)

    return None

def detect_rectangle_features(image, debug=False):
    gray = cv2.cvtColor(image, cv2.COLOR_BGR2GRAY)
    orb = cv2.ORB_create(5000)
    keypoints, descriptors = orb.detectAndCompute(gray, None)

    if debug:
        kp_image = cv2.drawKeypoints(image, keypoints, None, color=(0, 255, 0), flags=0)
        debug_show("ORB Keypoints", kp_image)
        print(f"Number of keypoints detected: {len(keypoints)}")

    # Placeholder: Returning None as feature-based detection is non-trivial without a reference
    return None

# TODO: This is not working yet
def detect_rectangle_growing(image, debug=False):
    gray = cv2.cvtColor(image, cv2.COLOR_BGR2GRAY)
    height, width = gray.shape
    
    def is_edge(x, y, threshold=30):
        if x == 0 or y == 0 or x == width-1 or y == height-1:
            return True
        differences = [
            abs(int(gray[y, x]) - int(gray[y-1, x])),
            abs(int(gray[y, x]) - int(gray[y+1, x])),
            abs(int(gray[y, x]) - int(gray[y, x-1])),
            abs(int(gray[y, x]) - int(gray[y, x+1]))
        ]
        return max(differences) > threshold

    def grow_region(start_x, start_y):
        stack = [(start_x, start_y)]
        region = set()
        iterations = 0
        max_iterations = width * height  # Set a maximum number of iterations
        jump_size = min(width, height) // 10  # Start with a large jump

        while stack and iterations < max_iterations:
            x, y = stack.pop()
            if (x, y) not in region and 0 <= x < width and 0 <= y < height:
                if not is_edge(x, y):
                    region.add((x, y))
                    # Add points with larger jumps
                    for dx, dy in [(jump_size, 0), (-jump_size, 0), (0, jump_size), (0, -jump_size)]:
                        new_x, new_y = x + dx, y + dy
                        if 0 <= new_x < width and 0 <= new_y < height:
                            stack.append((new_x, new_y))
                else:
                    # If we hit an edge, refine the search
                    if jump_size > 1:
                        jump_size //= 2
                        stack.append((x, y))  # Re-add the point for refined search
                    else:
                        # Add adjacent points for fine-grained edge detection
                        for dx, dy in [(-1, 0), (1, 0), (0, -1), (0, 1)]:
                            new_x, new_y = x + dx, y + dy
                            if 0 <= new_x < width and 0 <= new_y < height:
                                stack.append((new_x, new_y))

            iterations += 1
            if iterations % 1000 == 0 and debug:
                print(f"Grow iteration: {iterations}, Region size: {len(region)}, Jump size: {jump_size}")

        if iterations == max_iterations and debug:
            print("Warning: Maximum iterations reached in region growing")
        return region
    # Start with multiple random points
    attempts = 10
    regions = []
    for i in range(attempts):
        start_x = random.randint(width//4, 3*width//4)
        start_y = random.randint(height//4, 3*height//4)
        if debug:
            print(f"Attempt {i+1}: Starting point ({start_x}, {start_y})")
        region = grow_region(start_x, start_y)
        regions.append(region)
        if debug:
            print(f"Attempt {i+1}: Region size: {len(region)}")

    # Select the largest region
    largest_region = max(regions, key=len)
    if debug:
        print(f"Largest region size: {len(largest_region)}")

    # Find the bounding rectangle of the largest region
    min_x = min(x for x, _ in largest_region)
    max_x = max(x for x, _ in largest_region)
    min_y = min(y for _, y in largest_region)
    max_y = max(y for _, y in largest_region)

    if debug:
        print(f"Bounding rectangle: ({min_x}, {min_y}) to ({max_x}, {max_y})")
        debug_image = image.copy()
        for x, y in largest_region:
            debug_image[y, x] = [0, 255, 0]  # Mark region in green
        cv2.rectangle(debug_image, (min_x, min_y), (max_x, max_y), (0, 0, 255), 2)
        debug_show("Growing Rectangle Detection", debug_image)

    # Convert rectangle to points
    pts = np.array([
        [min_x, min_y],  # Top-left
        [max_x, min_y],  # Top-right
        [max_x, max_y],  # Bottom-right
        [min_x, max_y]   # Bottom-left
    ], dtype="float32")

    return pts
def extract_and_correct_photo(image_path, output_path=None, display=True, debug=False, method='contour', interactive=False):
    image = cv2.imread(image_path)
    if image is None:
        print(f"Error: Unable to load image at path '{image_path}'.")
        sys.exit(1)

    if debug:
        debug_show("Original Image", image)

    if method == 'hough':
        pts = detect_rectangle_hough(image, debug)
    elif method == 'morphology':
        pts = detect_rectangle_morphology(image, debug, interactive)
    elif method == 'features':
        pts = detect_rectangle_features(image, debug)
    elif method == 'growing':
        pts = detect_rectangle_growing(image, debug)
    else:  # Default to contour method
        edged = cv2.Canny(cv2.cvtColor(image, cv2.COLOR_BGR2GRAY), 50, 200)
        if debug:
            debug_show("Edge Detection", edged)
        pts = find_photo_contour(image, edged, debug)

    if pts is None:
        print(f"Failed to detect rectangle using {method} method.")
        sys.exit(1)

    if debug:
        corner_image = image.copy()
        for point in pts:
            cv2.circle(corner_image, tuple(map(int, point)), 5, (0, 0, 255), -1)
        debug_show("Detected Corners", corner_image)

    warped = four_point_transform(image, pts)

    if debug:
        debug_show("Warped Image", warped)

    if output_path:
        success = cv2.imwrite(output_path, warped)
        if success:
            print(f"Corrected photo saved to '{output_path}'.")
        else:
            print(f"Error: Could not save image to '{output_path}'.")

    if display:
        cv2.imshow("Corrected Photo", warped)
        print("Press any key on the image window to exit.")
        cv2.waitKey(0)
        cv2.destroyAllWindows()

def main():
    parser = argparse.ArgumentParser(description="Extract and correct a warped rectangular photo from a scanned book image.")
    parser.add_argument("image_path", help="Path to the input scanned book image.")
    parser.add_argument("-o", "--output", help="Path to save the corrected photo.", default=None)
    parser.add_argument("-d", "--display", action="store_true", help="Display the corrected photo.")
    parser.add_argument("--debug", action="store_true", help="Enable debug mode to step through the process.")
    parser.add_argument("-m", "--method", choices=['contour', 'hough', 'morphology', 'features', 'growing'], 
                        default='contour', help="Method to use for rectangle detection.")
    parser.add_argument("-i", "--interactive", action="store_true", help="Enable interactive mode for parameter tuning.")
    args = parser.parse_args()

    extract_and_correct_photo(args.image_path, args.output, args.display, args.debug, args.method, args.interactive)

if __name__ == "__main__":
    main()