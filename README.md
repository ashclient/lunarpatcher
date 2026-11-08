# LunarPatcher

An independent Windows desktop app for adjusting the **Lunar Client launcher**. Not affiliated with Lunar Client.

## Appearance modes

- **Ads Only:** hide supported ad placements, preserve the Lunar homepage layout.
- **Clean Layout:** use the full friends list, move news to Updates, and add Quick Play and shortcuts beneath Launch.
- **Sakura:** Clean Layout plus the fixed cherry-blossom theme and matching panels.
- **Custom Color:** Clean Layout with user-selected color, independent of Sakura.

The patch engine uses guarded renderer transforms and refuses unrecognized layout structures rather than rewriting an incompatible launcher. Backups and restoration are verified. Launcher updates may require a new patch engine. The original Lunar executable's digital signature is invalidated when patched.

## Optional Roblox Hub (Windows)

An off-by-default sidebar integration inside Lunar.

![Roblox Hub home (browser mock of injected renderer)](docs/roblox-hub-home.png)

![Discover (browser mock of injected renderer)](docs/roblox-hub-discover.png)
 Version 1.4.2 adds **real Roblox experience icons** from the official, unauthenticated thumbnail API. When you apply a patch with Roblox enabled, the engine fetches curated game icons by place ID, validates the returned image hosts and media types, and embeds the resulting images into Lunar’s ASAR. The UI also uses real icons on game cards, the featured experience and game details, with fallback artwork for unavailable thumbnails. Roblox requires internet access when the images are fetched; the patch still applies offline with fallbacks. No Roblox credentials or cookies are requested.

Version 1.4.1 fixes **React error #137** (void inputs previously received `children` from the injected helper) and introduces a more complete, Roblox-style navigation screen rather than a form-only page.

- **Home:** featured experience, continue playing, saved games and recommendations.
- **Discover:** local curated experiences, categories, search, individual game detail screens, one-click official Player launching.
- **Library:** save games from links and place IDs; view recent and saved experiences.
- **Friends, Avatar, Marketplace:** navigation hubs that hand authenticated functions to official Roblox web pages. These sections cannot show the user's live account data without an authorized Roblox integration.
- **Accounts:** locally stored username bookmarks, up to five. Real sign-in and switching take place on Roblox.com, never through collected credentials.
- **Settings:** launch mode, website fallback, history, confirmation, and official installation/support links.

Roblox website pages cannot simply be cloned into an iframe: framing and authentication restrictions apply. The in-Lunar native catalog is a small *curated* list, not a live copy of Roblox recommendations, and actual game thumbnails are best-effort, cached from Roblox’s official CDN when a patch is applied. The UI never collects session cookies, passwords or credentials and does not modify Roblox executable files or FastFlags.

**Regression checks:** a test traverses rendered React element trees for void-element children (the exact cause of #137), then tests all eight hub pages, game launch links, saving, private invitations, account bookmarks and settings. A browser visual smoke test checks each page. Both known Lunar renderer fixtures undergo patch/parse validation. Live Windows launch inside Lunar remains unverified.

## Roblox desktop launcher (v1.5.0)

The **Roblox** page in LunarPatcher now sends a validated game request to Windows using the officially registered `roblox://experiences/start?placeId=...` protocol. It is **not** just a link from the Lunar renderer:

- Detects Roblox/Roblox Player Windows protocol handlers and installed client files.
- Opens Roblox Player itself using `roblox-player:1+launchmode:app` when registered.
- Accepts Roblox experience links or numeric place IDs, private-server invitations, and an explicit browser launch mode.
- If a direct protocol is unavailable or rejected, opens the **official Roblox game page** and reports the fallback.
- Stores up to 40 saved experiences and 16 recent *launch requests*, with no credentials or cookies.
- Includes launch diagnostics, a live result log, and a safe local URL validation layer.
- Supports Compact, Comfortable, and Spacious hub layouts, an accent color, confirmation, and history settings.
- Adds a restricted `lunarpatcher://launch?placeId=...` Windows protocol so the optional in-Lunar Roblox page can hand a game to the desktop launcher. Depending on Lunar's external-link policy this handoff may be blocked; direct Roblox links and the browser remain fallback options.

**FPS and graphics:** Roblox's supported Windows client settings include an in-game maximum frame rate up to 240 FPS. The FPS selector in LunarPatcher is a **preference/reminder**, not an FPS unlocker: it cannot forcibly change Roblox's in-game FPS limit or exceed it. Graphics quality, performance stats (Shift+F5) and fullscreen are configured inside Roblox. The app does not patch Roblox FastFlags, anti-cheat, account sessions, or binaries. Multiple simultaneous Roblox instances are not promised.

**Limit of testing:** Windows protocol handoff and fallback behavior pass mock tests. The Electron UI passes a Chromium smoke test, and the Go engine passes tests for known 3.7.31/3.7.32 renderer patterns. Windows actually joining a Roblox game requires real-device confirmation; a successful OS handoff alone does not prove the game session started.

## Desktop application

The Electron UI is organized like Windows Settings: Appearance, Roblox, Updates, and About. It follows Windows light/dark system preference and keeps Apply/Restore accessible in a persistent action bar. It uses a secure preload bridge with `contextIsolation`, sandboxing, and a restrictive content security policy.

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
node tests/roblox_launch_smoke.cjs
python tests/ui_smoke.py
python tests/roblox_desktop_ui_smoke.py
npm install
npm run build:engine:win
npm run dist:win
```

The UI smoke script uses mocked IPC APIs and tests initial status, four modes, color isolation, Apply, Restore confirmation, update download/install state, responsive layout, and light/dark presentation. It **does not run Lunar Client itself**. The Node updater smoke test executes the actual Electron main-process update wiring with controlled mocks. GitHub Actions runs these tests before packaging and publishing an NSIS installer with `latest.yml` metadata to the Releases feed.

## Safety and compatibility

Close Lunar Client before installing or restoring a patch. Keep the verified backups. The 1.2.1 home-tools compatibility fix is preserved in the engine. Support for the user's uploaded 3.7.32 renderer has been tested at source/fixture level, but launcher runtime and silent Windows update installation must be verified on Windows.

This is an independent community project and may conflict with Lunar Client's terms. It never handles or requests account credentials.
