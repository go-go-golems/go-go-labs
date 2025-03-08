To set up your Linux system with Visual Studio Code (VS Code) for programming the Raspberry Pi Pico W using Pimoroni's MicroPython firmware, follow these steps:

**1. Install Python 3.9 or Later:**

Ensure that Python 3.9 or a newer version is installed on your Linux system. You can check your Python version by running:

```bash
python3 --version
```

If you need to install or upgrade Python, refer to your distribution's package manager or the official Python website.

**2. Install Visual Studio Code:**

VS Code is available for Linux and can be installed using your distribution's package manager or by downloading the installer from the [official VS Code website](https://code.visualstudio.com/).

For Debian-based distributions like Ubuntu, you can install VS Code by running:

```bash
sudo apt update
sudo apt install code
```

After installation, you can launch VS Code by typing `code` in the terminal or by selecting it from your applications menu.

**3. Install Required VSCode Extensions:**

You'll need several VSCode extensions for the best development experience:

- Open VS Code
- Click on the Extensions icon in the sidebar or press `Ctrl+Shift+X`
- Install the following extensions:
  - MicroPico (by paulober) - For MicroPython development
  - Python - For Python language support
  - Pylance - For enhanced Python language features and type checking

**4. Flash Pimoroni's MicroPython Firmware to Raspberry Pi Pico W:**

To run Pimoroni's MicroPython libraries on your Pico W, you need to flash their custom firmware:

- Download the latest Pimoroni MicroPython firmware for the Pico W (`picow-v1.24.0-beta2-pimoroni-micropython.uf2` or newer) from the [Pimoroni GitHub releases page](https://github.com/pimoroni/pimoroni-pico/releases).
- Connect your Pico W to your computer using a USB cable while holding down the BOOTSEL button.
- A new mass storage device named "RPI-RP2" should appear.
- Drag and drop the downloaded `picow-*.uf2` firmware file onto the "RPI-RP2" drive.
- The device will reboot automatically after the transfer completes.
- The Pico W is now running Pimoroni's MicroPython firmware with all their libraries pre-installed.

**5. Configure Your Project in VS Code:**

1. Create a new folder for your project and open it in VS Code.
2. Create a `typings` directory in your project folder for type stubs:
   ```bash
   mkdir typings
   ```
3. Install Pimoroni's type stubs:
   ```bash
   pip install pimoroni-pico-stubs --target ./typings --no-user
   ```
4. Press `Ctrl+Shift+P` to open the command palette and type "Preferences: Open Workspace Settings (JSON)".
5. Add or update the following settings:
   ```json
   {
       "python.languageServer": "Pylance",
       "python.analysis.typeCheckingMode": "basic",
       "python.analysis.diagnosticSeverityOverrides": {
           "reportMissingModuleSource": "none"
       },
       "python.analysis.typeshedPaths": [
           "./typings/"
       ]
   }
   ```
6. Run "MicroPico: Configure Project" from the command palette to set up MicroPython project files.

**6. Write and Run MicroPython Code with Pimoroni Libraries:**

Create a new Python file (e.g., `main.py`) in your project folder. You can now use Pimoroni's libraries with full type hints and autocompletion. For example, to use the PicoGraphics library:

```python
from picographics import PicoGraphics, DISPLAY_PIMORONI_PICO_DISPLAY
from pimoroni import Button

display = PicoGraphics(display=DISPLAY_PIMORONI_PICO_DISPLAY)
display.set_pen(255, 255, 255)  # White
display.text("Hello, Pimoroni!", 10, 10, scale=2)
display.update()
```

To run the script:

- Ensure your Pico W is connected to your computer.
- In VS Code, press `Ctrl+Shift+P` and select "MicroPico: Run current file on Pico".
- The script will execute on your Pico W.

Note: Since you're using Pimoroni's firmware, all their libraries are pre-installed and ready to use. There's no need for additional package installation using `mip` or `upip`. The type stubs provide code completion and type checking in VSCode but don't affect the code running on the Pico W.

