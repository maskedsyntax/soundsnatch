# SoundSnatch TODO

## Phase 1: UI/UX Enhancements 🚀
- [x] **Real Progress Bars**: Replace the spinner with a 0-100% progress bar using `yt-dlp` output parsing.
- [x] **Format Selection**: Add a menu to choose between MP3, FLAC, and WAV.
- [x] **Search Functionality**: Search for songs by name instead of just pasting URLs.

## Phase 2: Engineering & Distribution 🛠️
- [x] **Unit Testing**: Add `main_test.go` and implement tests for core logic.
- [x] **Binary Releases**: Set up GitHub Actions with GoReleaser for automated cross-platform builds.
- [ ] **Configuration**: Add a `.yaml` config file to save user preferences (like default download paths).

## Phase 3: Architectural Refactoring 🏗️
- [ ] **Code Organization**: Split `main.go` into dedicated packages (`ui`, `downloader`, `fs`).

## Phase 4: Final Polish ✨
- [ ] **Update Assets**: Replace `assets/soundsnatch_ss.png` with a new TUI screenshot. (User task)
