// LunarPatcher: official Roblox website hosted in a secure BrowserView within Lunar.
// No fake game catalog, account handling, passwords or external launcher app.
lunarPatcherRobloxIcon=props=>(0,__JSX__.jsx)(`svg`,{viewBox:`0 0 24 24`,width:props?.size||22,height:props?.size||22,fill:`currentColor`,"aria-hidden":!0,children:(0,__JSX__.jsx)(`path`,{fillRule:`evenodd`,d:`M7.56 1.15 22.85 7.56 16.44 22.85 1.15 16.44 7.56 1.15ZM10.23 8.45 8.45 13.77 13.77 15.55 15.55 10.23 10.23 8.45Z`})}),
lunarPatcherRobloxPage=()=>{
 const bridge=window.lunarPatcherRobloxBrowser;
 const viewport=(0,__REACT__.useRef)(null);
 const [url,setUrl]=(0,__REACT__.useState)(`https://www.roblox.com/home`);
 const [status,setStatus]=(0,__REACT__.useState)(bridge?`Loading Roblox…`:`Browser extension unavailable. Reapply the Roblox patch.`);
 const [back,setBack]=(0,__REACT__.useState)(false);
 const [forward,setForward]=(0,__REACT__.useState)(false);
 const [entry,setEntry]=(0,__REACT__.useState)(`https://www.roblox.com/home`);
 const [busy,setBusy]=(0,__REACT__.useState)(true);
 const run=(name,...args)=>{
  if(!bridge||typeof bridge[name]!==`function`)return;
  Promise.resolve(bridge[name](...args)).catch(error=>setStatus(String(error&&error.message||error)));
 };
 (0,__REACT__.useEffect)(()=>{
   if(!bridge)return;
   let active=true,hasOpened=false,frame=0;
   const sync=()=>{
     if(!active||!viewport.current)return;
     const r=viewport.current.getBoundingClientRect();
     const b={x:Math.round(r.left),y:Math.round(r.top),width:Math.round(r.width),height:Math.round(r.height)};
     if(b.width<120||b.height<120)return;
     if(!hasOpened){hasOpened=true;Promise.resolve(bridge.show(b)).catch(error=>{if(active)setStatus(String(error.message||error))});}
     else Promise.resolve(bridge.bounds(b)).catch(()=>{});
   };
   const queue=()=>{if(frame)cancelAnimationFrame(frame);frame=requestAnimationFrame(sync)};
   const stop=bridge.onState(data=>{
     if(!active||!data)return;
     if(data.url){setUrl(data.url);setEntry(data.url)}
     if(typeof data.canGoBack===`boolean`)setBack(data.canGoBack);
     if(typeof data.canGoForward===`boolean`)setForward(data.canGoForward);
     if(data.phase===`ready`){setBusy(false);setStatus(``)}
     else if(data.phase===`loading`){setBusy(true);setStatus(`Loading Roblox…`)}
     else if(data.message){setBusy(false);setStatus(data.message)}
   });
   const observer=typeof ResizeObserver!==`undefined`?new ResizeObserver(queue):null;
   if(observer&&viewport.current)observer.observe(viewport.current);
   window.addEventListener(`resize`,queue);
   document.addEventListener(`scroll`,queue,true);
   queue();
   return ()=>{
     active=false;
     if(frame)cancelAnimationFrame(frame);
     observer&&observer.disconnect();
     window.removeEventListener(`resize`,queue);
     document.removeEventListener(`scroll`,queue,true);
     stop&&stop();
     Promise.resolve(bridge.hide()).catch(()=>{});
   };
 },[]);
 const go=url=>{
  let valid=false;
  try {const u=new URL(url.startsWith(`http`)?url:`https://www.roblox.com/`+url.replace(/^\/+/,``));valid=u.protocol===`https:`&&(u.hostname===`roblox.com`||u.hostname.endsWith(`.roblox.com`));if(valid){setEntry(u.href);run(`navigate`,u.href)}}catch{}
  if(!valid)setStatus(`Only official Roblox pages can open in this tab.`);
 };
 const btn=(label,action,disabled=false,title=label)=>(0,__JSX__.jsx)(`button`,{type:`button`,title,disabled:disabled||!bridge,onClick:action,children:label});
 return (0,__JSX__.jsxs)(`section`,{className:`lunar-patcher-roblox-browser`,children:[
  (0,__JSX__.jsxs)(`header`,{className:`lpr-browser-toolbar`,children:[
   (0,__JSX__.jsx)(`span`,{className:`lpr-browser-heading`,children:`Roblox`}),
   btn(`←`,()=>run(`back`),!back,`Back`),btn(`→`,()=>run(`forward`),!forward,`Forward`),btn(`↻`,()=>run(`reload`),false,`Reload`),btn(`⌂`,()=>run(`home`),false,`Home`),
   (0,__JSX__.jsx)(`form`,{className:`lpr-browser-address`,onSubmit:e=>{e.preventDefault();go(entry)},children:(0,__JSX__.jsx)(`input`,{type:`text`,value:entry,onChange:e=>setEntry(e.target.value),spellCheck:false,autoComplete:`off`,"aria-label":`Roblox page address`})}),
   btn(`↗`,()=>run(`external`,url),false,`Open this page in your browser`)
  ]}),
  (0,__JSX__.jsx)(`div`,{ref:viewport,className:`lpr-browser-viewport`,"aria-label":`Official Roblox website`}),
  (0,__JSX__.jsxs)(`div`,{className:`lpr-browser-footer`,children:[
    (0,__JSX__.jsx)(`span`,{children:status||(busy?`Loading Roblox…`:`Official roblox.com · Login stays with Roblox`)}),
    btn(`Open in browser`,()=>run(`external`,url),false,`Open Roblox website externally`)
  ]})
 ]});
},
