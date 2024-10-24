import json
from typing import List, Tuple

def load_json_data(file_path: str) -> dict:
    with open(file_path, 'r') as f:
        return json.load(f)

def scale_coordinates(polygon: List[List[float]], max_width: int, max_height: int) -> List[Tuple[int, int]]:
    scaled_polygon = []
    for x, y in polygon:
        scaled_x = int(x * max_width / 1535)
        scaled_y = int(y * max_height / 1988)
        scaled_polygon.append((scaled_x, scaled_y))
    return scaled_polygon

def create_empty_canvas(width: int, height: int) -> List[List[str]]:
    return [[' ' for _ in range(width)] for _ in range(height)]

def draw_text_on_canvas(canvas: List[List[str]], text: str, polygon: List[Tuple[int, int]]) -> None:
    min_x = min(p[0] for p in polygon)
    min_y = min(p[1] for p in polygon)
    max_x = max(p[0] for p in polygon)
    max_y = max(p[1] for p in polygon)

    text_length = len(text)
    available_width = max_x - min_x + 1

    if text_length <= available_width:
        start_x = min_x + (available_width - text_length) // 2
        for i, char in enumerate(text):
            if 0 <= start_x + i < len(canvas[0]) and min_y < len(canvas):
                canvas[min_y][start_x + i] = char
    else:
        words = text.split()
        current_line = []
        current_length = 0
        y_offset = 0

        for word in words:
            if current_length + len(word) + (1 if current_line else 0) <= available_width:
                current_line.append(word)
                current_length += len(word) + (1 if current_length > 0 else 0)
            else:
                if current_line:
                    line_text = ' '.join(current_line)
                    start_x = min_x + (available_width - len(line_text)) // 2
                    for i, char in enumerate(line_text):
                        if 0 <= start_x + i < len(canvas[0]) and min_y + y_offset < len(canvas):
                            canvas[min_y + y_offset][start_x + i] = char
                    y_offset += 1
                current_line = [word]
                current_length = len(word)

        if current_line:
            line_text = ' '.join(current_line)
            start_x = min_x + (available_width - len(line_text)) // 2
            for i, char in enumerate(line_text):
                if 0 <= start_x + i < len(canvas[0]) and min_y + y_offset < len(canvas):
                    canvas[min_y + y_offset][start_x + i] = char

def main():
    json_data = load_json_data('results.json')
    text_lines = json_data['test'][0]['text_lines']

    canvas_width = 120
    canvas_height = 80
    canvas = create_empty_canvas(canvas_width, canvas_height)

    for line in text_lines:
        polygon = scale_coordinates(line['polygon'], canvas_width, canvas_height)
        text = line['text']
        draw_text_on_canvas(canvas, text, polygon)

    for row in canvas:
        print(''.join(row))

if __name__ == "__main__":
    main()