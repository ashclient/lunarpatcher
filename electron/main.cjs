'use strict';
const {app, BrowserWindow, ipcMain, dialog, shell} = require('electron');
const {execFile, spawn} = require('node:child_process');
const fs = require('node:fs');
const path = require('node:path');
const os = require('node:os');
const {autoUpdater} = require('electron-updater');
const {diagnose:diagnoseRoblox, launchRoblox, cleanPrefs:cleanRobloxPrefs,parseDesktopLink}=require('./roblox_launcher.cjs');
const REPO = 'https://github.com/ashclient/lunarpatcher';
const DEFAULT_FOLDER = path.join(process.env.LOCALAPPDATA || path.join(os.homedir(), 'AppData', 'Local'), 'Programs', 'Lunar Client');
const VALID_MODES = new Set(['ads','layout','sakura','color']);
const HEX = /^#[\da-fA-F]{6}$/;
let win;
let busy = false;
let readyToInstall = false;
let updateCheckInFlight = false;
let updateState = {phase:'idle', message:'Updates are checked automatically'};

// Registered desktop-link handoff from the optional in-Lunar Roblox page.
// Requests are intentionally restricted to numeric game IDs.
function incomingDesktopLink(argv){return (argv||[]).map(parseDesktopLink).find(Boolean)||null;}
async function handoffDesktopLink(argv){
  const request=incomingDesktopLink(argv);
  if(!request)return false;
  try{const outcome=await launchRoblox(request,{shell});
    if(win&&!win.isDestroyed())win.webContents.send('roblox:handoff',outcome);
  }catch(e){if(win&&!win.isDestroyed())win.webContents.send('roblox:handoff',{error:String(e?.message||e)});}
  return true;
}
if(!app.requestSingleInstanceLock())app.quit();
else app.on('second-instance',(_event,argv)=>{
  if(win&&!win.isDestroyed()){win.show();win.focus();}
  if(app.isReady())void handoffDesktopLink(argv);
});
function prefsPath() { return path.join(app.getPath('userData'), 'preferences.json'); }
function getPreferences() {
  try {
    const p = JSON.parse(fs.readFileSync(prefsPath(), 'utf8'));
    return {folder:typeof p.folder==='string' ? p.folder : DEFAULT_FOLDER,
            mode:VALID_MODES.has(p.mode)?p.mode:'sakura',
            color:HEX.test(p.color)?p.color:'#E59CBC', robloxEnabled:p.robloxEnabled===true};
  } catch { return {folder: DEFAULT_FOLDER, mode:'sakura', color:'#E59CBC',robloxEnabled:false}; }
}
function savePreferences(p) {
  const clean = {
    folder:typeof p?.folder==='string' && p.folder.length<1000?p.folder:DEFAULT_FOLDER,
    mode:VALID_MODES.has(p?.mode)?p.mode:'sakura',
    color:HEX.test(p?.color||'')?p.color:'#E59CBC',
    robloxEnabled:p?.robloxEnabled===true
  };
  fs.mkdirSync(path.dirname(prefsPath()), {recursive:true});
  fs.writeFileSync(prefsPath(), JSON.stringify(clean,null,2),{encoding:'utf8',mode:0o600});
  return clean;
}
function robloxPrefsPath(){return path.join(app.getPath('userData'),'roblox-preferences.json');}
function getRobloxPreferences(){try{return cleanRobloxPrefs(JSON.parse(fs.readFileSync(robloxPrefsPath(),'utf8')))}catch{return cleanRobloxPrefs({})}}
function saveRobloxPreferences(v){const p=cleanRobloxPrefs(v);fs.mkdirSync(path.dirname(robloxPrefsPath()),{recursive:true});fs.writeFileSync(robloxPrefsPath(),JSON.stringify(p,null,2),{encoding:'utf8',mode:0o600});return p;}
function appUpdate(state) {
  updateState={...updateState,...state};
  if(win && !win.isDestroyed()) win.webContents.send('updates:status',updateState);
}
function backendPath() {
  return app.isPackaged
    ? path.join(process.resourcesPath, 'engine', 'LunarPatcherEngine.exe')
    : path.join(__dirname, '..', 'build', 'bin', process.platform==='win32'?'LunarPatcherEngine.exe':'lunarpatcher-engine');
}
function asFolder(folder) {
  if(typeof folder !== 'string' || folder.length === 0 || folder.length > 1000 || folder.includes('\0')) throw new Error('Select a valid Lunar Client directory');
  return path.resolve(folder);
}
function execBackend(args) {
  return new Promise((resolve,reject)=> {
    const engine=backendPath();
    if(!fs.existsSync(engine)) return reject(new Error('Patch engine missing. Reinstall LunarPatcher from GitHub Releases.'));
    execFile(engine,args,{windowsHide:true,timeout:120000,maxBuffer:4*1024*1024},(err,stdout,stderr)=> {
      if(err) return reject(new Error((stderr||stdout||err.message).trim().slice(0,1600)));
      resolve((stdout||'').trim());
    });
  });
}
function createWindow() {
  win = new BrowserWindow({
    width:1080,height:790,minWidth:780,minHeight:650,
    show:false,backgroundColor:'#17191e',
    icon:path.join(__dirname,'assets','icon.png'),
    title:'LunarPatcher',
    webPreferences:{preload:path.join(__dirname,'preload.cjs'),contextIsolation:true,nodeIntegration:false,sandbox:true,
      webSecurity:true,devTools:!app.isPackaged}
  });
  win.loadFile(path.join(__dirname,'renderer','index.html'));
  win.once('ready-to-show',()=>win.show());
  win.webContents.setWindowOpenHandler(()=>({action:'deny'}));
  win.webContents.on('will-navigate',(event)=>event.preventDefault());
  win.on('closed',()=>{win=null});
}
async function checkForUpdates() {
  if(!app.isPackaged) return {phase:'development',message:'Automatic updates are enabled in installed builds'};
  if(readyToInstall || updateCheckInFlight) return updateState;
  updateCheckInFlight=true;
  try {
    await autoUpdater.checkForUpdates();
    return updateState;
  } catch (error) {
    appUpdate({phase:'error',message:'Update check failed: '+error.message});
    throw error;
  } finally {
    updateCheckInFlight=false;
  }
}
app.whenReady().then(()=>{
  if(process.platform==='win32'&&app.isPackaged)app.setAsDefaultProtocolClient('lunarpatcher');
  createWindow();
  if(incomingDesktopLink(process.argv))setTimeout(()=>void handoffDesktopLink(process.argv),800);
  ipcMain.handle('app:info',()=>({version:app.getVersion(),platform:process.platform,repository:REPO,update:updateState,defaultFolder:DEFAULT_FOLDER}));
  ipcMain.handle('prefs:read',()=>getPreferences());
  ipcMain.handle('prefs:save',(_event,prefs)=>savePreferences(prefs));
  ipcMain.handle('app:browse',async()=> {
    const res=await dialog.showOpenDialog(win,{title:'Select Lunar Client installation',defaultPath:getPreferences().folder,properties:['openDirectory']});
    return res.canceled?null:res.filePaths[0];
  });
  ipcMain.handle('app:repo',()=>shell.openExternal(REPO));
  ipcMain.handle('roblox:status',()=>diagnoseRoblox());
  ipcMain.handle('roblox:prefs-read',()=>getRobloxPreferences());
  ipcMain.handle('roblox:prefs-save',(_event,prefs)=>saveRobloxPreferences(prefs));
  ipcMain.handle('roblox:launch',async(_event,request)=>{
    const prefs=getRobloxPreferences();
    // Validate the target before allowing any OS-level URL handoff.
    const result=await launchRoblox({target:request?.target,app:request?.app===true,mode:request?.mode||prefs.launchMode},{shell});
    return result;
  });

  ipcMain.handle('app:launch',async(_event,folder)=>{
    const lunarExe=path.join(asFolder(folder),'Lunar Client.exe');
    if(!fs.existsSync(lunarExe)) throw new Error('Lunar Client.exe was not found in that directory');
    const child=spawn(lunarExe,[],{cwd:path.dirname(lunarExe),detached:true,stdio:'ignore',windowsHide:false});
    child.unref();return true;
  });
  ipcMain.handle('engine:run',async(_event,args)=> {
    const operation=args?.operation;
    if(!['status','check','apply','restore'].includes(operation)) throw new Error('Unsupported operation');
    const folder=asFolder(args.folder);
    const mode=VALID_MODES.has(args?.mode)?args.mode:null;
    const color=HEX.test(args?.color||'')?args.color:'#E59CBC';
    const robloxEnabled=args?.robloxEnabled===true;
    if(['check','apply'].includes(operation)&&!mode) throw new Error('Select a valid patch mode');
    if(busy) throw new Error('An operation is already running');
    busy=true;
    try {
      const argv=['--dir',folder];
      if(operation==='status')argv.push('--status');
      else if(operation==='restore')argv.push('--restore');
      else argv.push('--mode',mode,'--color',color,...(robloxEnabled?['--roblox']:[]),...(operation==='check'?['--check']:[]));
      const output=await execBackend(argv);
      if(operation==='restore')return {ok:true,restored:true};
      let result;
      try{result=JSON.parse(output)}catch{throw new Error(`Unexpected patch-engine response: ${output.slice(0,250)}`)}
      return {ok:true,...result};
    } finally {busy=false}
  });
  ipcMain.handle('updates:check',async()=>{
    return checkForUpdates();
  });
  ipcMain.handle('updates:install',()=>{
    if(!readyToInstall) throw new Error('No verified update has been downloaded');
    // electron-updater uses the verified installer from its cache, not Downloads.
    // Silent NSIS update; force the app to relaunch when installation finishes.
    autoUpdater.quitAndInstall(true,true);return true;
  });
  autoUpdater.autoDownload=true;
  autoUpdater.autoInstallOnAppQuit=true;
  autoUpdater.on('checking-for-update',()=>appUpdate({phase:'checking',message:'Checking GitHub Releases…'}));
  autoUpdater.on('update-available',info=>appUpdate({phase:'downloading',version:info.version,message:`Downloading LunarPatcher ${info.version}…`}));
  autoUpdater.on('update-not-available',()=>appUpdate({phase:'current',progress:0,version:null,message:'You’re running the latest version'}));
  autoUpdater.on('download-progress',p=>appUpdate({phase:'downloading',progress:Math.round(p.percent),message:`Downloading update · ${Math.round(p.percent)}%`}));
  autoUpdater.on('update-downloaded',info=>{readyToInstall=true;appUpdate({phase:'ready',version:info.version,progress:100,message:`Version ${info.version} is ready to install`})});
  autoUpdater.on('error',e=>appUpdate({phase:'error',message:`Update check failed: ${e.message}`}));
  if(app.isPackaged) {
    setTimeout(()=>checkForUpdates().catch(()=>{}),5000);
    // Recheck periodically for long-running sessions. An already downloaded
    // update remains ready until restart rather than being downloaded again.
    setInterval(()=>{if(!readyToInstall)checkForUpdates().catch(()=>{});},6*60*60*1000).unref?.();
  }
});
app.on('window-all-closed',()=>{if(process.platform!=='darwin')app.quit()});
app.on('activate',()=>{if(BrowserWindow.getAllWindows().length===0)createWindow()});
