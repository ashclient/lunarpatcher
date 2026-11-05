# LunarPatcher

An unofficial Windows desktop application for compatible Electron-based Lunar Client launcher versions. Built with Go, with a native WPF interface hosted by Windows PowerShell. It can remove recognized ad slots, apply cosmetic styles, and restore the original launcher files.

## Modes

- **Remove ads:** hide supported launcher advertisement slots.
- **Clean layout:** ad cleanup, safe styling, and removal of the homepage announcement feed when the supported launch/friends structure matches. Full friends-sidebar relocation is not supported for all builds.
- **Sakura:** fixed cherry-blossom wallpaper and pink styling, plus homepage announcement-feed cleanup.
- **Custom color:** choose a hex color for compatible launcher panels without using the Sakura wallpaper; it also removes supported homepage announcements.

The selected appearance is saved locally in `%APPDATA%\LunarPatcher\settings.json`.

## Installation

Open the latest **GitHub Release** and download `LunarPatcher_Setup.exe`. The installed app lives in your user profile and provides a Start Menu shortcut and an optional desktop shortcut. The included `.iss` definition builds a standard Inno Setup installer, not a bespoke setup wizard.

**Automatic updates:** the app checks GitHub Releases at startup. By default, it downloads the latest stable installer, verifies its SHA-256 against the release's checksum file, and launches the installer in silent mode. You can disable this in the application. It updates **LunarPatcher only**, not Lunar Client itself, and does not automatically re-patch Lunar after Lunar updates.

GitHub Actions automatically builds and publishes a Windows release when a new version number in `VERSION` is pushed to `main`. It uses `vX.Y.Z` tags, checksums, and a standard setup executable. Subsequent releases appear in the in-app update checker. This repository includes source code; no proprietary Lunar files are bundled.

## Development

Requirements: Go 1.23+; Windows 10/11 for GUI runtime; Inno Setup 6 for packaging.

```powershell
go test ./...
go build -ldflags '-H=windowsgui' -o LunarPatcher.exe .
```

CLI supports:

```text
--check --mode ads --dir "C:\...\Lunar Client"
--mode ads|layout|sakura|color --color #D676A5 --dir "C:\...\Lunar Client"
--restore --dir "C:\...\Lunar Client"
--check-update
--download-update
--version
```

## Safety and compatibility

The patcher reads local `app.asar`, locates a recognized ad-renderer guard and (for Clean Layout, Sakura, and Custom Color) a guarded homepage-news component boundary, and refuses to patch when it cannot prove compatibility. It backs up the original launcher EXE and ASAR under `.lunar-adaptive` and offers restoration. Patching changes Lunar's signed executable integrity record, invalidating the original Windows digital signature. Future Lunar versions may fail compatibility checks; **`3.7.32-master.15672/4fc0400` has not been tested against its real files**. Keep backups and close Lunar fully before applying patches. This project is not affiliated with or endorsed by Lunar Client.

The WPF GUI and Inno installer require Windows runtime testing. Use the app at your own risk.
