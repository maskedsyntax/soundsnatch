#!/bin/bash

# Get the directory of the current script
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Define project paths and binary details
PROJECT_ROOT="$SCRIPT_DIR/.."
BINARY_NAME="soundsnatch"
MAIN_PY_PATH="$PROJECT_ROOT/src/soundsnatch.py"
ASSETS_DIR="$PROJECT_ROOT/assets"
DIST_DIR="$PROJECT_ROOT/dist"
TARGET_DIR="/usr/local/bin"
VENV_DIR="$PROJECT_ROOT/venv"
REQUIREMENTS_FILE="$PROJECT_ROOT/requirements.txt"

# Check if python3 is installed
if ! command -v python3 &> /dev/null; then
    echo "Error: python3 is not installed. Please install Python 3."
    exit 1
fi

# Check if pip is installed
if ! command -v pip &> /dev/null; then
    echo "Error: pip is not installed. Please install pip for Python and ensure it's in your PATH."
    exit 1
fi

# Check if requirements.txt exists
if [ ! -f "$REQUIREMENTS_FILE" ]; then
    echo "Error: requirements.txt not found at $REQUIREMENTS_FILE."
    echo "Please create a requirements.txt file with at least 'pyinstaller>=7.0'."
    exit 1
fi

# Create or reuse virtual environment
if [ ! -d "$VENV_DIR" ]; then
    echo "Creating virtual environment in $VENV_DIR..."
    python3 -m venv "$VENV_DIR"
    if [ $? -ne 0 ]; then
        echo "Error: Failed to create virtual environment in $VENV_DIR."
        exit 1
    fi
fi

# Activate the virtual environment
source "$VENV_DIR/bin/activate"
if [ $? -ne 0 ]; then
    echo "Error: Failed to activate virtual environment at $VENV_DIR."
    exit 1
fi

# Install dependencies from requirements.txt
echo "Installing dependencies from $REQUIREMENTS_FILE..."
pip install -r "$REQUIREMENTS_FILE"
if [ $? -ne 0 ]; then
    echo "Error: Failed to install dependencies from $REQUIREMENTS_FILE."
    deactivate
    exit 1
fi
echo "Dependencies installed successfully."

# Check if PyInstaller is available in the virtual environment
if ! command -v pyinstaller &> /dev/null; then
    echo "Error: PyInstaller not found in virtual environment. Ensure 'pyinstaller' is listed in $REQUIREMENTS_FILE."
    deactivate
    exit 1
fi

# Check if sudo is available
if ! command -v sudo &> /dev/null; then
    echo "Error: sudo is not installed. Please install sudo or run as root."
    deactivate
    exit 1
fi

# Check if the main Python file exists
if [ ! -f "$MAIN_PY_PATH" ]; then
    echo "Error: Main Python file not found at $MAIN_PY_PATH."
    deactivate
    exit 1
fi

# Check if the assets directory exists
if [ ! -d "$ASSETS_DIR" ]; then
    echo "Error: Assets directory not found at $ASSETS_DIR."
    deactivate
    exit 1
fi

# Clean previous build and dist directories
echo "Cleaning previous build and dist directories..."
rm -rf "$PROJECT_ROOT/build" "$DIST_DIR"

# Build the binary using PyInstaller from the virtual environment
echo "Building the binary..."
# Use platform-specific separator for --add-data
if [[ "$OSTYPE" == "darwin"* ]]; then
    DATA_SEPARATOR=":"
elif [[ "$OSTYPE" == "msys" || "$OSTYPE" == "win32" ]]; then
    DATA_SEPARATOR=";"
else
    DATA_SEPARATOR=":"
fi

pyinstaller --onefile --windowed \
    --add-data "$ASSETS_DIR${DATA_SEPARATOR}soundsnatch/assets" \
    --distpath "$DIST_DIR" \
    --workpath "$PROJECT_ROOT/build" \
    "$MAIN_PY_PATH"

# Check if the binary was created
BINARY_PATH="$DIST_DIR/$BINARY_NAME"
if [ ! -f "$BINARY_PATH" ]; then
    echo "Error: Binary not found at $BINARY_PATH. Build failed."
    deactivate
    exit 1
fi

# Deactivate virtual environment
deactivate

# Check if the target directory exists
if [ ! -d "$TARGET_DIR" ]; then
    echo "Error: Target directory $TARGET_DIR does not exist."
    exit 1
fi

# Copy the binary to the target directory using sudo
echo "Copying the binary to $TARGET_DIR..."
if ! sudo cp "$BINARY_PATH" "$TARGET_DIR/$BINARY_NAME"; then
    echo "Error: Failed to copy binary to $TARGET_DIR. Check sudo permissions."
    exit 1
fi

# Make the binary executable using sudo
echo "Making the binary executable..."
if ! sudo chmod +x "$TARGET_DIR/$BINARY_NAME"; then
    echo "Error: Failed to set executable permissions on $TARGET_DIR/$BINARY_NAME."
    exit 1
fi

# Verify the binary is in place and executable
if [ -x "$TARGET_DIR/$BINARY_NAME" ]; then
    echo "Installation complete! You can now run '$BINARY_NAME' from anywhere."
else
    echo "Error: Binary is not executable at $TARGET_DIR/$BINARY_NAME."
    exit 1
fi
