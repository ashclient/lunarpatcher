// LunarPatcher Roblox workspace. Independently implemented from published feature descriptions.
// Roblox browsing, authentication and game joining remain handled by the official site/client.
lunarPatcherRobloxIcon=props=>(0,__JSX__.jsx)(`svg`,{viewBox:`0 0 24 24`,width:props?.size||22,height:props?.size||22,fill:`currentColor`,"aria-hidden":!0,children:(0,__JSX__.jsx)(`path`,{fillRule:`evenodd`,d:`M7.56 1.15 22.85 7.56 16.44 22.85 1.15 16.44 7.56 1.15ZM10.23 8.45 8.45 13.77 13.77 15.55 15.55 10.23 10.23 8.45Z`})}),
lunarPatcherRobloxPage=()=>{
 const bridge=window.lunarPatcherRobloxBrowser;
 const h=(tag,props,children)=>{
  const p={...(props||{})};
  if(children!==undefined)p.children=children;
  return Array.isArray(children)?(0,__JSX__.jsxs)(tag,p):(0,__JSX__.jsx)(tag,p);
 };
 const categories=[
  {title:`YOUR SPACE`,items:[[`Home`,`/home`,`⌂`],[`Discover`,`/discover`,`◇`],[`Friends`,`/users/friends#!/friends`,`♧`],[`Messages`,`/messages`,`▤`]]},
  {title:`ROBLOX`,items:[[`Avatar`,`/my/avatar`,`◉`],[`Marketplace`,`/catalog`,`▦`],[`Communities`,`/communities`,`▧`],[`Trades`,`/trades`,`⇄`]]},
  {title:`ACCOUNT`,items:[[`Settings`,`/my/account#!/info`,`⚙`]]}
 ];
 const routeUrl=path=>`https://www.roblox.com`+path;
 const [current,setCurrent]=(0,__REACT__.useState)(`Home`);
 const [url,setUrl]=(0,__REACT__.useState)(routeUrl(`/home`));
 const [entry,setEntry]=(0,__REACT__.useState)(routeUrl(`/home`));
 const [search,setSearch]=(0,__REACT__.useState)(``);
 const [status,setStatus]=(0,__REACT__.useState)(bridge?`Connecting to Roblox…`:`Roblox integration unavailable. Reapply the patch.`);
 const [back,setBack]=(0,__REACT__.useState)(false);
 const [forward,setForward]=(0,__REACT__.useState)(false);
 const [busy,setBusy]=(0,__REACT__.useState)(true);
 const [saved,setSaved]=(0,__REACT__.useState)(()=>{
   try {const a=JSON.parse(window.localStorage?.getItem(`lunarpatcher.roblox.saved.v2`)||`[]`);
    return Array.isArray(a)?a.filter(x=>x&&typeof x.url===`string`&&x.url.startsWith(`https://www.roblox.com/`)).slice(0,12):[];
   } catch {return []}
 });
 const viewport=(0,__REACT__.useRef)(null);
 const run=(name,...args)=>{
  if(!bridge||typeof bridge[name]!==`function`)return Promise.resolve(false);
  return Promise.resolve(bridge[name](...args)).catch(error=>{setStatus(String(error?.message||error));return false});
 };
 const validUrl=value=>{
  try {const u=new URL(value.startsWith(`http`)?value:`https://www.roblox.com/`+value.replace(/^\/+/,``));
   return u.protocol===`https:`&&(u.hostname===`roblox.com`||u.hostname.endsWith(`.roblox.com`))&&!u.username&&!u.password?u.href:null;
  } catch {return null}
 };
 const navigate=value=>{
  const dest=validUrl(value);
  if(!dest){setStatus(`Only official Roblox pages are allowed.`);return}
  setEntry(dest);setUrl(dest);setBusy(true);run(`navigate`,dest);
 };
 const goTo=(label,path)=>{setCurrent(label);navigate(routeUrl(path))};
 const bookmark=()=>{
  const safe=validUrl(url);if(!safe)return;
  const updated=[{label:current===`Home`?`Saved page`:current,url:safe},...saved.filter(x=>x.url!==safe)].slice(0,12);
  setSaved(updated);
  try {window.localStorage?.setItem(`lunarpatcher.roblox.saved.v2`,JSON.stringify(updated))}catch{}
 };
 (0,__REACT__.useEffect)(()=>{
  if(!bridge)return;
  let active=true,opened=false,raf=0;
  const sync=()=>{
   if(!active||!viewport.current)return;
   const r=viewport.current.getBoundingClientRect();
   const bounds={x:Math.round(r.left),y:Math.round(r.top),width:Math.round(r.width),height:Math.round(r.height)};
   if(bounds.width<120||bounds.height<120)return;
   if(!opened){opened=true;run(`show`,bounds)}else run(`bounds`,bounds);
  };
  const queue=()=>{if(raf)cancelAnimationFrame(raf);raf=requestAnimationFrame(sync)};
  const unsubscribe=bridge.onState(data=>{
   if(!active||!data)return;
   if(data.url){setUrl(data.url);setEntry(data.url)}
   if(typeof data.canGoBack===`boolean`)setBack(data.canGoBack);
   if(typeof data.canGoForward===`boolean`)setForward(data.canGoForward);
   if(data.phase===`ready`){setBusy(false);setStatus(``)}
   else if(data.phase===`loading`){setBusy(true);setStatus(`Loading…`)}
   else if(data.message){setBusy(false);setStatus(data.message)}
  });
  const observer=typeof ResizeObserver!==`undefined`?new ResizeObserver(queue):null;
  if(observer&&viewport.current)observer.observe(viewport.current);
  window.addEventListener(`resize`,queue);document.addEventListener(`scroll`,queue,true);queue();
  return ()=>{active=false;if(raf)cancelAnimationFrame(raf);observer&&observer.disconnect();
   window.removeEventListener(`resize`,queue);document.removeEventListener(`scroll`,queue,true);
   unsubscribe&&unsubscribe();Promise.resolve(bridge.hide()).catch(()=>{});
  };
 },[]);
 const button=(label,handler,title=label,disabled=false,className=`lpr-tool`)=>h(`button`,{type:`button`,className,title,disabled:disabled||!bridge,onClick:handler},label);
 const sidebar=categories.map(group=>h(`div`,{className:`lpr-navigation-group`,key:group.title},[
  h(`p`,{className:`lpr-group-heading`},group.title),
  ...group.items.map(([label,path,symbol])=>h(`button`,{type:`button`,key:label,className:`lpr-nav-link`+(current===label?` is-active`:``),onClick:()=>goTo(label,path)},[
   h(`span`,{className:`lpr-nav-symbol`,"aria-hidden":true},symbol),h(`span`,{},label)
  ]))
 ]));
 return h(`section`,{className:`lunar-patcher-roblox-browser lpr-workspace`},[
  h(`aside`,{className:`lpr-navigation`},[
   h(`div`,{className:`lpr-brand`},[lunarPatcherRobloxIcon({size:28}),h(`span`,{},`Roblox`)]),
   h(`div`,{className:`lpr-nav-scroll`},[
    ...sidebar,
    h(`div`,{className:`lpr-navigation-group lpr-saved`},[
     h(`p`,{className:`lpr-group-heading`},`SAVED PAGES`),
     ...(saved.length?saved.map((item,i)=>h(`button`,{type:`button`,key:item.url,className:`lpr-nav-link lpr-saved-link`,title:item.url,onClick:()=>{setCurrent(item.label);navigate(item.url)}},[
      h(`span`,{className:`lpr-nav-symbol`,"aria-hidden":true},`☆`),h(`span`,{},item.label),
      h(`span`,{className:`lpr-remove`,title:`Remove saved page`,onClick:e=>{e.stopPropagation();const next=saved.filter((_,j)=>j!==i);setSaved(next);try{window.localStorage?.setItem(`lunarpatcher.roblox.saved.v2`,JSON.stringify(next))}catch{}}},`×`)
     ])): [h(`p`,{className:`lpr-empty-saved`},`Pin pages you visit for quick access.`)])
    ])
   ]),
   h(`footer`,{className:`lpr-navigation-footer`},`Roblox runs through its official website and player`)
  ]),
  h(`main`,{className:`lpr-main`},[
   h(`header`,{className:`lpr-workspace-bar`},[
    h(`div`,{className:`lpr-heading`},[h(`span`,{className:`lpr-page-title`},current),h(`span`,{className:`lpr-page-subtitle`},`Roblox`)]),
    h(`form`,{className:`lpr-search`,onSubmit:e=>{e.preventDefault();const q=search.trim();if(q){setCurrent(`Discover`);navigate(`https://www.roblox.com/discover?Keyword=`+encodeURIComponent(q))}}},[
     h(`span`,{className:`lpr-search-icon`,"aria-hidden":true},`⌕`),
     h(`input`,{type:`search`,placeholder:`Search experiences…`,value:search,onChange:e=>setSearch(e.target.value),autoComplete:`off`,"aria-label":`Search Roblox experiences`}),
     h(`button`,{type:`submit`,disabled:!search.trim()},`Search`)
    ]),
    button(`↗`,()=>run(`external`,url),`Open Roblox page in external browser`)
   ]),
   h(`div`,{className:`lpr-browser-toolbar`},[
    button(`←`,()=>run(`back`),`Back`,!back),
    button(`→`,()=>run(`forward`),`Forward`,!forward),
    button(`↻`,()=>run(`reload`),`Reload`),
    button(`⌂`,()=>goTo(`Home`,`/home`),`Home`),
    h(`form`,{className:`lpr-browser-address`,onSubmit:e=>{e.preventDefault();navigate(entry)}},
     h(`input`,{type:`text`,value:entry,onChange:e=>setEntry(e.target.value),spellCheck:false,autoComplete:`off`,"aria-label":`Roblox page address`})
    ),
    button(`☆`,bookmark,`Save this Roblox page`)
   ]),
   h(`div`,{ref:viewport,className:`lpr-browser-viewport`,"aria-label":`Official Roblox website`}),
   h(`footer`,{className:`lpr-browser-footer`},[
    h(`span`,{className:`lpr-browser-status`,role:`status`},status||(busy?`Loading…`:`Official roblox.com · Game launches use the installed Roblox Player`)),
    button(`Open in browser`,()=>run(`external`,url),`Open in external browser`,false,`lpr-external`)
   ])
  ])
 ]);
},
