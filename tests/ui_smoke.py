"""Renderer-only Chromium smoke checks for the Electron app (IPC is mocked).
Run: python tests/ui_smoke.py
"""
from pathlib import Path
from playwright.sync_api import sync_playwright

ROOT=Path(__file__).resolve().parents[1]
HTML_PATH='electron/renderer/index.html'
OUTPUT=ROOT/'test-output'
OUTPUT.mkdir(exist_ok=True)

BRIDGE="""
window._bridgeCalls = [];
window._onUpdate = null;
window.lunarPatcher = {
  getPreferences: async () => ({folder:'C:\\\\Users\\\\Tester\\\\AppData\\\\Local\\\\Programs\\\\Lunar Client',mode:'sakura',color:'#E59CBC'}),
  getAppInfo: async () => ({version:'1.3.0',defaultFolder:'C:\\\\Lunar Client',update:{phase:'current',message:'You are running the latest version'}}),
  getStatus: async () => ({found:true,version:'3.7.32-ow',mode:'sakura',canRestore:true,changedSincePatch:false}),
  check: async (...args) => {window._bridgeCalls.push(['check',...args]);return {ok:true,version:'3.7.32-ow'}},
  apply: async (...args) => {window._bridgeCalls.push(['apply',...args]);return {ok:true}},
  restore: async (...args) => {window._bridgeCalls.push(['restore',...args]);return {ok:true}},
  getPreferences: async () => ({folder:'C:\\\\Users\\\\Tester\\\\AppData\\\\Local\\\\Programs\\\\Lunar Client',mode:'sakura',color:'#E59CBC'}),
  savePreferences: async (...args) => {window._bridgeCalls.push(['save',...args]);return {}},
  pickFolder: async () => 'C:\\\\Lunar Client',
  checkUpdates: async () => ({phase:'checking',message:'Checking now'}),
  installUpdate: async () => {window._bridgeCalls.push(['update-install']); return true},
  openRepository: async () => true,
  onUpdate: fn => {window._onUpdate=fn;return ()=>{window._onUpdate=null}}
};
"""

def go():
  with sync_playwright() as p:
    launch={'headless':True,'args':['--no-sandbox','--disable-dev-shm-usage']}
    if Path('/usr/bin/chromium').exists():launch['executable_path']='/usr/bin/chromium'
    browser=p.chromium.launch(**launch)
    page=browser.new_page(viewport={'width':1080,'height':790},device_scale_factor=1)
    failures=[]
    page.on('pageerror', lambda e:failures.append(str(e)))
    page.add_init_script(BRIDGE)
    import re, base64
    html=(ROOT/HTML_PATH).read_text()
    html=re.sub(r'<meta http-equiv="Content-Security-Policy"[^>]*>', '', html)
    html=re.sub(r'<link rel="stylesheet"[^>]*>', '', html)
    html=re.sub(r'<script src="app.js" defer></script>', '', html)
    img=base64.b64encode((ROOT/'electron/assets/sakura_wallpaper.jpg').read_bytes()).decode()
    html=html.replace('../assets/sakura_wallpaper.jpg','data:image/jpeg;base64,'+img)
    page.set_content(html)
    page.evaluate(BRIDGE)
    page.add_style_tag(content=(ROOT/'electron/renderer/style.css').read_text())
    page.add_script_tag(content=(ROOT/'electron/renderer/app.js').read_text())
    page.evaluate('initialize()')
    page.wait_for_function("document.querySelector('#statusLabel')?.textContent === 'Customized'")
    assert page.locator('#sakuraArtwork').is_visible()
    assert page.locator('.mode-option.selected').get_attribute('data-mode')=='sakura'
    assert page.locator('#colorRow').is_hidden()
    assert page.locator('#restoreBtn').is_enabled()
    page.screenshot(path=str(OUTPUT/'appearance.png'),full_page=True)

    # 2. Mode switching and color selection must not affect Sakura.
    page.locator('label[data-mode="color"]').click()
    assert page.locator('#colorRow').is_visible()
    page.locator('#colorPicker').fill('#4798a8')
    assert page.locator('#colorHex').inner_text()=='#4798A8'
    page.locator('label[data-mode="sakura"]').click()
    assert page.locator('#colorRow').is_hidden()
    assert page.locator('#themePreview').get_attribute('class')=='theme-preview sakura'
    page.locator('label[data-mode="layout"]').click()
    assert page.locator('.mode-option.selected').get_attribute('data-mode')=='layout'
    # Optional Roblox sidebar is independent of the visual theme.
    page.locator('label[for="robloxToggle"]').click()
    assert page.locator('#robloxToggle').is_checked()
    page.locator('#applyBtn').click()
    page.wait_for_function("document.querySelector('#actionMessage').textContent.includes('Changes applied')")
    calls=page.evaluate("window._bridgeCalls")
    assert any(c[0]=='check' and c[2]=='layout' for c in calls),calls
    assert any(c[0]=='apply' and c[2]=='layout' and c[4] is True for c in calls),calls

    # 3. Restore flow requires explicit confirmation.
    page.locator('#restoreBtn').click()
    assert page.locator('#restoreDialog').is_visible()
    page.locator('#confirmRestoreBtn').click()
    page.wait_for_function("document.querySelector('#actionMessage').textContent.includes('restored')")
    assert any(c[0]=='restore' for c in page.evaluate('window._bridgeCalls'))

    # 4. Update flow: UI knows about a ready update, can trigger in-app install.
    page.locator('button[data-view="updates"]').click()
    assert page.locator('#updatesPage').is_visible()
    page.evaluate("window._onUpdate({phase:'downloading',version:'1.3.1',progress:62,message:'Downloading update · 62%'})")
    assert page.locator('#updateProgress').evaluate("e => e.style.width")=='62%'
    page.evaluate("window._onUpdate({phase:'ready',version:'1.3.1',progress:100,message:'Ready to restart'})")
    assert page.locator('#installUpdateBtn').is_visible()
    assert page.locator('#updateIndicator').is_visible()
    page.screenshot(path=str(OUTPUT/'updates.png'),full_page=True)
    page.locator('#installUpdateBtn').click()
    assert any(c[0]=='update-install' for c in page.evaluate('window._bridgeCalls'))

    # 5. Narrow viewport: appearance remains accessible, no horizontal overflow.
    page.set_viewport_size({'width':785,'height':750})
    page.locator('button[data-view="appearance"]').click()
    assert page.locator('#applyBtn').is_visible()
    assert page.evaluate('document.documentElement.scrollWidth <= innerWidth')
    page.screenshot(path=str(OUTPUT/'appearance-narrow.png'),full_page=True)
    # 6. System dark mode is readable and uses the same controls.
    page.set_viewport_size({'width':1080,'height':790})
    page.emulate_media(color_scheme='dark',reduced_motion='reduce')
    page.wait_for_timeout(250)
    assert page.locator('#applyBtn').is_visible()
    assert page.evaluate('getComputedStyle(document.body).color') != 'rgb(36, 38, 49)'
    page.screenshot(path=str(OUTPUT/'appearance-dark.png'),full_page=True)
    assert not failures, failures
    browser.close()
  print('PASS: startup, mode switching, color, apply, restore, updater, narrow viewport, console errors')

if __name__=='__main__': go()
