import os

import yt_dlp as youtube_dl

# from dearpygui.dearpygui import *
from dearpygui.dearpygui import (
    add_button,
    add_dummy,
    add_input_text,
    add_spacer,
    add_spacing,
    add_text,
    create_context,
    create_viewport,
    destroy_context,
    get_value,
    render_dearpygui_frame,
    set_primary_window,
    set_value,
    setup_dearpygui,
    show_viewport,
    start_dearpygui,
    window,
)

from downloader import download_audio, fetch_video_info

# Default download folder
DEFAULT_PATH = os.path.expanduser("~/Music")


# -------------------------
# Callbacks
# -------------------------
def fetch_info_callback(sender, app_data, user_data):
    url = get_value("url_input").strip()
    if not url:
        set_value("status_text", "[Error] Enter a YouTube URL.")
        return

    title, webpage_url, duration, success = fetch_video_info(url)
    if not success:
        set_value(
            "status_text",
            "[Error] Could not fetch video info. Check the URL or your network.",
        )
        return

    set_value(
        "status_text",
        f"Title: {title}\nDuration: {duration} sec\nReady to download.",
    )
    set_value("title_input", title)
    set_value("download_path", DEFAULT_PATH)


def download_callback(sender, app_data, user_data):
    url = get_value("url_input").strip()
    path = get_value("download_path").strip()
    filename = get_value("title_input").strip()

    if not url or not path or not filename:
        set_value("status_text", "[Error] URL, path, or filename missing!")
        return

    full_path = os.path.join(path, f"{filename}.mp3")
    if os.path.exists(full_path):
        set_value("status_text", f"[Error] File already exists: {full_path}")
        return

    set_value("status_text", "Downloading...")
    render_dearpygui_frame()  # force update

    success = download_audio(url, full_path)
    if success:
        set_value("status_text", f"🎉 Download complete: {full_path}")
    else:
        set_value("status_text", "[Error] Download failed!")


# -------------------------
# GUI Setup
# -------------------------
create_context()

# Main window
with window(tag="main_window"):
    add_text(
        tag="title_text",
        default_value="Download Songs, Podcasts, and Audio from YouTube 🎵",
    )
    add_spacer(height=2)
    add_input_text(tag="url_input", label="YouTube URL", width=500)
    add_spacer(height=2)
    add_button(tag="fetch_info_btn", label="Fetch Info", callback=fetch_info_callback)
    add_spacer(height=1)
    add_input_text(tag="title_input", label="Filename", width=500)
    add_input_text(
        tag="download_path", label="Save Path", default_value=DEFAULT_PATH, width=500
    )
    add_spacer(height=2)
    add_button(tag="download_btn", label="Download Audio", callback=download_callback)
    add_spacer(height=2)
    add_text(tag="status_label", default_value="Status:", bullet=True)
    add_text(tag="status_text", default_value="", wrap=500)

# Create viewport instead of old add_window()
create_viewport(title="SoundSnatch - GUI", width=600, height=400)
setup_dearpygui()
show_viewport()
set_primary_window("main_window", True)
start_dearpygui()
destroy_context()
