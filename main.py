import os
import yt_dlp as youtube_dl

def get_mp3():
    try:
        # Configure yt-dlp to suppress verbose output and disable playlist processing
        ydl_opts = {
            'quiet': True,           # Suppress most output
            'no_warnings': True,     # Suppress warnings (e.g., HTTP 500 retries)
            'noplaylist': True,      # Only download the single video, not playlists
        }
        with youtube_dl.YoutubeDL(ydl_opts) as ydl:
            print("Fetching video info...")
            video_info = ydl.extract_info(
                url=input("Enter video URL: "), 
                download=False
            )

        path = ""
        path_ = input("Select Destination: ").strip()
        if path_ == "":
            path = "~/Music"
        else:
            path = path_
            
        print("Selected Path: ", path)
        file_rename = input(f"Rename File? {video_info['title']}: ")
        if file_rename != "":
            path = path + "/" + file_rename + ".mp3"
        else:
            path = path + f"/{video_info['title']}.mp3"

        # Ensure the directory exists
        expanded_path = os.path.expanduser(path)
        os.makedirs(os.path.dirname(expanded_path), exist_ok=True)

        # Download options
        options = {
            'format': 'bestaudio/best',
            'keepvideo': False,
            'outtmpl': expanded_path,
            'quiet': True,
            'no_warnings': True,
            'noplaylist': True,
        }
        print("Final audio file path: ", expanded_path)
        print("Downloading started...")
        with youtube_dl.YoutubeDL(options) as ydl:
            ydl.download([video_info['webpage_url']])
        print("Download completed!")

    except Exception as e:
        print(f"Error: {str(e)}")

if __name__ == "__main__":
    get_mp3()
