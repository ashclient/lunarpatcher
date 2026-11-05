'use strict';
const $ = id => document.getElementById(id);
const api = window.lunarPatcher || {
  getAppInfo:async()=>({version:'1.1.0',defaultFolder:'C:\\Users\\You\\AppData\\Local\\Programs\\Lunar Client'}),
  getPreferences:async()=>({folder:'C:\\Users\\You\\AppData\\Local\\Programs\\Lunar Client',mode:'sakura',color:'#E59CBC'}),
  savePreferences:async x=>x,
  getStatus:async()=>({found:true,version:'3.7.32-ow',mode:'original',canRestore:false}),
  check:async()=>({ok:true}),apply:async()=>({ok:true,mode:'sakura',ads:'Disabled',announcements:'News-feed component disabled'}),
  restore:async()=>({ok:true}),pickFolder:async()=>null,checkUpdates:async()=>({phase:'current',message:'Latest version'}),
  onUpdate:()=>()=>{},installUpdate:async()=>{},openRepository:async()=>{}
};
let state = {mode:'sakura',color:'#E59CBC',folder:'',busy:false,details:'No recent activity.',page:'home',updatePhase:'idle',currentInstall:null};
let toastTimer;
const MODES={
  ads:{name:'Ads Only',description:'Original Lunar appearance, without recognized ad placements.'},
  layout:{name:'Clean Layout',description:'A simpler homepage with news clutter removed.'},
  sakura:{name:'Sakura',description:'Soft pink, cherry blossoms, and matching panels.'},
  color:{name:'Custom Color',description:'Your own background color throughout the launcher.'}
};
function notify(title,message,error=false){
  $('toastTitle').textContent=title;
  $('toastMessage').textContent=message;
  $('toastSymbol').textContent=error?'!':'✓';
  $('toast').hidden=false;
  clearTimeout(toastTimer);
  toastTimer=setTimeout(()=>{$('toast').hidden=true},7000);
}
function setBusy(flag,label='Ready'){
  state.busy=flag;
  for(const id of ['applyBtn','restoreBtn','appearanceApply','refreshStatus','browseFolder','checkUpdatesBtn']){const x=$(id);if(x)x.disabled=flag}
  $('applyBtn').querySelector('span').textContent=flag?'Working…':'Apply changes';
  $('operationStatus').textContent=label;
}
function setLog(kind,details){
  state.details=`${new Date().toLocaleString()} · ${kind}\n\n${typeof details==='string'?details:JSON.stringify(details,null,2)}`;
  $('detailText').textContent=state.details;
}
function cleanError(e){return String(e?.message||e||'Unknown error').replace(/^ERROR:\s*/,'').slice(0,700)}
function normalizeHex(v){return /^#[0-9a-f]{6}$/i.test(v)?v.toUpperCase():'#E59CBC'}
function syncColor(){
 const color=normalizeHex(state.color);
 $('colorInput').value=color;
 $('colorInputAppearance').value=color;
 $('colorLabel').textContent=color;
 $('colorAppearanceLabel').textContent=color;
 document.documentElement.style.setProperty('--custom-color',color);
}
function selectMode(mode){
 if(!MODES[mode]||state.busy)return;
 state.mode=mode;
 document.querySelectorAll('.mode-item').forEach(btn=>{
  const on=btn.dataset.mode===mode;btn.classList.toggle('selected',on);btn.setAttribute('aria-checked',String(on));
 });
 document.querySelectorAll('.preset').forEach(btn=>btn.classList.toggle('active',btn.dataset.preset===mode));
 $('customColorInline').hidden=mode!=='color';
 $('previewScene').className=`preview-scene mode-${mode}`;
 $('previewTitle').textContent=MODES[mode].name;
 $('previewDescription').textContent=MODES[mode].description;
 persist();
}
function persist(){
 api.savePreferences({mode:state.mode,color:state.color,folder:state.folder}).catch(()=>{});
}
function switchPage(page){
 if(!['home','appearance','updates','about'].includes(page))return;
 state.page=page;
 document.querySelectorAll('.page').forEach(x=>x.classList.toggle('active',x.id==='page-'+page));
 document.querySelectorAll('[data-page]').forEach(x=>x.classList.toggle('selected',x.dataset.page===page));
 $('breadcrumbCurrent').textContent={home:'Overview',appearance:'Appearance',updates:'Updates',about:'About'}[page];
 document.querySelector('.content').scrollTop=0;
}
async function refreshStatus(silent=false){
 state.folder=$('folderInput').value.trim();
 if(!state.folder){$('installedVersion').textContent='Choose a Lunar Client installation folder';$('installStatus').textContent='Not configured';return}
 try{
  const info=await api.getStatus(state.folder);
  state.currentInstall=info;
  $('installedVersion').textContent=`Lunar Client ${info.version} · ${info.folder || state.folder}`;
  $('installStatus').className='status-pill '+(info.changedSincePatch?'warn':'good');
  $('installStatus').innerHTML='<span class="small-dot"></span>';
  $('installStatus').append(document.createTextNode(info.changedSincePatch?'Version changed':info.mode&&info.mode!=='original'?'Customized':'Detected'));
  $('restoreBtn').disabled=state.busy||!info.canRestore;
  if(info.changedSincePatch&&!silent) notify('Launcher changed','Lunar appears to have updated since the patch. Backups will not overwrite newer files.',true);
 }catch(e){
  state.currentInstall=null;
  $('installedVersion').textContent='Lunar Client could not be detected';
  $('installStatus').className='status-pill warn';
  $('installStatus').innerHTML='<span class="small-dot"></span>Not found';
  $('restoreBtn').disabled=true;
  if(!silent)notify('Check installation',cleanError(e),true);
 }
}
async function apply(){
 if(state.busy)return;
 state.folder=$('folderInput').value.trim();
 persist();
 setBusy(true,'Checking compatibility…');
 try{
  const checked=await api.check(state.folder,state.mode,state.color);
  setLog('Compatibility check',checked);
  setBusy(true,'Applying patch; keep Lunar closed…');
  const applied=await api.apply(state.folder,state.mode,state.color);
  setLog('Patch applied',applied);
  notify('Changes applied',`${MODES[state.mode].name} is installed. You can now launch Lunar normally.`);
 }catch(e){
  const msg=cleanError(e);
  setLog('Patch error',msg);
  notify('Patch not applied',msg,true);
 }finally{
  setBusy(false);
  await refreshStatus(true);
 }
}
async function restore(){
 if(state.busy)return;
 state.folder=$('folderInput').value.trim();
 if(!window.confirm('Restore the original Lunar Client files from your verified backup?\n\nClose Lunar Client before continuing.'))return;
 setBusy(true,'Restoring original files…');
 try{
  const result=await api.restore(state.folder);
  setLog('Restore',result);
  notify('Original restored','Your verified original launcher files are back in place.');
 }catch(e){const msg=cleanError(e);setLog('Restore error',msg);notify('Unable to restore',msg,true)}
 finally{setBusy(false);await refreshStatus(true)}
}
function renderUpdate(status){
 if(!status)return;
 state.updatePhase=status.phase||'idle';
 $('updateHeader').textContent={checking:'Checking for updates',downloading:'Downloading an update',ready:'Update ready to install',current:'You’re up to date',error:'Update unavailable',development:'Development build',idle:'Keeping you up to date'}[state.updatePhase]||'Application updates';
 $('updateMessage').textContent=status.message||'Updates are managed through GitHub Releases.';
 $('updateProgressOuter').hidden=state.updatePhase!=='downloading';
 $('updateProgress').style.width=Math.min(100,Math.max(0,Number(status.progress||0)))+'%';
 $('restartUpdateBtn').hidden=state.updatePhase!=='ready';
 $('checkUpdatesBtn').disabled=state.updatePhase==='checking'||state.updatePhase==='downloading';
 $('updateDot').hidden=state.updatePhase!=='ready';
}
function wire(){
 document.querySelectorAll('[data-page]').forEach(btn=>btn.addEventListener('click',()=>switchPage(btn.dataset.page)));
 document.querySelectorAll('[data-mode]').forEach(btn=>btn.addEventListener('click',()=>selectMode(btn.dataset.mode)));
 document.querySelectorAll('[data-preset]').forEach(btn=>btn.addEventListener('click',()=>selectMode(btn.dataset.preset)));
 for(const id of ['colorInput','colorInputAppearance']) $(id).addEventListener('input',event=>{
   state.color=normalizeHex(event.target.value);syncColor();persist();
 });
 $('folderInput').addEventListener('change',()=>{state.folder=$('folderInput').value.trim();persist();refreshStatus(true)});
 $('browseFolder').addEventListener('click',async()=>{
   try{const picked=await api.pickFolder();if(picked){state.folder=picked;$('folderInput').value=picked;persist();refreshStatus()}}
   catch(e){notify('Folder unavailable',cleanError(e),true)}
 });
 $('refreshStatus').addEventListener('click',()=>refreshStatus());
 $('applyBtn').addEventListener('click',apply);
 $('appearanceApply').addEventListener('click',()=>{switchPage('home');apply()});
 $('restoreBtn').addEventListener('click',restore);
 $('checkUpdatesBtn').addEventListener('click',async()=>{try{renderUpdate({phase:'checking',message:'Checking GitHub Releases…'});renderUpdate(await api.checkUpdates())}catch(e){renderUpdate({phase:'error',message:cleanError(e)})}});
 $('restartUpdateBtn').addEventListener('click',async()=>{try{await api.installUpdate()}catch(e){notify('Update error',cleanError(e),true)}});
 $('topGithub').addEventListener('click',()=>api.openRepository());
 $('aboutGithub').addEventListener('click',()=>api.openRepository());
 $('viewReleases').addEventListener('click',()=>api.openRepository());
 $('toastClose').addEventListener('click',()=>{$('toast').hidden=true});
 $('showDetailsBtn').addEventListener('click',()=>$('detailsDialog').showModal());
 $('closeDialog').addEventListener('click',()=>$('detailsDialog').close());
 api.onUpdate(renderUpdate);
}
async function init(){
 wire();
 try{
  const [p,app]=await Promise.all([api.getPreferences(),api.getAppInfo()]);
  state.folder=p.folder||app.defaultFolder||'';
  state.color=normalizeHex(p.color||'#E59CBC');
  $('folderInput').value=state.folder;
  $('appVersion').textContent='v'+app.version;
  $('aboutVersion').textContent=app.version;
  syncColor();
  selectMode(MODES[p.mode]?p.mode:'sakura');
  if(app.update) renderUpdate(app.update);
  await refreshStatus(true);
 }catch(e){setLog('Initialization error',cleanError(e));notify('Initialization error',cleanError(e),true)}
}
document.addEventListener('DOMContentLoaded',init);
