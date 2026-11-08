'use strict';
const assert=require('node:assert/strict');
const fs=require('node:fs');
const vm=require('node:vm');
const code=fs.readFileSync('engine/assets/roblox_widget.js','utf8')
 .split('\n').slice(1).join('\n')
 .replaceAll('__JSX__','V').replaceAll('__REACT__','B').replaceAll('__INITIATOR__','tp.sidebar')
 .trim().replace(/,\s*$/,'');
const opened=[], stored=new Map();
const hooks=[]; let hookIndex=0;
const context={
 V:{jsx:(type,props)=>({type,props}),jsxs:(type,props)=>({type,props})},
 B:{useState:init=>{const i=hookIndex++;if(!(i in hooks)) hooks[i]=typeof init==='function'?init():init;return[hooks[i],next=>hooks[i]=typeof next==='function'?next(hooks[i]):next]}},
 tp:{sidebar:'sidebar'},
 window:{lunar:{misc:{openExternalLink:async opts=>{opened.push(opts);return true}}},confirm:()=>true},
 localStorage:{getItem:k=>stored.get(k)||null,setItem:(k,v)=>stored.set(k,v)},
 URL,console,Promise,JSON
};
vm.createContext(context);
vm.runInContext('var '+code+';',context);
const render=()=>{hookIndex=0;return context.lunarPatcherRobloxPage()};
function walk(node, fn){if(!node||typeof node!=='object')return;fn(node);const children=node.props?.children;for(const x of Array.isArray(children)?children:[children])walk(x,fn)}
function findText(node,text,type='button'){
 let found;walk(node,x=>{if(x.type===type&&x.props?.children===text)found=x});return found;
}
function findInput(node,placeholder){let found;walk(node,x=>{if(x.type==='input'&&x.props?.placeholder===placeholder)found=x});return found}
function allNodes(node,type){const res=[];walk(node,x=>{if(x.type===type)res.push(x)});return res;}
const enter=(node,text)=>{assert.ok(node,`Missing input`);node.props.onChange({target:{value:text}})};
const tap=(node)=>{assert.ok(node,`Missing control`);node.props.onClick()};
const tick=async()=>{await Promise.resolve();await Promise.resolve()};
(async()=>{
 assert.equal(typeof context.lunarPatcherRobloxPage,'function');
 let page=render();
 assert.ok(findText(page,'Games'));
 assert.ok(findText(page,'Accounts'));
 assert.ok(findText(page,'Settings'));
 // A curated game launches using only the well-formed official protocol URL.
 const plays=allNodes(page,'button').filter(x=>x.props?.children==='Play');
 assert.ok(plays.length>=3);
 tap(plays[1]);await tick();
 assert.equal(opened.at(-1).url,'roblox://experiences/start?placeId=4924922222');
 // Favorites store IDs and labels; can be renamed and removed.
 page=render();enter(findInput(page,'Game URL, place ID or private server invite'),'https://www.roblox.com/games/2753915549/Blox-Fruits');
 page=render();tap(findText(page,'Save'));
 assert.equal(JSON.parse(stored.get('lunarpatcher.roblox.favorites.v1'))[0].id,'2753915549');
 page=render();const edits=allNodes(page,'button').filter(x=>x.props?.children==='Edit');assert.equal(edits.length,1);tap(edits[0]);
 page=render();enter(findInput(page,'missing')||allNodes(page,'input').find(x=>x.props?.['aria-label']==='New game name'),'My game');page=render();
 const saveButtons=allNodes(page,'button').filter(x=>x.props?.children==='Save'&&x.props?.className?.includes('textbtn'));assert.equal(saveButtons.length,1);tap(saveButtons[0]);
 assert.equal(JSON.parse(stored.get('lunarpatcher.roblox.favorites.v1'))[0].name,'My game');
 // Accounts are saved without credentials and switching opens only the official website.
 page=render();tap(findText(page,'Accounts'));page=render();
 enter(findInput(page,'Roblox username'),'Sample_User');page=render();enter(findInput(page,'Optional label'),'Main');page=render();tap(findText(page,'Add profile'));
 const accounts=JSON.parse(stored.get('lunarpatcher.roblox.accounts.v1'));
 assert.equal(accounts.length,1);assert.equal(accounts[0].username,'Sample_User');
 assert.ok(!JSON.stringify(accounts).match(/cookie|password|token/i));
 page=render();tap(findText(page,'Switch on Roblox'));await tick();assert.equal(opened.at(-1).url,'https://www.roblox.com/home');
 page=render();tap(findText(page,'Prefer'));assert.equal(JSON.parse(stored.get('lunarpatcher.roblox.preferred.v1')),'Sample_User');
 page=render();tap(findText(page,'Remove'));assert.equal(JSON.parse(stored.get('lunarpatcher.roblox.accounts.v1')).length,0);
 // Settings keep launch mode, fallback, and history behavior independent.
 page=render();tap(findText(page,'Settings'));page=render();
 const selects=allNodes(page,'select');assert.equal(selects.length,1);selects[0].props.onChange({target:{value:'website'}});
 assert.equal(JSON.parse(stored.get('lunarpatcher.roblox.settings.v1')).launchMode,'website');
 page=render();tap(findText(page,'Games'));page=render();tap(allNodes(page,'button').find(x=>x.props?.children==='Play'&&x.props?.className?.includes('play')));await tick();
 // blank target causes no new website launch
 const countBefore=opened.length;
 page=render();enter(findInput(page,'Game URL, place ID or private server invite'),'https://evil.example/games/123');page=render();tap(allNodes(page,'button').find(x=>x.props?.children==='Play'));await tick();assert.equal(opened.length,countBefore);
 // A private server invitation is only opened at Roblox's website.
 page=render();enter(findInput(page,'Game URL, place ID or private server invite'),'https://www.roblox.com/share?code=AbcDef0123&type=Server');page=render();tap(allNodes(page,'button').find(x=>x.props?.children==='Play'));await tick();
 assert.equal(opened.at(-1).url,'https://www.roblox.com/share?code=AbcDef0123&type=Server');
 console.log('PASS Roblox page: launch, favorites rename, account bookmark, official switcher, preferences, link validation');
})().catch(err=>{console.error(err);process.exit(1)});
