import os
import subprocess
import yt_dlp as youtube_dl
import readline
import sys


# ANSI color codes
GREEN = "\033[92m"
RED = "\033[91m"
CYAN = "\033[96m"
YELLOW = "\033[93m"
MAGENTA = "\033[95m"
BLUE = "\033[94m"
BROWN = "\033[38;5;94m"
RESET = "\033[0m"


BRIGHT_BLACK = "\033[90m"
BRIGHT_RED = "\033[91m"
BRIGHT_GREEN = "\033[92m"
BRIGHT_YELLOW = "\033[93m"
BRIGHT_BLUE = "\033[94m"
BRIGHT_MAGENTA = "\033[95m"
BRIGHT_CYAN = "\033[96m"
BRIGHT_WHITE = "\033[97m"


def get_mp3():

    INTRO = "\nDownload Songs, Podcasts and other audio files from Youtube 🎵 \nPress Ctrl-q anytime to quit!\n"

    print_ascii_art("SOUNDSNATCH - CLI")
    print(f"{BLUE}{INTRO}{RESET}")

    title, webpage_url, duration = fetch_video_info(input(f"Enter video URL: {GREEN}"))

    if not title:
        return

    DEAFAULT_PATH = "~/Music"
    DEAFAULT_FILENAME = title
    VIDEO_INFO = f"\n{BRIGHT_YELLOW}Video info fetched: \n{RESET}Title: {BRIGHT_RED}{title}\n{RESET}Video URL: {BRIGHT_RED}{webpage_url}\n{RESET}Duration: {BRIGHT_RED}{duration}{RESET}\n"

    print(VIDEO_INFO)

    # Select destination path for audio file
    print(f"{YELLOW}✨ Where would you like to save your audio file? {RESET}")
    path = set_input_default(DEAFAULT_PATH).strip()

    # Allow user to change the filename
    print(f"\n{YELLOW}📝 What would you like to name your audio file? {RESET}")
    file_rename = set_input_default(DEAFAULT_FILENAME).strip()
    print()

    if file_rename != "":
        path = os.path.join(path, f"{file_rename}.mp3")
    else:
        path = os.path.join(path, f"{title}.mp3")

    if not os.path.exists(path):

        options = {
            "format": "bestaudio/best",
            "keepvideo": False,
            "outtmpl": path,
            "no_warnings": True,
            "noplaylist": True,
            "quiet": True,
        }

        with youtube_dl.YoutubeDL(options) as ydl:
            ydl.download([webpage_url])

        # print(f"{GREEN}Successfully downloaded {title}! Enjoy your audio!{RESET}")
        print(
            f"{GREEN}🎉 Download Complete! '{title}' has been successfully saved. Enjoy your audio experience! 🎧{RESET}"
        )

    else:
        print(
            f"{RED}A file already exists at the specified location: ({title}.mp3){RESET}"
        )


def set_input_default(text: str):
    """Set default text for readline input."""

    def startup_hook():
        readline.insert_text(text)

    readline.set_startup_hook(startup_hook)

    try:
        return input()
    finally:
        readline.set_startup_hook(None)


class NullLogger:
    """Custom logger to suppress all yt-dlp output."""

    def debug(self, msg):
        pass

    def info(self, msg):
        pass

    def warning(self, msg):
        pass

    def error(self, msg):
        pass


def fetch_video_info(url):
    """Fetch video title, webpage URL, and duration without downloading."""
    print(f"{RESET}Fetching video info...")
    try:
        ydl_opts = {
            "quiet": True,
            "no_warnings": True,
            "noplaylist": True,
            "logger": NullLogger(),
        }
        with youtube_dl.YoutubeDL(ydl_opts) as ydl:
            info = ydl.extract_info(url, download=False)
            duration = info.get("duration", None)  # type: ignore # Duration in seconds
            return info["title"], info["webpage_url"], duration  # type: ignore
    except Exception:
        return None, None, None


def print_ascii_art(text):
    # Construct the command
    command = ["toilet", "-t", "-f", "smblock", "-F", "border", "-F", "metal", text]

    # Run the command
    try:
        subprocess.run(command, check=True)
    except subprocess.CalledProcessError as e:
        print(f"An error occurred: {e}")


if __name__ == "__main__":
    try:
        get_mp3()
    except KeyboardInterrupt:
        print(f"\n{RED}Exiting SoundSnatch...{RESET}")
        sys.exit(1)
