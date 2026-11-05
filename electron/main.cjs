'use strict';
const {app, BrowserWindow, ipcMain, dialog, shell} = require('electron');
const {execFile, spawn} = require('node:child_process');
const fs = require('node:fs');
const path = require('node:path');
const os = require('node:os');
const {autoUpdater} = require('electron-updater');
const REPO = 'https://github.com/ashclient/lunarpatcher';
const DEFAULT_FOLDER = path.join(process.env.LOCALAPPDATA || path.join(os.homedir(), 'AppData', 'Local'), 'Programs', 'Lunar Client');
const VALID_MODES = new Set(['ads','layout','sakura','color']);
const HEX = /^#[\da-fA-F]{6}$/;
let win;
let busy = false;
let readyToInstall = false;
let updateState = {phase:'idle', message:'Updates are checked automatically'};

function prefsPath() { return path.join(app.getPath('userData'), 'preferences.json'); }
function getPreferences() {
  try {
    const p = JSON.parse(fs.readFileSync(prefsPath(), 'utf8'));
    return {folder:typeof p.folder==='string' ? p.folder : DEFAULT_FOLDER,
            mode:VALID_MODES.has(p.mode)?p.mode:'sakura',
            color:HEX.test(p.color)?p.color:'#E59CBC'};
  } catch { return {folder: DEFAULT_FOLDER, mode:'sakura', color:'#E59CBC'}; }
}
function savePreferences(p) {
  const clean = {
    folder:typeof p?.folder==='string' && p.folder.length<1000?p.folder:DEFAULT_FOLDER,
    mode:VALID_MODES.has(p?.mode)?p.mode:'sakura',
    color:HEX.test(p?.color||'')?p.color:'#E59CBC'
  };
  fs.mkdirSync(path.dirname(prefsPath()), {recursive:true});
  fs.writeFileSync(prefsPath(), JSON.stringify(clean,null,2),{encoding:'utf8',mode:0o600});
  return clean;
}
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
    width:1130,height:750,minWidth:940,minHeight:650,
    show:false,backgroundColor:'#141118',
    icon:path.join(__dirname,'assets','icon.png'),
    title:'LunarPatcher',
    ...(process.platform==='win32'?{titleBarStyle:'hidden',titleBarOverlay:{color:'#19151F',symbolColor:'#E9DBE5',height:46}}:{}),
    webPreferences:{preload:path.join(__dirname,'preload.cjs'),contextIsolation:true,nodeIntegration:false,sandbox:true,
      webSecurity:true,devTools:!app.isPackaged}
  });
  win.loadFile(path.join(__dirname,'renderer','index.html'));
  win.once('ready-to-show',()=>win.show());
  win.webContents.setWindowOpenHandler(()=>({action:'deny'}));
  win.webContents.on('will-navigate',(event)=>event.preventDefault());
  win.on('closed',()=>{win=null});
}
app.whenReady().then(()=>{
  createWindow();
  ipcMain.handle('app:info',()=>({version:app.getVersion(),platform:process.platform,repository:REPO,update:updateState,defaultFolder:DEFAULT_FOLDER}));
  ipcMain.handle('prefs:read',()=>getPreferences());
  ipcMain.handle('prefs:save',(_event,prefs)=>savePreferences(prefs));
  ipcMain.handle('app:browse',async()=> {
    const res=await dialog.showOpenDialog(win,{title:'Select Lunar Client installation',defaultPath:getPreferences().folder,properties:['openDirectory']});
    return res.canceled?null:res.filePaths[0];
  });
  ipcMain.handle('app:repo',()=>shell.openExternal(REPO));
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
    if(['check','apply'].includes(operation)&&!mode) throw new Error('Select a valid patch mode');
    if(busy) throw new Error('An operation is already running');
    busy=true;
    try {
      const argv=['--dir',folder];
      if(operation==='status')argv.push('--status');
      else if(operation==='restore')argv.push('--restore');
      else argv.push('--mode',mode,'--color',color,...(operation==='check'?['--check']:[]));
      const output=await execBackend(argv);
      if(operation==='restore')return {ok:true,restored:true};
      let result;
      try{result=JSON.parse(output)}catch{throw new Error(`Unexpected patch-engine response: ${output.slice(0,250)}`)}
      return {ok:true,...result};
    } finally {busy=false}
  });
  ipcMain.handle('updates:check',async()=>{
    if(!app.isPackaged) return {phase:'development',message:'Automatic updates are available in installed builds'};
    await autoUpdater.checkForUpdates();
    return updateState;
  });
  ipcMain.handle('updates:install',()=>{
    if(!readyToInstall) throw new Error('No verified update has been downloaded');
    autoUpdater.quitAndInstall(false,true);return true;
  });
  autoUpdater.autoDownload=true;
  autoUpdater.autoInstallOnAppQuit=true;
  autoUpdater.on('checking-for-update',()=>appUpdate({phase:'checking',message:'Checking GitHub Releases…'}));
  autoUpdater.on('update-available',info=>appUpdate({phase:'downloading',version:info.version,message:`Downloading LunarPatcher ${info.version}…`}));
  autoUpdater.on('update-not-available',()=>appUpdate({phase:'current',message:'You’re running the latest version'}));
  autoUpdater.on('download-progress',p=>appUpdate({phase:'downloading',progress:Math.round(p.percent),message:`Downloading update · ${Math.round(p.percent)}%`}));
  autoUpdater.on('update-downloaded',info=>{readyToInstall=true;appUpdate({phase:'ready',version:info.version,progress:100,message:`Version ${info.version} is ready to install`})});
  autoUpdater.on('error',e=>appUpdate({phase:'error',message:`Update check failed: ${e.message}`}));
  if(app.isPackaged) setTimeout(()=>autoUpdater.checkForUpdates().catch(()=>{}),3000);
});
app.on('window-all-closed',()=>{if(process.platform!=='darwin')app.quit()});
app.on('activate',()=>{if(BrowserWindow.getAllWindows().length===0)createWindow()});
