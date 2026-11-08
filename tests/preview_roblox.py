from pathlib import Path
import re
from playwright.sync_api import sync_playwright
root=Path(__file__).resolve().parents[1]
code=(root/'engine/assets/roblox_widget.js').read_text().split('\n',1)[1].replace('__JSX__','V').replace('__REACT__','B').replace('__INITIATOR__','tp.sidebar')
code=re.sub(r',\s*$', ';',code)
css=(root/'engine/roblox_layout.go').read_text().split('const robloxCSS = `',1)[1].split('`\n',1)[0]
bootstrap="""
const hooks=[];let hookIndex=0;window.opened=[];
window.lunar={misc:{openExternalLink:async opt=>{window.opened.push(opt);return true}}};
window.tp={sidebar:'sidebar'};
window.V={jsx:(type,props)=>({type,props}),jsxs:(type,props)=>({type,props})};
window.B={useState:init=>{const i=hookIndex++;if(!(i in hooks))hooks[i]=typeof init==='function'?init():init;return [hooks[i],next=>{hooks[i]=typeof next==='function'?next(hooks[i]):next;requestAnimationFrame(render)}]}};
function dom(node){
 if(node===null||node===undefined||typeof node==='boolean')return document.createTextNode('');
 if(typeof node==='string'||typeof node==='number')return document.createTextNode(String(node));
 if(Array.isArray(node)){let f=document.createDocumentFragment();for(const n of node)f.append(dom(n));return f;}
 let {type,props={}}=node;
 if(typeof type==='function')return dom(type(props));
 let svg=['svg','path','circle','rect','g','line','polyline','polygon'].includes(type);
 let e=svg?document.createElementNS('http://www.w3.org/2000/svg',type):document.createElement(type);
 for(let [k,v] of Object.entries(props)){
   if(k==='children'||k==='key'||v===null||v===undefined)continue;
   if(k==='className'){e.setAttribute('class',v);continue;}
   if(k==='fillRule'){e.setAttribute('fill-rule',v);continue;}
   if(k==='aria-hidden'||k==='aria-selected'||k==='aria-label'||k==='title'||k==='placeholder'||k==='role'||k==='name'||k==='type'||k==='maxLength'){e.setAttribute(k==='maxLength'?'maxlength':k,String(v));continue;}
   if(k==='value'||k==='checked'){e[k]=v;continue;}
   if(k.startsWith('on')&&typeof v==='function'){e.addEventListener(k.slice(2).toLowerCase(),v);continue;}
   if(k==='style'&&typeof v==='object'){Object.assign(e.style,v);continue;}
   if(typeof v==='string'||typeof v==='number')e.setAttribute(k,String(v));
 }
 e.append(dom(props.children));return e;
}
function render(){hookIndex=0;document.querySelector('#root').replaceChildren(dom(lunarPatcherRobloxPage()));}
"""
html='''<!doctype html><meta charset="utf-8"><style>html,body{padding:0;margin:0;background:#11131a;color:#edf0f7;font:13px "Segoe UI",sans-serif;height:100%;}#root{--colors-space-1:#171920;--colors-space-2:#21242c;--colors-space-3:#292e39;--colors-space-11:#9da9bb;--colors-space-12:#f1f4f9;--colors-space-a4:rgba(180,190,210,.16);height:100%;width:100%;}button,input,select{font:inherit}button{cursor:pointer}</style><style>'''+css+'''</style><div id="root"></div><script>'''+bootstrap+'\nvar '+code+'\nrender();</script>'
preview=root/'roblox_preview.html';preview.write_text(html)
with sync_playwright() as p:
 b=p.chromium.launch(executable_path='/usr/bin/chromium',headless=True,args=['--no-sandbox','--disable-dev-shm-usage'])
 page=b.new_page(viewport={'width':1260,'height':800},device_scale_factor=1)
 page.set_content(html);page.wait_for_timeout(200)
 if page.locator('h1').count()<1:raise Exception(page.content()[-400:])
 for name,tab in [('games','Games'),('accounts','Accounts'),('settings','Settings')]:
  page.get_by_role('tab',name=tab).click();page.wait_for_timeout(80)
  page.screenshot(path=str(root/f'roblox_preview_{name}.png'))
  print(name, page.locator('body').inner_text()[:250].replace('\n',' | '))
 b.close()
