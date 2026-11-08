'use strict';
const assert=require('node:assert/strict');
const {parseExperience,launchRoblox,cleanPrefs,parseDesktopLink}=require('../electron/roblox_launcher.cjs');
async function main(){
  assert.deepEqual(parseDesktopLink('lunarpatcher://launch?placeId=2753915549'),{target:'2753915549',mode:'direct'});
  assert.equal(parseDesktopLink('lunarpatcher://launch?placeId=0'),null);
  assert.equal(parseDesktopLink('lunarpatcher://evil?placeId=2753915549'),null);
  assert.deepEqual(parseExperience('2753915549'),{type:'place',placeId:'2753915549'});
  assert.equal(parseExperience('https://www.roblox.com/games/2753915549/Blox-Fruits?x=1').placeId,'2753915549');
  for(const url of ['javascript:alert(1)','https://evil.com/games/123','roblox://evil.com/','file:///tmp/secret','0','https://roblox.com/games/0'])assert.throws(()=>parseExperience(url));
  const opened=[];const shell={openExternal:async url=>{opened.push(url);}};
  const diag=async()=>({platform:'win32',protocol:true,playerProtocol:true});
  const a=await launchRoblox({target:'2753915549',mode:'direct'},{shell,platform:'win32',diagnoseFn:diag});
  assert.equal(a.method,'protocol');assert.equal(opened.pop(),'roblox://experiences/start?placeId=2753915549');
  const b=await launchRoblox({target:'2753915549',mode:'browser'},{shell,platform:'win32',diagnoseFn:diag});
  assert.equal(b.method,'browser');assert.equal(opened.pop(),'https://www.roblox.com/games/2753915549');
  const c=await launchRoblox({target:'2753915549'},{shell,platform:'win32',diagnoseFn:async()=>({protocol:false})});
  assert.equal(c.method,'browser-fallback');assert.equal(opened.pop(),'https://www.roblox.com/games/2753915549');
  const d=await launchRoblox({app:true},{shell,platform:'win32',diagnoseFn:diag});assert.equal(d.method,'protocol');assert.equal(opened.pop(),'roblox-player:1+launchmode:app');
  const errorShell={openExternal:async url=>{if(url.startsWith('roblox:'))throw new Error('denied');opened.push(url);}};
  const e=await launchRoblox({target:'1818'},{shell:errorShell,platform:'win32',diagnoseFn:diag});assert.equal(e.method,'browser-fallback');
  const prefs=cleanPrefs({fps:'9999',layout:'weird',favorites:[{placeId:'1',name:'Hi'},{placeId:'javascript',name:'no'}]});assert.equal(prefs.fps,'default');assert.equal(prefs.layout,'comfortable');assert.equal(prefs.favorites.length,1);
  console.log('PASS: Roblox direct protocol, web, fallback, app, validation, preferences');
}
main().catch(err=>{console.error(err);process.exit(1)});
