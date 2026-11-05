# LunarPatcher

An independent Windows desktop app for adjusting the **Lunar Client launcher**. Not affiliated with Lunar Client.

## Appearance modes

- **Ads Only:** hide supported ad placements, preserve the Lunar homepage layout.
- **Clean Layout:** use the full friends list, move news to Updates, and add Quick Play and shortcuts beneath Launch.
- **Sakura:** Clean Layout plus the fixed cherry-blossom theme and matching panels.
- **Custom Color:** Clean Layout with user-selected color, independent of Sakura.

The patch engine uses guarded renderer transforms and refuses unrecognized layout structures rather than rewriting an incompatible launcher. Backups and restoration are verified. Launcher updates may require a new patch engine. The original Lunar executable's digital signature is invalidated when patched.

## Desktop application

The Electron UI is organized like Windows Settings: Appearance, Updates, and About. It follows Windows light/dark system preference and keeps Apply/Restore accessible in a persistent action bar. It uses a secure preload bridge with `contextIsolation`, sandboxing, and a restrictive content security policy.

### In-app auto-updates (no Downloads folder)

The first installation uses the Windows installer in [GitHub Releases](https://github.com/ashclient/lunarpatcher/releases). **Subsequent application updates do not require you to download or run setup again.** On installed builds, `electron-updater`:

1. Checks GitHub Releases shortly after startup and every six hours while running.
2. Downloads available updates to its **internal updater cache**, and verifies against release metadata/checksums; no dependency on a file in the user's Downloads folder.
3. Shows progress on the Updates page. When ready, **Restart to update** silently applies the cached NSIS update and restarts the app. A downloaded update can also install on normal application quit.

If an installer is blocked by Windows, the Updates page reports the error; a manual installer from Releases remains available as a fallback. LunarPatcher updates are **not** automatic Lunar Client launcher patches. Windows installers are currently unsigned, so SmartScreen can warn.

## Development

Requirements: Go 1.23+, Node.js 22+, Python 3 with Playwright for visual smoke testing.

```bash
cd engine && go test ./... && cd ..
node tests/updater_smoke.cjs
python tests/ui_smoke.py
npm install
npm run build:engine:win
npm run dist:win
```

The UI smoke script uses mocked IPC APIs and tests initial status, four modes, color isolation, Apply, Restore confirmation, update download/install state, responsive layout, and light/dark presentation. It **does not run Lunar Client itself**. The Node updater smoke test executes the actual Electron main-process update wiring with controlled mocks. GitHub Actions runs these tests before packaging and publishing an NSIS installer with `latest.yml` metadata to the Releases feed.

## Safety and compatibility

Close Lunar Client before installing or restoring a patch. Keep the verified backups. The 1.2.1 home-tools compatibility fix is preserved in the engine. Support for the user's uploaded 3.7.32 renderer has been tested at source/fixture level, but launcher runtime and silent Windows update installation must be verified on Windows.

This is an independent community project and may conflict with Lunar Client's terms. It never handles or requests account credentials.
