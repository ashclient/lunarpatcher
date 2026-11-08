'use strict';
// Roblox launch routing. No credentials, cookies, Roblox binaries or FastFlags modified.
const {execFile} = require('node:child_process');
const fs = require('node:fs');
const path = require('node:path');
const os = require('node:os');
const PLACE=/^[1-9]\d{0,17}$/;
const FPS=new Set(['default','60','120','144','240']);
const LAYOUT=new Set(['comfortable','compact','spacious']);
const MODES=new Set(['direct','browser']);
function parseExperience(raw) {
  const s=String(raw??'').trim();
  if(PLACE.test(s)) return {type:'place',placeId:s};
  if(s.length>1500) throw new Error('Game link is too long');
  let url;
  try { url=new URL(s); } catch { throw new Error('Paste a Roblox experience URL or a numeric place ID'); }
  if(url.protocol==='roblox:' && url.hostname.toLowerCase()==='experiences' && url.pathname==='/start') {
    const id=url.searchParams.get('placeId');
    if(PLACE.test(id)) return {type:'place',placeId:id};
  }
  if(url.protocol==='https:' && ['www.roblox.com','roblox.com'].includes(url.hostname.toLowerCase())) {
    const match=url.pathname.match(/^\/games\/([1-9]\d{0,17})(?:\/(?:[^?#]*))?\/?$/i);
    if(match)return {type:'place',placeId:match[1]};
    if(url.pathname==='/share' && url.searchParams.get('type')==='Server') {
      const code=url.searchParams.get('code');
      if(code && /^[a-zA-Z0-9_+\/-]{3,128}$/.test(code)) return {type:'server',url:`https://www.roblox.com/share?code=${encodeURIComponent(code)}&type=Server`};
    }
  }
  throw new Error('Only official Roblox game links, private-server invitations, or place IDs are supported');
}
function gameUrl(placeId) { return `https://www.roblox.com/games/${placeId}`; }
function protocolUrl(placeId) { return `roblox://experiences/start?placeId=${placeId}`; }
function cleanPrefs(input) {
  const v=input && typeof input==='object' && !Array.isArray(input)?input:{};
  const favorites=(Array.isArray(v.favorites)?v.favorites:[]).filter(x=>x&&PLACE.test(String(x.placeId))).slice(0,40).map(x=>({placeId:String(x.placeId),name:String(x.name||`Experience ${x.placeId}`).slice(0,80)}));
  const recent=(Array.isArray(v.recent)?v.recent:[]).filter(x=>x&&PLACE.test(String(x.placeId))).slice(0,16).map(x=>({placeId:String(x.placeId),name:String(x.name||`Experience ${x.placeId}`).slice(0,80),time:Math.max(0,Math.trunc(Number(x.time)||0))}));
  return {launchMode:MODES.has(v.launchMode)?v.launchMode:'direct',fps:FPS.has(v.fps)?v.fps:'default',layout:LAYOUT.has(v.layout)?v.layout:'comfortable',
    accent:/^#[0-9a-f]{6}$/i.test(v.accent||'')?v.accent:'#E5A1BB',
    showDetails:v.showDetails!==false,confirmLaunch:v.confirmLaunch===true,keepHistory:v.keepHistory!==false,favorites,recent};
}
function registryCheck(key) {return new Promise(resolve=>{
  execFile('reg.exe',['query',key],{windowsHide:true,timeout:4000},err=>resolve(!err));
});}
async function diagnose(platform=process.platform) {
  if(platform!=='win32')return {platform,protocol:false,playerProtocol:false,playerExe:null,reason:'Roblox launch integration requires Windows'};
  const base=process.env.LOCALAPPDATA || path.join(os.homedir(),'AppData','Local');
  const protocols=await Promise.all(['roblox','roblox-player'].map(p=>registryCheck(`HKCR\\${p}\\shell\\open\\command`)));
  const roots=[path.join(base,'Roblox','Player','RobloxPlayerBeta.exe')];
  try {const versions=path.join(base,'Roblox','Versions');for(const dir of fs.readdirSync(versions).slice(0,100))roots.push(path.join(versions,dir,'RobloxPlayerBeta.exe'));} catch{}
  const found=roots.find(f=>fs.existsSync(f))||null;
  return {platform,protocol:protocols[0],playerProtocol:protocols[1],playerExe:found,reason:protocols[0]?'Roblox deeplinks registered':found?'Roblox files found but deeplink registration is missing':'Roblox Player not detected; install it from Roblox.com'};
}
async function launchRoblox(input,{shell,platform=process.platform,diagnoseFn=()=>diagnose(platform)}={}) {
  if(!shell || typeof shell.openExternal!=='function')throw new Error('Desktop shell is unavailable');
  const target=input?.app?{type:'app'}:parseExperience(input?.target);
  const mode=MODES.has(input?.mode)?input.mode:'direct';
  if(target.type==='server') { await shell.openExternal(target.url); return {handedOff:true,method:'browser',url:target.url,message:'Opened private-server invitation in your browser'}; }
  const website=target.type==='app'?'https://www.roblox.com/home':gameUrl(target.placeId);
  if(mode==='browser'){await shell.openExternal(website);return {handedOff:true,method:'browser',url:website,message:'Opened Roblox in your browser'};}
  const status=await diagnoseFn();
  if(platform!=='win32'||(target.type==='app'?!status.playerProtocol:!status.protocol)) {
    await shell.openExternal(website);
    return {handedOff:true,method:'browser-fallback',url:website,message:'Roblox protocol was not registered; opened Roblox website instead',diagnostics:status};
  }
  const url=target.type==='app'?'roblox-player:1+launchmode:app':protocolUrl(target.placeId);
  try {await shell.openExternal(url,{activate:true}); return {handedOff:true,method:'protocol',url,message:'Windows accepted the Roblox launch link; the Roblox Player will determine whether a session can start'};}
  catch(e){await shell.openExternal(website);return {handedOff:true,method:'browser-fallback',url:website,message:'Roblox protocol failed; opened its official website instead',error:String(e?.message||e).slice(0,200)};}
}
module.exports={parseExperience,gameUrl,protocolUrl,cleanPrefs,diagnose,launchRoblox};

// Only a place ID may be passed through LunarPatcher's optional URI handoff.
module.exports.parseDesktopLink=function(raw){
  let url;try{url=new URL(String(raw||''));}catch{return null}
  if(url.protocol!=='lunarpatcher:' || url.hostname!=='launch'||!['','/'].includes(url.pathname))return null;
  const id=url.searchParams.get('placeId');
  if(!PLACE.test(id))return null;
  return {target:id,mode:'direct'};
};
