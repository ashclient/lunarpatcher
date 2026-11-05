# LunarPatcher

An independent, open-source Windows application for customizing the Lunar Client **launcher**. Not affiliated with Lunar Client.

## Features

LunarPatcher is a compact Electron desktop app (single settings window with native Windows frame). The modes are:

- **Ads Only:** hide recognized launcher ad placements, keeping Lunar's homepage layout.
- **Clean Layout:** remove supported ads; restore the actual full friends list to the right; show the real server Quick Play panel and version/server/mod shortcuts beneath Launch; move the existing news feed to a dedicated Updates route.
- **Sakura:** Clean Layout with cherry-blossom wallpaper and pink panels; always uses the fixed Sakura artwork, independent of the color picker.
- **Custom Color:** Clean Layout with a chosen color and no Sakura art.

The **Updates** command in LunarPatcher checks the app's own GitHub Releases for new installers. It does not automatically bypass compatibility checks when Lunar changes.

## Install and update

Download `LunarPatcher_Setup_<version>.exe` from [Releases](https://github.com/ashclient/lunarpatcher/releases). The Windows installer creates Start Menu and desktop shortcuts. The installed app uses `electron-updater` to check GitHub Releases and download subsequent versions; it prompts to restart and install. This is an unsigned community application, so Windows SmartScreen may warn.

Close Lunar Client completely before applying a patch or restoring files. The original EXE and ASAR are backed up, verified and restored together. The patch invalidates the launcher executable's existing digital signature and may be incompatible with Lunar terms or future updates. It does **not** alter accounts, purchases or game runtime modules.

## Compatibility

The patch engine checks for recognizable ad and home-layout signatures before touching files. Full friends, Quick Play, and the Updates route are verified against the `3.7.31-ow` and `3.7.32-ow` renderer structures in the test fixtures; it refuses unknown layouts. The uploaded `3.7.32-ow` file was already modified by an earlier patch, so the tests reconstruct the earlier changes only inside memory and are not a live Windows launcher test. Always keep original backups.

## Development

Requires Go 1.23+ and Node.js 22+.

```powershell
cd engine
go test ./...
go build -o ..\build\bin\LunarPatcherEngine.exe .
cd ..
npm install
npm start
```

To build the Windows NSIS installer:

```powershell
npm run dist:win
```

GitHub Actions performs Go tests, JS syntax checks, builds the Go helper and publishes a Windows setup installer when the version changes. The installer includes the helper in Electron `extraResources`.

## Architecture

- `engine/`: version-aware ASAR rebuild, Electron integrity-header update, backups, rollback, compatibility checks, offline Go tests.
- `engine/home_layout.go`: guarded React JSX transforms for friends list, Quick Play, shortcuts and Updates page. Does not replace launch controls.
- `electron/renderer/`: compact Electron UI with separate Sakura and Custom Color modes.
- `electron/main.cjs`: secure Electron main process and update events.
- `.github/workflows/`: Windows CI and release packaging.

The app currently checks its own updates, not Lunar's. Patching should be reapplied after a supported Lunar update.
