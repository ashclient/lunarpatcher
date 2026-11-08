'use strict';
// Exercise the real Electron main-process update wiring with mocked Electron APIs.
const assert=require('node:assert/strict');
const fs=require('node:fs');
const path=require('node:path');
const vm=require('node:vm');
const {EventEmitter}=require('node:events');
const source=fs.readFileSync(path.join(__dirname,'../electron/main.cjs'),'utf8');
const handlers=new Map();
const timers=[];
const updateCalls=[];
const sent=[];
class MockWindow extends EventEmitter {
  constructor(options){super();this.options=options;this.destroyed=false;this.webContents={send:(...p)=>sent.push(p),setWindowOpenHandler:()=>{},on:()=>{}};}
  loadFile(){} isDestroyed(){return false;} show(){}
}
const updater=new EventEmitter();
updater.checkForUpdates=async()=>{updateCalls.push('check');updater.emit('checking-for-update');updater.emit('update-available',{version:'1.3.1'});updater.emit('download-progress',{percent:63});updater.emit('update-downloaded',{version:'1.3.1'});return {version:'1.3.1'}};
updater.quitAndInstall=(...args)=>updateCalls.push(['install',...args]);
const app=new EventEmitter();
app.isPackaged=true;
app.whenReady=()=>Promise.resolve();
app.getVersion=()=> '1.3.0';
app.getPath=() => '/tmp/lunarpatcher-test';
app.requestSingleInstanceLock=()=>true;
app.setAsDefaultProtocolClient=()=>true;
const electron={app,BrowserWindow:MockWindow,ipcMain:{handle:(key,handler)=>handlers.set(key,handler)},dialog:{},shell:{}};
const fakeRequire=id=>id==='electron'?electron:id==='electron-updater'?{autoUpdater:updater}:id==='./roblox_launcher.cjs'?require(path.join(__dirname,'../electron/roblox_launcher.cjs')):require(id);
const context={require:fakeRequire,__dirname:path.join(__dirname,'../electron'),module:{exports:{}},exports:{},process,console,Buffer,setTimeout:(fn,delay)=>{timers.push({fn,delay});return {unref(){}}},setInterval:(fn,delay)=>{timers.push({fn,delay});return {unref(){}}}};
vm.runInNewContext(source,context,{filename:'electron/main.cjs'});
(async()=>{
  await new Promise(setImmediate);
  assert(handlers.has('updates:check')&&handlers.has('updates:install'));
  assert.equal(updater.autoDownload,true,'automatic download must be enabled');
  assert.equal(updater.autoInstallOnAppQuit,true,'auto install on graceful quit must be enabled');
  assert(timers.some(x=>x.delay===5000),'startup check missing');
  assert(timers.some(x=>x.delay===6*60*60*1000),'periodic check missing');
  await handlers.get('updates:check')();
  assert(updateCalls.includes('check'));
  const repeat=updateCalls.filter(c=>c==='check').length;
  await handlers.get('updates:check')();
  assert.equal(updateCalls.filter(c=>c==='check').length,repeat,'must not redownload a cached ready update');
  assert(sent.some(x=>x[0]==='updates:status'&&x[1].phase==='ready'),'renderer was not notified');
  await handlers.get('updates:install')();
  assert.deepEqual(updateCalls.at(-1),['install',true,true],'must silently install from updater cache and restart');
  assert(!source.includes("app.getPath('downloads')")&&!source.includes('Downloads\\\\'),'must not rely on user Downloads');
  console.log('PASS: updater startup check, periodic check, auto-download, deduped download, ready event, silent cached installation');
})().catch(e=>{console.error(e);process.exitCode=1});
