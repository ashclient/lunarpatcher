"""Real Chromium trusted-click smoke test of the narrowly scoped Roblox Player bridge.
The site is intercepted locally; no Roblox credentials or real game client needed.
"""
import re
import sys
from pathlib import Path
from playwright.sync_api import sync_playwright

source=(Path(__file__).parents[1]/'engine/assets/roblox_bridge_main.js').read_text()
match=re.search(r'const playHookSource = String\.raw`([\s\S]*?)`;',source)
assert match, 'BrowserView launch hook missing'
hook=match.group(1)
page_markup='''<!doctype html><html><body>
  <button id="game-detail-play-button" aria-label="Play"><svg aria-label="Play"></svg></button>
  <button id="install" aria-label="Download Roblox">Download Roblox</button>
  </body></html>'''
with sync_playwright() as playwright:
    kwargs={'headless':True}
    if sys.platform != 'win32': kwargs['executable_path']='/usr/bin/chromium'
    browser=playwright.chromium.launch(**kwargs)
    page=browser.new_page()
    # The sandbox blocks roblox.com network navigations. Substitute only the
    # test pathname; the click-handler logic and DOM events are unchanged.
    page.set_content(page_markup)
    page.evaluate('window.__testGamePath="/games/80959592833753/Stack-A-Butter-Squishy"')
    hook=hook.replace('location.pathname','window.__testGamePath')
    page.evaluate('window.__launched=[]; window.open=(url)=>window.__launched.push(url); true')
    page.evaluate(hook)
    page.locator('#install').click()
    assert page.evaluate('window.__launched') == [], 'Download action was incorrectly intercepted'
    page.locator('#game-detail-play-button').click()
    assert page.evaluate('window.__launched') == ['roblox://experiences/start?placeId=80959592833753']
    page.evaluate('window.__testGamePath="/home"')
    page.locator('#game-detail-play-button').click()
    assert len(page.evaluate('window.__launched')) == 1, 'Click on Home should not launch Player'
    browser.close()
print('PASS: actual trusted Chromium Play click launches selected experience; Download/Home do not')
