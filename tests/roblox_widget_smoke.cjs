'use strict';
const assert=require('node:assert/strict');
const fs=require('node:fs');
const vm=require('node:vm');
const code=fs.readFileSync('engine/assets/roblox_widget.js','utf8')
 .split('\n').slice(1).join('\n')
 .replaceAll('__JSX__','V').replaceAll('__REACT__','B').replaceAll('__INITIATOR__','tp.sidebar')
 .trim().replace(/,\s*$/,'');
const opened=[],stored=new Map();const hooks=[];let hookIndex=0;
const ctx={V:{jsx:(type,props)=>({type,props})},B:{useState:init=>{const i=hookIndex++;if(!(i in hooks))hooks[i]=typeof init==='function'?init():init;return[hooks[i],next=>{hooks[i]=typeof next==='function'?next(hooks[i]):next}]}},tp:{sidebar:'sidebar'},
window:{lunar:{misc:{openExternalLink:async args=>{opened.push(args);return true}}},confirm:()=>true},
localStorage:{getItem:key=>stored.get(key)||null,setItem:(key,v)=>stored.set(key,v)},URL,Promise,JSON,console};
vm.createContext(ctx);vm.runInContext('var '+code+';',ctx);
const VOID=new Set('area base br col embed hr img input link meta param source track wbr'.split(' '));
const render=()=>{hookIndex=0;const tree=ctx.lunarPatcherRobloxPage();walk(tree,el=>{
  if(typeof el.type==='string'&&VOID.has(el.type)){
   assert.equal(Object.prototype.hasOwnProperty.call(el.props||{},'children'),false,
     `React #137: ${el.type} must not receive children`);
  }
});return tree};
function walk(node,fn){if(node==null||typeof node!=='object')return;fn(node);const c=node.props?.children;for(const x of Array.isArray(c)?c:[c])walk(x,fn)}
function all(tree,type){const a=[];walk(tree,n=>{if(n.type===type)a.push(n)});return a}
function button(tree,text){return all(tree,'button').find(n=>n.props?.children===text||(Array.isArray(n.props?.children)&&n.props.children.at(-1)===text))}
function input(tree,placeholder){return all(tree,'input').find(n=>n.props?.placeholder===placeholder)}
function tap(x){assert.ok(x,'Missing button');x.props.onClick()}
function type(x,value){assert.ok(x,'Missing input');x.props.onChange({target:{value}})}
const tick=async()=>{await Promise.resolve();await Promise.resolve()};
(async()=>{
 let tree=render();
 for(const nav of ['Home','Discover','Library','Friends','Avatar','Marketplace','Accounts','Settings'])
  assert.ok(button(tree,nav),`Missing ${nav} navigation`);
 // Home -> details -> launch -> saved
 const details=button(tree,'Game details');tap(details);tree=render();
 assert.ok(button(tree,'Save to library'));tap(button(tree,'Save to library'));
 assert.ok(button(tree,'▶ Play'));tap(button(tree,'▶ Play'));await tick();
 assert.equal(opened.at(-1).url,'roblox://experiences/start?placeId=2753915549');
 assert.equal(JSON.parse(stored.get('lunarpatcher.roblox.favorites.v1')).length,1);
 tap(button(tree,'← Back'));tree=render();
 tap(button(tree,'Discover'));tree=render();
 assert.ok(all(tree,'article').length>=7);
 tap(button(tree,'Horror'));tree=render();assert.equal(all(tree,'article').length,1);
 tap(button(tree,'All'));tree=render();assert.ok(all(tree,'article').length>=7);
 // Search filters catalog, and search input has no React children.
 type(input(tree,'Search experiences'),'DOORS');tree=render();assert.equal(all(tree,'article').length,1);
 type(input(tree,'Search experiences'),'');tree=render();
 // Library URL, saved shortcuts, invalid URL rejection.
 tap(button(tree,'Library'));tree=render();
 type(input(tree,'Paste a Roblox game link or place ID'),'https://www.roblox.com/games/4924922222/Brookhaven-RP');
 tree=render();tap(button(tree,'Save'));
 assert.equal(JSON.parse(stored.get('lunarpatcher.roblox.favorites.v1')).length,2);
 tree=render();type(input(tree,'Paste a Roblox game link or place ID'),'https://evil.example/games/999');tree=render();const before=opened.length;tap(button(tree,'Play'));await tick();assert.equal(opened.length,before);
 type(input(tree,'Paste a Roblox game link or place ID'),'https://www.roblox.com/share?code=AbcDef1234&type=Server');tree=render();tap(button(tree,'Play'));await tick();assert.equal(opened.at(-1).url,'https://www.roblox.com/share?code=AbcDef1234&type=Server');
 // Official-only friends/avatar/marketplace paths.
 for(const [nav,label,url] of [['Friends','Open friends ↗','https://www.roblox.com/users/friends'],['Avatar','Open Avatar Editor ↗','https://www.roblox.com/my/avatar'],['Marketplace','Browse Marketplace ↗','https://www.roblox.com/catalog']]){
  tree=render();tap(button(tree,nav));tree=render();tap(button(tree,label));await tick();assert.equal(opened.at(-1).url,url);
 }
 // Account profiles are bookmarks, not a session mechanism.
 tree=render();tap(button(tree,'Accounts'));tree=render();
 type(input(tree,'Roblox username'),'Sample_User');tree=render();type(input(tree,'Optional label'),'Main');tree=render();tap(button(tree,'Add'));
 assert.equal(JSON.parse(stored.get('lunarpatcher.roblox.accounts.v1')).length,1);
 tree=render();tap(button(tree,'Switch'));await tick();assert.equal(opened.at(-1).url,'https://www.roblox.com/home');
 assert.ok(!JSON.stringify([...stored]).match(/authcookie|password|sessiontoken/i));
 // Settings changes.
 tree=render();tap(button(tree,'Settings'));tree=render();
 const select=all(tree,'select')[0];assert.ok(select);select.props.onChange({target:{value:'website'}});
 assert.equal(JSON.parse(stored.get('lunarpatcher.roblox.settings.v1')).launchMode,'website');
 console.log('PASS Roblox hub: real void-element validation, 8 pages, details, launch, saving, link safety, account bookmarks, settings');
})().catch(e=>{console.error(e);process.exit(1)});
