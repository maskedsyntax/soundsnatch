import yt_dlp as youtube_dl


class NullLogger:
    def debug(self, msg):
        pass

    def info(self, msg):
        pass

    def warning(self, msg):
        pass

    def error(self, msg):
        pass


def fetch_video_info(url):
    try:
        ydl_opts: dict = {
            "quiet": True,
            "no_warnings": True,
            "noplaylist": True,
            "logger": NullLogger(),
        }
        with youtube_dl.YoutubeDL(ydl_opts) as ydl:
            info = ydl.extract_info(url, download=False)
            title = info.get("title", "Unknown Title")
            webpage_url = info.get("webpage_url", url)
            duration = info.get("duration", 0)
            return title, webpage_url, duration, True
    except Exception:
        return None, None, None, False


def download_audio(url, path):
    try:
        options: dict = {
            "format": "bestaudio/best",
            "keepvideo": False,
            "outtmpl": path,
            "no_warnings": True,
            "noplaylist": True,
            "quiet": True,
        }
        with youtube_dl.YoutubeDL(options) as ydl:
            ydl.download([url])
        return True
    except Exception:
        return False
