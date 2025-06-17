
# SoundSnatch
Download Songs, Podcasts and other audio files from Youtube. 🎵
<hr>
This repository contains a Python script that allows you to download audio from YouTube videos and save it as an MP3 file on your local machine. The script leverages the `yt_dlp` library, which is a fork of the popular `youtube-dl` with additional features and improvements.

## Features 🚀

- **Interactive URL Input**: Enter the YouTube video URL interactively via the console.
- **Customizable Destination Path**: Choose where to save the MP3 file. If no path is provided, the file defaults to the `~/Music` directory.
- **Automatic File Naming**: The MP3 file is named after the title of the YouTube video.
- **Audio-Only Download**: The script extracts and converts the video’s audio into MP3 format.

## Prerequisites

Ensure the following are installed on your system:

- **Python 3.6+**
- **yt-dlp**: Installable via pip.

To install `yt-dlp`, run:

```bash
pip install yt-dlp
```

## Usage

1. **Clone the Repository**:

   ```bash
   git clone https://github.com/maskedsyntax/soundsnatch.git
   cd soundsnatch
   ```

2. **Run the Script**:

   Execute the script using Python:

   ```bash
   python3 main.py
   ```

3. **Enter the Video URL**:

   The script will prompt you to enter the URL of the YouTube video you want to download:

   ```
   Enter video URL:
   ```

   Paste the URL and press Enter.

4. **Choose the Destination Folder**:

   The script will prompt you to specify a destination folder:

   ```
   Select Destination:
   ```

   - Press Enter without typing anything to save the MP3 in the default `~/Music` directory.
   - Alternatively, enter the full path to the folder where you want to save the file.

5. **Download and Conversion**:

   The script will download the audio from the specified video, convert it to MP3 format, and save it to the specified destination.

## Example

Here’s an example session using the script:

```bash
$ python3 main.py
Enter video URL: https://www.youtube.com/watch?v=dQw4w9WgXcQ
Select Destination: ~/Downloads

Downloading and converting...
MP3 saved to: ~/Downloads/Never Gonna Give You Up.mp3
```

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for more details.

## Contributing

Feel free to open issues or submit pull requests if you want to contribute to this project.

## Acknowledgments

- **yt-dlp**: A powerful tool for downloading videos from various platforms, which this script depends on.

---

Happy downloading!

