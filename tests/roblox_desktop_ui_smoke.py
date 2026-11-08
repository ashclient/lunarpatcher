"""Exercise Roblox desktop launch UI with a mocked Electron IPC bridge."""
from pathlib import Path
from playwright.sync_api import sync_playwright
import re
ROOT=Path(__file__).resolve().parents[1]
HTML=ROOT/'electron/renderer/index.html'
BRIDGE=r'''
window._rbCalls=[];
window.lunarPatcher={
 robloxStatus:async()=>({platform:'win32',protocol:true,playerProtocol:true,playerExe:'C:\\Roblox\\Player\\RobloxPlayerBeta.exe',reason:'Roblox deeplinks registered'}),
 robloxPreferences:async()=>({launchMode:'direct',fps:'default',layout:'comfortable',accent:'#E5A1BB',showDetails:true,confirmLaunch:false,keepHistory:true,favorites:[],recent:[]}),
 saveRobloxPreferences:async x=>{window._rbCalls.push(['save',x]);return JSON.parse(JSON.stringify(x));},
 launchRoblox:async x=>{window._rbCalls.push(['launch',x]);return {handedOff:true,method:x.mode==='browser'?'browser':'protocol',message:'Windows accepted the Roblox launch link',url:x.app?'roblox-player:1+launchmode:app':('roblox://experiences/start?placeId='+x.target)}}
};
'''

def main():
  with sync_playwright() as p:
    browser=p.chromium.launch(headless=True,args=['--no-sandbox','--disable-dev-shm-usage'], executable_path='/usr/bin/chromium' if Path('/usr/bin/chromium').exists() else None)
    page=browser.new_page(viewport={'width':1080,'height':790})
    errors=[];page.on('pageerror',lambda e:errors.append(str(e)))
    html=HTML.read_text()
    html=re.sub(r'<meta http-equiv="Content-Security-Policy"[^>]*>', '', html)
    html=re.sub(r'<script src="(?:app|roblox).js" defer></script>', '', html)
    html=re.sub(r'<link rel="stylesheet"[^>]*>', '', html)
    page.set_content(html)
    page.add_style_tag(content=(ROOT/'electron/renderer/style.css').read_text())
    page.evaluate(BRIDGE)
    page.locator('#appearancePage').evaluate('(el)=>{el.hidden=true}')
    page.locator('#robloxPage').evaluate('(el)=>{el.hidden=false}')
    page.add_script_tag(content=(ROOT/'electron/renderer/roblox.js').read_text())
    page.evaluate("document.dispatchEvent(new Event('DOMContentLoaded'))")
    page.wait_for_function("document.querySelector('#robloxHealth').textContent.includes('ready')")
    assert page.locator('#robloxDiagnostics').is_hidden()
    assert page.locator('#rbProtocol').inner_text()=='Registered'
    assert page.locator('#robloxLaunchForm').is_visible()
    page.locator('#robloxTarget').fill('https://www.roblox.com/games/2753915549/Blox-Fruits')
    page.locator('#robloxSaveGame').click()
    assert page.locator('#robloxFavorites .saved-row').count()==1
    page.locator('#robloxLaunchForm button[type=submit]').click()
    page.wait_for_function("window._rbCalls.some(c=>c[0]==='launch' && c[1].target.includes('2753915549'))")
    page.locator('[data-roblox-panel=performance]').click()
    assert page.locator('#robloxPerformance').is_visible()
    page.locator('#robloxFps').select_option('240')
    assert '240 FPS' in page.locator('#robloxFpsGuide').inner_text()
    page.locator('[data-roblox-panel=customize]').click()
    page.locator('#robloxLayout').select_option('compact')
    assert 'compact' in page.locator('#robloxLayoutPreview').get_attribute('class')
    page.locator('#robloxAccent').fill('#ff99aa')
    assert '#ff99aa' in page.locator('#robloxLayoutPreview').get_attribute('style')
    page.locator('#robloxMode').select_option('browser')
    page.locator('[data-roblox-panel=diagnostics]').click()
    assert page.locator('#robloxDiagnostics').is_visible()
    page.locator('#robloxTestLaunch').click()
    page.wait_for_function("window._rbCalls.some(c=>c[0]==='launch' && c[1].app===true)")
    page.locator('[data-roblox-panel=games]').click()
    page.locator('#viewport').evaluate('(el)=>{el.scrollTop=0}')
    page.locator('#topbarTitle').evaluate('(el)=>{el.textContent="Roblox"}')
    page.locator('#actionBar').evaluate('(el)=>{el.hidden=true}')
    page.locator('[data-view=appearance]').evaluate('(el)=>{el.classList.remove("current")}')
    page.locator('[data-view=roblox]').evaluate('(el)=>{el.classList.add("current")}')
    page.screenshot(path=str(ROOT/'test-output/roblox-desktop.png'),full_page=True)
    page.set_viewport_size({'width':785,'height':740})
    assert page.evaluate('document.documentElement.scrollWidth <= innerWidth')
    page.screenshot(path=str(ROOT/'test-output/roblox-desktop-narrow.png'),full_page=True)
    assert not errors,errors
    browser.close()
  print('PASS Roblox UI: detection, library, OS launch, FPS guide, layout, accent, diagnostics, narrow viewport, runtime errors')
if __name__=='__main__': main()
