'use strict';
const assert=require('node:assert/strict');
const fs=require('node:fs');
const vm=require('node:vm');
const path=require('node:path');
const src=fs.readFileSync(path.join(__dirname,'../engine/assets/roblox_widget.js'),'utf8');
const calls=[],listeners=[],effects=[];
const bridge={
 show:bounds=>{calls.push(['show',bounds]);return Promise.resolve(true)},
 bounds:bounds=>{calls.push(['bounds',bounds]);return Promise.resolve(true)},
 hide:()=>{calls.push(['hide']);return Promise.resolve(true)},
 navigate:url=>{calls.push(['navigate',url]);return Promise.resolve(true)},
 external:url=>{calls.push(['external',url]);return Promise.resolve(true)},
 reload:()=>{calls.push(['reload']);return Promise.resolve(true)},
 home:()=>{calls.push(['home']);return Promise.resolve(true)},
 back:()=>Promise.resolve(true),forward:()=>Promise.resolve(true),
 onState:fn=>{listeners.push(fn);return()=>calls.push(['unsubscribe'])}
};
const state=[],refs=[];let hook=0;
const JSX={jsx:(type,props)=>({type,props:props||{}}),jsxs:(type,props)=>({type,props:props||{}})};
const React={
 useState:initial=>{const index=hook++;if(!(index in state))state[index]=typeof initial==='function'?initial():initial;return [state[index],next=>{state[index]=next}]},
 useRef:value=>{const ref={current:value};refs.push(ref);return ref},
 useEffect:fn=>effects.push(fn)
};
const frames=[];
const ctx={window:{lunarPatcherRobloxBrowser:bridge,addEventListener(){},removeEventListener(){}},document:{addEventListener(){},removeEventListener(){}},ResizeObserver:class{observe(){} disconnect(){}},requestAnimationFrame:fn=>{frames.push(fn);return frames.length},cancelAnimationFrame:()=>{},URL,Promise,console,JSX,React};
vm.runInNewContext('var '+src.replaceAll('__JSX__','JSX').replaceAll('__REACT__','React')+'dW=()=>{};\nthis.page=lunarPatcherRobloxPage;',ctx);
const walk=(tree,list=[])=>{if(!tree||typeof tree!=='object')return list;list.push(tree);const c=tree.props?.children;if(Array.isArray(c))c.forEach(x=>walk(x,list));else walk(c,list);return list};
const tree=ctx.page();
const elements=walk(tree);
for(const e of elements){if(e.type==='input'&&('children' in e.props))throw Error('React void input received children');}
assert.equal(elements.filter(e=>e.type==='input').length,1);
assert.equal(elements.filter(e=>e.type==='section').length,1);
const viewport=elements.find(e=>e.props?.className==='lpr-browser-viewport');assert(viewport);
viewport.props.ref.current={getBoundingClientRect:()=>({left:110,top:125,width:700,height:520})};
const cleanup=effects[0]();assert.equal(typeof cleanup,'function');frames.shift()?.();
(async()=>{
 await new Promise(resolve=>setImmediate(resolve));
 assert.equal(calls[0][0],'show');assert.equal(calls[0][1].width,700);
 const inputs=elements.filter(e=>e.type==='input'); assert.equal(inputs[0].props.value,'https://www.roblox.com/home');
 const buttons=elements.filter(e=>e.type==='button');
 buttons.find(e=>e.props.title==='Home').props.onClick();
 buttons.find(e=>e.props.title==='Reload').props.onClick();
 await new Promise(resolve=>setImmediate(resolve));assert(calls.some(x=>x[0]==='home'));assert(calls.some(x=>x[0]==='reload'));
 cleanup();await new Promise(resolve=>setImmediate(resolve));
 assert.equal(calls.at(-1)[0],'hide');
 console.log('PASS: in-Lunar Roblox website, native toolbar, safe React void elements, browser lifecycle');
})().catch(e=>{console.error(e);process.exitCode=1});
