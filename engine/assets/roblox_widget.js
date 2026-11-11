// Roblox.com lives directly in Lunar's content area: no duplicate launcher UI.
// Credentials and site navigation stay in the dedicated, sandboxed BrowserView.
lunarPatcherRobloxIcon=props=>(0,__JSX__.jsx)(`svg`,{viewBox:`0 0 24 24`,width:props?.size||22,height:props?.size||22,fill:`currentColor`,"aria-hidden":!0,children:(0,__JSX__.jsx)(`path`,{fillRule:`evenodd`,d:`M7.56 1.15 22.85 7.56 16.44 22.85 1.15 16.44 7.56 1.15ZM10.23 8.45 8.45 13.77 13.77 15.55 15.55 10.23 10.23 8.45Z`})}),
lunarPatcherRobloxPage=()=>{
 const bridge=window.lunarPatcherRobloxBrowser;
 const viewport=(0,__REACT__.useRef)(null);
 const [error,setError]=(0,__REACT__.useState)(bridge?``:`Roblox browser unavailable. Reapply the patch.`);
 (0,__REACT__.useEffect)(()=>{
  if(!bridge)return;
  let active=true,opened=false,frame=0;
  const sync=()=>{
   if(!active||!viewport.current)return;
   const r=viewport.current.getBoundingClientRect();
   const bounds={x:Math.round(r.left),y:Math.round(r.top),width:Math.round(r.width),height:Math.round(r.height)};
   if(bounds.width<100||bounds.height<100)return;
   const result=opened?bridge.bounds(bounds):bridge.show(bounds);
   opened=true;
   Promise.resolve(result).catch(e=>{if(active)setError(String(e.message||e))});
  };
  const queue=()=>{if(frame)cancelAnimationFrame(frame);frame=requestAnimationFrame(sync)};
  const stop=bridge.onState(data=>{
   if(!active||!data)return;
   if(data.phase===`error` && data.message)setError(data.message);
  });
  const observer=typeof ResizeObserver!==`undefined`?new ResizeObserver(queue):null;
  if(observer&&viewport.current)observer.observe(viewport.current);
  window.addEventListener(`resize`,queue);document.addEventListener(`scroll`,queue,true);queue();
  return ()=>{active=false;if(frame)cancelAnimationFrame(frame);observer&&observer.disconnect();window.removeEventListener(`resize`,queue);document.removeEventListener(`scroll`,queue,true);stop&&stop();Promise.resolve(bridge.hide()).catch(()=>{})};
 },[]);
 return (0,__JSX__.jsxs)(`section`,{className:`lunar-patcher-roblox-browser`,children:[
  (0,__JSX__.jsx)(`div`,{ref:viewport,className:`lpr-browser-viewport`,"aria-label":`Roblox website`}),
  error?(0,__JSX__.jsxs)(`div`,{className:`lpr-browser-error`,role:`alert`,children:[
   (0,__JSX__.jsx)(`span`,{children:error}),
   bridge?(0,__JSX__.jsx)(`button`,{type:`button`,onClick:()=>bridge.external(`https://www.roblox.com/home`),children:`Open Roblox in browser`}):null
  ]}):null
 ]});
},
