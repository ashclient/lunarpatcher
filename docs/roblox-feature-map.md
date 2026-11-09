# Roblox workspace feature map

Reference: `chiikabu/lunar-for-roblox-decompiled/docs/INDEX.md` (Lunar for Roblox 1.0.16), inspected November 9, 2026. The index describes 54 application IPC controllers and 445 IPC/event channels. The original code belongs to third parties and is **not copied** into this project.

| Subsystem | Existing original | v1.7 status | Implementation |
| --- | --- | --- | --- |
| Launcher Home / Discovery | Discovery, Games, Search controllers | Integrated | Navigate official Roblox Home, Discover, Search |
| Game launch | Launch, Instances, Versions controllers | Partial | Roblox.com joins via OS registered Roblox Player |
| Friends / Messages / Communities | Friends, Chat, Communities controllers | Website-backed | Official Roblox pages in an isolated browser |
| Avatar / Catalog / Trades | Avatar, Catalog, Trades controllers | Website-backed | Official Roblox pages; no native purchase handling |
| User accounts | AuthController with account sessions | Website-backed only | Official Roblox sign-in; no secret or session extraction |
| Saved shortcuts | Game favorites | Integrated local | User-pinned Roblox pages in localStorage |
| Isolated instances / versions / FastFlags | Instances, Versions controllers | Not implemented | Requires supported Roblox runtime integration and user testing |
| Macros / mods / captures / second screen | Macros, Mods, Captures, SecondScreen controllers | Not implemented | Separate major backend work, not a feature toggle |

## Integration constraints

- The current Minecraft Lunar renderer and the leaked standalone app have different Electron main/preload lifecycles and React bundles.
- Website navigation is confined to Roblox HTTPS domains; game links may open the OS Roblox Player.
- Do not load decompiled native DLLs, copy login cookies between browsers, or import unlicensed assets.
- Do not claim an authenticated Roblox feature works without live Windows testing.
