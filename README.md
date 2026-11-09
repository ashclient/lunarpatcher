# LunarPatcher

An unofficial Windows customizer for Lunar Client. Version **1.7.0**.

LunarPatcher is a **patch-only application**. Select an appearance, optionally enable Roblox, click Apply, and then close LunarPatcher. Roblox browsing and game launching happen **inside Lunar Client**, not in the patcher.

## Features

- Ads Only, Clean Layout, Sakura, and Custom Color modes
- Full friends panel, Quick Play, and Updates page in the clean layout
- Optional **Roblox** workspace in Lunar: an integrated navigation sidebar (Home, Discover, Friends, Messages, Avatar, Marketplace, Communities, Trades, and Settings), official Roblox website browsing, game search, and locally saved page shortcuts
- Back, Forward, Reload, Home, URL, and Open in Browser controls in Lunar
- Official Roblox Player protocol handoff for game launches, via Windows
- Verified backups, restoration, version-aware safety checks, and GitHub Releases auto-updates for LunarPatcher

## Setup

1. Install LunarPatcher from the GitHub Releases page.
2. Fully close Lunar Client (including the system tray).
3. Choose a mode and enable **Roblox website in Lunar** if desired.
4. Click **Apply changes**, then launch Lunar from its normal shortcut.
5. Select the Roblox icon on **Lunar's left sidebar** to browse the actual Roblox website.

**To switch themes or disable Roblox**, close Lunar and reapply. Restore Original returns the verified original launcher files. Patching Lunar's signed EXE invalidates its original signature.

## Roblox browser architecture

The installer modifies the ASAR with a small React route, an isolated Electron `BrowserView` in Lunar's main process, and a restricted preload IPC bridge. We use a native browser surface rather than an iframe; ordinary iframes may be blocked by Roblox's framing restrictions. The real site controls its own features and login flow.

- The remote Roblox page runs with `nodeIntegration:false`, `contextIsolation:true`, `sandbox:true`, and `webSecurity:true`.
- The browser is isolated in the `persist:lunarpatcher-roblox` session. LunarPatcher never collects Roblox cookies, credentials, or account data.
- Only HTTPS URLs on `roblox.com` and its subdomains are allowed in this view. Roblox Player protocol requests use Windows' registered handler. No external programs or DLLs are injected.
- Browser permissions (camera, microphone, etc.) default to denied. Some Roblox features may need the system browser.
- The browser view detaches when leaving the Roblox tab. Opening the external-browser button is always available as a fallback.
- Roblox's website can change its behavior or decline an embedded login. Game handoff is not guaranteed until tested with an installed Roblox Player on Windows.

## Development

Requirements: Go, Node.js 22+, and Windows for the released installer.

```bash
cd engine && go test ./...
cd .. && npm run test:smoke
```

The GitHub Actions release job builds an Electron NSIS installer when a new commit is pushed to `main`, and publishes its `latest.yml` for auto-updates.

## Limitations

This is unofficial and is not affiliated with Roblox or Lunar Client. The patcher depends on internal Electron/React implementation details and deliberately refuses unknown signatures. Only the available 3.7.31 and 3.7.32 renderer fixtures have been tested locally. Real Windows runtime behavior, Roblox sign-in, and game joining still require installation testing.

## Lunar for Roblox reference

The third-party repository [`chiikabu/lunar-for-roblox-decompiled`](https://github.com/chiikabu/lunar-for-roblox-decompiled) was used **only as a functional feature inventory**. Its README identifies the source as proprietary Moonsworth/Overwolf material. No decompiled application code, native binaries, proprietary icons, or assets are bundled or copied.

This version is **not** a 1:1 recreation of Lunar for Roblox 1.0.16. It preserves the Minecraft Lunar Client launcher and adds an original Roblox workspace that works through official Roblox pages. The decompiled application is a standalone Electron product with its own IPC controllers, backend services, native components, and account/session infrastructure; directly transplanting those into Minecraft Lunar is not equivalent to adding a sidebar route.

Unsupported here: native isolated Roblox client instances, separate authenticated account sessions, FastFlags/FPS modifications, macros, native avatar/marketplace checkout, capture/overlay services and custom game versions. No fake switches for these features are displayed. Roblox login occurs through the official Roblox website in an isolated browser session. The OS-installed Roblox Player is responsible for game joins.
