# LunarPatcher

An independent Windows application for customizing the Lunar Client launcher. **Not affiliated with, endorsed by, or supported by Lunar Client.**

## Download

Use the Windows installer from [GitHub Releases](https://github.com/ashclient/lunarpatcher/releases). The installer creates a normal Windows application with Start Menu and desktop shortcuts.

## Features

- Electron desktop app with a proper native window, navigation, installation detection, and progress/status feedback.
- **Ads Only** — disables a verified ad renderer component while keeping the default look.
- **Clean Layout** — removes compatible home-news components; keeps the launch, settings, and friends controls intact.
- **Sakura** — pink cherry-blossom theme, independent of custom colors.
- **Custom Color** — standalone color mode with its own picker.
- Version-aware validation, verified original-file backups, switch modes, and restore support.
- App checks GitHub Releases for updates at startup, downloads releases with `electron-updater`, and installs when restarted.

## Important compatibility notes

The source analysis and patch-source tests cover the `3.7.31-ow` renderer and the uploaded `3.7.32-ow` bundle. The latter was **already modified by v8.1**, so the test reconstructs known renderer changes in memory and is not a substitute for testing the unmodified Windows build. The uploader's earlier `3.7.32-master.15672/4fc0400` identifier differs from the bundle's internal `3.7.32-ow` version string.

The Go patch engine also checks the **installed Windows executable's** ASAR integrity metadata before changing anything. Updated executables, unexpected source code, missing original backups, or unsupported layouts may be refused. A successful file-level patch does not guarantee Lunar's GUI works: close Lunar and create a restore point/backup before installing.

**Do not remove the `.lunar-adaptive` directory from the Lunar Client installation** while a patch is installed; it stores the original EXE and ASAR required to restore. Do not upload those original Lunar files to this public repository.

Modifying `Lunar Client.exe` changes its Authenticode signature. Updates and security products may block patched versions, and modding may conflict with the launcher's terms. The project does not alter credentials, entitlements, game content, or subscription functionality.

## Developer setup

Requirements: Go 1.23+, Node.js 22+, Windows 10/11. 

```powershell
# From repository root
cd engine
go test ./...
go build -o ../build/bin/LunarPatcherEngine.exe .
cd ..
npm install
npm start
```

Build an installable Windows NSIS setup:

```powershell
npm run dist:win
```

The release workflow builds on Windows and publishes a new GitHub Release whenever `package.json` has a version that is not already tagged. The build also publishes `latest.yml`, needed by the app updater. To release a new version, bump the version in `package.json` and push to `main`.

## Security design

The Electron renderer uses context isolation, sandbox, disabled Node integration, a strict Content Security Policy, and a small allowlisted IPC bridge. Patching runs in a separate Go executable and refuses unfamiliar renderer structures, rather than blind string replacements. The official game launch/authentication pipeline is not modified.

## Project folders

- `electron/` — Electron main/preload processes, desktop renderer, visual assets
- `engine/` — Go ASAR parser, patch validation, theme CSS, backup/restore CLI
- `build/` — app icon; the Windows engine is generated into `build/bin`
- `.github/workflows/` — Windows CI and release automation

## License

MIT for LunarPatcher code. Lunar Client, Electron, and other product names and assets belong to their respective owners.
