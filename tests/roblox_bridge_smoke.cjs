'use strict';
const assert=require('node:assert/strict');
const fs=require('node:fs');
const vm=require('node:vm');
const {EventEmitter}=require('node:events');
const path=require('node:path');
const handles=new Map(), opened=[], windows=[], views=[], permissions=[];
const mainContents=new EventEmitter();
mainContents.getURL=()=> 'file:///C:/Users/test/AppData/Local/Programs/Lunar%20Client/resources/app.asar/dist/html/main.html';
mainContents.isDestroyed=()=>false;
mainContents.send=(...args)=>mainContents.emit('message',...args);
class BrowserView {
 constructor({webPreferences}){
  this.preferences=webPreferences;
  this.webContents=new EventEmitter();
  this.webContents.getURL=()=>this.url||'';
  this.webContents.isDestroyed=()=>!!this.destroyed;
  this.webContents.destroy=()=>{this.destroyed=true};
  this.webContents.loadURL=async url=>{this.url=url;return true};
  this.webContents.canGoBack=()=>false;
  this.webContents.canGoForward=()=>false;
  this.webContents.goBack=()=>{};
  this.webContents.goForward=()=>{};
  this.webContents.reload=()=>{};
  this.webContents.setWindowOpenHandler=handler=>{this.windowHandler=handler};
  views.push(this);
 }
 setBounds(bounds){this.bounds=bounds}
}
const win=new EventEmitter();win.id=7;win.webContents=mainContents;win.isDestroyed=()=>false;
win.addBrowserView=v=>{win.child=v;};win.removeBrowserView=v=>{if(win.child===v)win.child=null};win.setTopBrowserView=v=>{win.top=v;};windows.push(win);
const electron={ipcMain:{handle:(k,v)=>handles.set(k,v)},BrowserView,BrowserWindow:{fromWebContents:contents=>contents===mainContents?win:null},shell:{openExternal:async url=>opened.push(url)},session:{fromPartition:key=>({setPermissionRequestHandler:cb=>permissions.push({key,cb})})}};
const context={require:id=>{assert.equal(id,'electron');return electron},URL,Promise,Map,Object,Error,Number,String,Math,RegExp,console};
vm.runInNewContext(fs.readFileSync(path.join(__dirname,'../engine/assets/roblox_bridge_main.js'),'utf8'),context);
(async()=>{
 const call=(name,...args)=>handles.get('lunarpatcher:roblox:'+name)({sender:mainContents},...args);
 assert.equal(handles.size,9);
 await call('show',{x:90,y:120,width:600,height:550});
 assert.equal(views.length,1);assert.deepEqual(JSON.parse(JSON.stringify(win.child.bounds)),{x:90,y:120,width:600,height:550});
 assert.equal(views[0].preferences.nodeIntegration,false);
 assert.equal(views[0].preferences.sandbox,true);assert.equal(views[0].preferences.webSecurity,true);
 assert.equal(views[0].preferences.partition,'persist:lunarpatcher-roblox');
 assert.equal(permissions[0].key,'persist:lunarpatcher-roblox');
 let denied;permissions[0].cb(null,'camera',allow=>denied=allow);assert.equal(denied,false);
 assert.equal(views[0].url,'https://www.roblox.com/home');
 await call('navigate','https://www.roblox.com/games/2753915549/test');
 await assert.rejects(call('navigate','https://evil.example/phish'),/Only https:\/\/roblox\.com/);
 await assert.rejects(call('external','file:///C:/secret'),/External navigation restricted/);
 const event={preventDefault(){this.prevented=true}};
 views[0].webContents.emit('will-navigate',event,'https://evil.example');assert.equal(event.prevented,true);
 const handler=views[0].windowHandler;
 assert.equal(handler({url:'https://roblox.com/games/42'}).action,'deny');
 assert.equal(handler({url:'https://evil.example'}).action,'deny');
 handler({url:'roblox-player:1+launchmode:play+gameinfo:test'});
 await new Promise(resolve=>setImmediate(resolve));assert.equal(opened.at(-1),'roblox-player:1+launchmode:play+gameinfo:test');
 await call('bounds',{x:120,y:70,width:640,height:470});assert.equal(views[0].bounds.x,120);
 await call('hide');assert.equal(win.child,null);assert.equal(views[0].destroyed,true);
 assert.throws(()=>handles.get('lunarpatcher:roblox:show')({sender:{getURL:()=> 'https://evil.example'}} ,{x:1,y:1,width:600,height:500}),/not the Lunar launcher/);
 console.log('PASS: BrowserView lifecycle, permission policy, safe navigation, Roblox Player launch, hide, IPC origin');
})().catch(e=>{console.error(e);process.exitCode=1});
