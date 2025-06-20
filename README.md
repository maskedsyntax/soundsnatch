<p align="center">
  <img src="assets/soundsnatch.svg" alt="SoundSnatch Logo" width="150" />
</p>

<h1 align="center">SoundSnatch</h1>

<p align="center">
  <b> Download Songs, Podcasts and other audio files from Youtube. 🎵 </b>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Platform-Linux-blue" />
  <img src="https://img.shields.io/badge/Platform-macOS-blue" />
  <img src="https://img.shields.io/badge/Built_with-Python-blueviolet" />
  <img src="https://img.shields.io/badge/UI-CLI-8bc34a" />
  <img src="https://img.shields.io/badge/License-MIT-blue.svg" />
</p>

## Overview

SoundSnatch is a CLI tool that allows users to download songs, podcasts and other audio files from YouTube within seconds. With a simple and user-friendly interface, you can easily fetch and save your favorite audio content in MP3 format.

## Features

- **Download Audio from YouTube**: Convert YouTube videos to high-quality MP3 files.
- **User-Friendly CLI**: Interactive prompts with default file paths and filenames.
- **Customizable Output**: Choose where to save files and rename them as desired.
- **ASCII Art**: Stylish ASCII art banner powered by the `toilet` utility (Linux/macOS).
- **Cross-Platform Support**: Works on Linux, macOS, and Windows with easy installation scripts.

## Prerequisites

- **Python 3.6+**: Required to run the script and create the virtual environment.
- **pip**: Python package manager for installing dependencies.
- **toilet** (optional, Linux/macOS): For ASCII art display. Install with:
  ```bash
  sudo apt install toilet  # Ubuntu/Debian
  sudo pacman -S toilet    # Arch Linux
  brew install toilet      # macOS (with Homebrew)
  ```
- Administrative privileges (Windows) or `sudo` (Linux/macOS) for installing the binary to system directories.

## Installation

SoundSnatch uses installation scripts to build a binary with PyInstaller and install it to a system directory. The process uses a virtual environment and `requirements.txt` for consistent dependency management.

### 1. Clone the Repository

```bash
git clone https://github.com/maskedsyntax/soundsnatch.git
cd soundsnatch
```

### 2. Run the Installation Script

#### Linux/macOS

1. Make the script executable:
   ```bash
   chmod +x scripts/install.sh
   ```
2. Run the script:
   ```bash
   ./scripts/install.sh
   ```
   - Creates a virtual environment in `venv/`.
   - Installs dependencies from `requirements.txt`.
   - Builds the binary with PyInstaller.
   - Installs the binary to `/usr/local/bin` (requires `sudo`).

#### Windows

1. Open Command Prompt or PowerShell as Administrator.
2. Navigate to the `scripts` directory:
   ```cmd
   cd path\to\soundsnatch\scripts
   ```
3. Run the script:
   ```cmd
   install.bat
   ```
   - Creates a virtual environment in `venv\`.
   - Installs dependencies from `requirements.txt`.
   - Builds the binary with PyInstaller.
   - Installs the binary to `C:\Program Files\SoundSnatch` and updates the system `PATH`.

### 4. Verify Installation

Run `soundsnatch` from any terminal:

```bash
soundsnatch
```

On Windows, you may need to restart your terminal for the `PATH` update to take effect.

## Usage

1. Launch SoundSnatch:
   ```bash
   soundsnatch
   ```
2. Enter a YouTube video URL when prompted.
3. Review the fetched video info (title, URL, duration).
4. Specify the save location (defaults to `~/Music` on Linux/macOS or equivalent on Windows).
5. Rename the output MP3 file (defaults to the video title).
6. Wait for the download to complete. The MP3 file will be saved to the specified location.

**Example**:

```
Enter video URL: https://www.youtube.com/watch?v=dQw4w9WgXcQ
Fetching video info...
Video info fetched:
Title: Rick Astley - Never Gonna Give You Up
Video URL: https://www.youtube.com/watch?v=dQw4w9WgXcQ
Duration: 212
✨ Where would you like to save your audio file? ~/Music
📝 What would you like to name your audio file? Never Gonna Give You Up
Download Complete! 'Rick Astley - Never Gonna Give You Up' has been successfully saved. Enjoy your audio experience! 🎧
```

**Exit**: Press `Ctrl+C` or `Ctrl+Q` to quit at any time.

## Troubleshooting

- **Missing `toilet`**: If ASCII art fails, install `toilet` (see Prerequisites) or ignore it, as it’s cosmetic.
- **Invalid URL**: Ensure the YouTube URL is valid and your network is active.
- **Permission Errors**:
  - Linux/macOS: Ensure you have `sudo` privileges for `/usr/local/bin`.
  - Windows: Run `install.bat` as Administrator.
- **Missing Dependencies**: Verify `requirements.txt` includes `pyinstaller` and `yt-dlp`. Check internet connectivity for `pip`.
- **Virtual Environment Issues**: Delete `venv/` and rerun the script to recreate it:
  ```bash
  rm -rf venv/
  ```

## Contributing

Contributions are welcome! Please:

1. Fork the repository.
2. Create a feature branch (`git checkout -b feature/your-feature`).
3. Commit changes (`git commit -m 'Add your feature'`).
4. Push to the branch (`git push origin feature/your-feature`).
5. Open a Pull Request.

## License

This project is licensed under the [MIT License](LICENSE). See the `LICENSE` file for details.

## Acknowledgments

- [yt-dlp](https://github.com/yt-dlp/yt-dlp) for audio downloading capabilities.
- [PyInstaller](https://www.pyinstaller.org/) for creating standalone binaries.
- [toilet](http://caca.zoy.org/wiki/toilet) for ASCII art generation.
