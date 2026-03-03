<p align="center">
  <img src="assets/soundsnatch.svg" alt="SoundSnatch Logo" width="150" />
</p>

<h1 align="center">SoundSnatch</h1>

<p align="center">
  <b> Download Songs, Podcasts and other audio files from YouTube with a sleek TUI. 🎵 </b>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Platform-Linux-blue" />
  <img src="https://img.shields.io/badge/Platform-macOS-blue" />
  <img src="https://img.shields.io/badge/Built_with-Go-00ADD8" />
  <img src="https://img.shields.io/badge/UI-TUI-8bc34a" />
  <img src="https://img.shields.io/badge/Framework-Bubble_Tea-FF4081" />
  <img src="https://img.shields.io/badge/License-MIT-blue.svg" />
</p>

## Overview

SoundSnatch is a modern Terminal User Interface (TUI) tool rewritten in Go using the [Bubble Tea](https://github.com/charmbracelet/bubbletea) framework. It allows users to download songs, podcasts, and entire playlists from YouTube and YouTube Music with ease, featuring an interactive directory picker and customizable filenames.

## Features

- **Sleek TUI**: Powered by Bubble Tea and Lip Gloss for a beautiful terminal experience.
- **Playlist Support**: Download entire playlists from YouTube or YouTube Music into dedicated folders.
- **Interactive Directory Picker**: Browse and select download destinations using a built-in file picker.
- **On-the-fly Folder Creation**: Create new subfolders directly within the TUI to organize your music.
- **Customizable Filenames**: Automatically suggests video titles as filenames, with full editing support.
- **Cross-Platform**: Natively compiled for Linux and macOS.

## Prerequisites

- **Go 1.21+**: Required to build the application.
- **yt-dlp**: Required for audio extraction. We recommend the latest native binary.
  ```bash
  # Linux (example using curl to install locally)
  mkdir -p ~/.local/bin
  curl -L https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -o ~/.local/bin/yt-dlp
  chmod a+rx ~/.local/bin/yt-dlp
  ```

## Installation

### 1. Clone the Repository

```bash
git clone https://github.com/maskedsyntax/soundsnatch.git
cd soundsnatch
```

### 2. Build and Install

```bash
go build -o soundsnatch .
# Move it to your path if desired
sudo mv soundsnatch /usr/local/bin/
```

Or install directly via Go:

```bash
go install .
```

## Usage

1. Launch SoundSnatch:
   ```bash
   soundsnatch
   ```
2. **Enter URL**: Paste a YouTube video or playlist URL.
3. **Choose Destination**: 
   - Use arrows/`j`/`k` to navigate.
   - Press `Enter`/`l`/`→` to enter folders.
   - Press **`n`** to create a new folder.
   - Press **`s`** to select the highlighted folder.
   - Press **`S`** to select the current browsing directory.
4. **Name Your Download**: Edit the suggested filename or provide a folder name for playlists.
5. **Download**: Watch the spinner as SoundSnatch handles the extraction and conversion to MP3.

## Troubleshooting

- **yt-dlp not found**: Ensure `yt-dlp` is in your `$PATH`.
- **Invalid URL**: Ensure the URL is accessible and valid.
- **Permission Denied**: Check write permissions for the target directory.

## Contributing

Contributions are welcome! Please fork the repository and submit a pull request.

## License

This project is licensed under the [MIT License](LICENSE). See the `LICENSE` file for details.

## Acknowledgments

- [Charm Bracelet](https://charm.sh/) for the amazing Bubble Tea TUI ecosystem.
- [yt-dlp](https://github.com/yt-dlp/yt-dlp) for the powerful media extraction engine.
