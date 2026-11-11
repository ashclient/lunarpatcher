"""Render the actual injected workspace markup in Chromium with a simulated Roblox BrowserView."""
from pathlib import Path
from playwright.sync_api import sync_playwright
import json, re
root=Path(__file__).resolve().parents[1]
source=(root/'engine/assets/roblox_widget.js').read_text()
go=(root/'engine/roblox_layout.go').read_text()
css=go.split('const robloxCSS = `',1)[1].split('`\n',1)[0]
script='''
window.__JSX__={jsx:(type,props)=>({type,props:props||{}}),jsxs:(type,props)=>({type,props:props||{}})};
window.__REACT__={useState:(initial)=>[typeof initial==='function'?initial():initial,()=>{}],useEffect:()=>{},useRef:()=>({current:null})};
window.lunarPatcherRobloxBrowser={onState:()=>()=>{},show:()=>Promise.resolve(),bounds:()=>Promise.resolve(),hide:()=>Promise.resolve(),navigate:()=>Promise.resolve(),external:()=>Promise.resolve()};
eval('var '+window.__SOURCE__+'dW=()=>{};');
const tree=lunarPatcherRobloxPage();
const attrMap={className:'class',htmlFor:'for'};
function node(item,parentNS){
 if(item==null||typeof item==='boolean')return document.createTextNode('');
 if(typeof item==='string'||typeof item==='number')return document.createTextNode(String(item));
 const svg=parentNS==='http://www.w3.org/2000/svg'||['svg','path','rect','circle','polygon'].includes(item.type);
 const ns=svg?'http://www.w3.org/2000/svg':null;
 const el=ns?document.createElementNS(ns,item.type):document.createElement(item.type);
 const p=item.props||{};
 for(const [key,val] of Object.entries(p)){
  if(key==='children'||key==='ref'||key==='key'||key.startsWith('on')||val==null||typeof val==='function')continue;
  if(key==='value')el.value=val;
  else if(typeof val==='boolean'){if(val)el.setAttribute(key,'');}
  else el.setAttribute(attrMap[key]||key,String(val));
 }
 for(const child of (Array.isArray(p.children)?p.children:[p.children]))if(child!==undefined)el.append(node(child,ns));
 return el;
}
document.getElementById('root').append(node(tree));
const el=document.querySelector('.lpr-browser-viewport');
if(el){el.innerHTML='<div style="height:100%;display:flex;align-items:center;justify-content:center;background:radial-gradient(circle at 55% 40%,#242934,#111318 65%);font-family:Arial;color:#9da5b5;text-align:center"><div><div style="font-size:44px;color:#cfd4e0;margin-bottom:10px">◇</div><div style="font-size:16px;font-weight:650;color:#e3e7f1">Roblox.com</div><div style="font-size:12px;margin-top:8px;max-width:290px">The official Roblox website occupies this area inside Lunar Client.</div></div></div>';}
'''
with sync_playwright() as p:
 browser=p.chromium.launch(headless=True,executable_path="/usr/bin/chromium",args=["--no-sandbox"])
 for width,height,out in [(1280,800,Path('/mnt/data/LunarPatcher_Roblox_FullBleed_Preview.png')),(840,690,Path('/mnt/data/LunarPatcher_Roblox_FullBleed_Narrow.png'))]:
  page=browser.new_page(viewport={'width':width,'height':height},device_scale_factor=1)
  page.set_content('<!doctype html><html><head><style>html,body,#root{height:100%;margin:0;background:#111318;font-family:Arial,sans-serif}'+css+'</style></head><body><div id="root"></div></body></html>')
  page.evaluate('window.__SOURCE__='+json.dumps(source))
  page.evaluate(script)
  assert page.locator('.lpr-navigation').count()==0
  assert page.locator('.lpr-browser-toolbar').count()==0
  assert page.locator('.lpr-search').count()==0
  assert page.locator('.lpr-browser-viewport').count()==1
  bounds=page.locator('.lpr-browser-viewport').bounding_box()
  assert bounds and bounds['width'] >= width-2 and bounds['height'] >= height-2
  page.screenshot(path=str(out),full_page=True)
  print(out.name, out.stat().st_size)
  page.close()
 browser.close()
