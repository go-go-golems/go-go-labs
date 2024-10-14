# Focus Stack

Focus Stack is a web application that allows users to capture multiple images at different focus depths using their device's camera. This tool is particularly useful for focus stacking in photography, a technique used to create images with a greater depth of field.

## Features

- Camera selection
- Resolution selection
- Customizable frame count and countdown speed
- Live camera preview
- Progress visualization during capture
- Automatic image download after capture

## Getting Started

1. Clone this repository or download the source files.
2. Open the `index.html` file in a modern web browser (Chrome, Firefox, Safari, or Edge).
3. Grant camera permissions when prompted.

## Usage

1. Select your desired camera from the dropdown menu.
2. Choose the preferred resolution.
3. Set the number of frames you want to capture.
4. Adjust the countdown speed between captures.
5. Click the "Start Capture" button to begin the process.
6. The application will countdown and capture the specified number of frames.
7. Each captured frame will be automatically downloaded to your device.

## Technical Details

This application is built using Lit, a lightweight library for building fast, lightweight web components. It uses the MediaDevices API to access the camera and capture frames.

## Browser Compatibility

Focus Stack works best on modern browsers that support the MediaDevices API and Web Components. Ensure your browser is up to date for the best experience.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is open source and available under the [MIT License](LICENSE).
