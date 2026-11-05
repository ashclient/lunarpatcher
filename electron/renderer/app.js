'use strict';
const api = window.lunarPatcher;
const $ = id => document.getElementById(id);
const MODE_INFO = {
  ads: {name:'Ads Only', text:'Original Lunar Client interface with supported advertisements removed.'},
  layout: {name:'Clean Layout', text:'Expanded friends list on the right, Quick Play and shortcuts below Launch, and news in Updates.'},
  sakura: {name:'Sakura', text:'Fixed cherry-blossom artwork, rose-colored panels and the Clean Layout features.'},
  color: {name:'Custom Color', text:'A solid-color theme using your chosen color and all Clean Layout features.'}
};
const state = {mode:'sakura',color:'#E59CBC',folder:'',busy:false,info:null,update:null,log:'No operation yet.'};
const safeError = e => String(e?.message || e || 'Unknown error').replace(/^ERROR:\s*/i,'').slice(0,1100);
function setMessage(text,type='') { $('actionMessage').textContent=text; $('actionMessage').className='action-message'+(type?' '+type:''); }
function setBusy(on,text) {
  state.busy=on;
  for(const id of ['applyBtn','browseBtn','refreshBtn','folderInput','colorPicker']) $(id).disabled=on;
  document.querySelectorAll('input[name="mode"]').forEach(i=>i.disabled=on);
  $('restoreBtn').disabled=on||!state.info?.canRestore||!!state.info?.changedSincePatch;
  if(text)setMessage(text);
  $('applyBtn').textContent=on?'Working…':'Apply changes';
}
function persist(){if(!api)return;api.savePreferences({folder:state.folder,mode:state.mode,color:state.color}).catch(()=>{});}
function setMode(mode){
  if(!MODE_INFO[mode]||state.busy)return;
  state.mode=mode;
  for(const input of document.querySelectorAll('input[name="mode"]')){
    input.checked=input.value===mode;
    input.closest('.mode-row').classList.toggle('selected',input.checked);
  }
  $('colorRow').hidden=mode!=='color';
  $('themePreview').className='theme-preview '+mode;
  $('themePreview').style.setProperty('--preview-color',state.color);
  $('modeDescription').textContent=MODE_INFO[mode].text;
  persist();
}
function setColor(hex){
  if(!/^#[\da-f]{6}$/i.test(hex))return;
  state.color=hex.toUpperCase();
  $('colorPicker').value=hex;
  $('colorHex').textContent=state.color;
  $('themePreview').style.setProperty('--preview-color',state.color);
  persist();
}
function updateStatusPill(text,type=''){
  const pill=$('statusPill');pill.className='status-pill'+(type?' '+type:'');
  pill.lastChild.textContent=text;
}
async function refresh(silent=false){
  state.folder=$('folderInput').value.trim();
  if(!api||!state.folder){$('statusTitle').textContent='Choose a Lunar Client folder';$('statusDetails').textContent='Install location not set';updateStatusPill('Not found','warn');return;}
  try {
    const info=await api.getStatus(state.folder);
    state.info=info;
    $('statusTitle').textContent='Lunar Client '+info.version;
    $('statusDetails').textContent=info.mode&&info.mode!=='original'?'Installed mode: '+(MODE_INFO[info.mode]?.name||info.mode):'Original launcher detected';
    const updated=!!info.changedSincePatch;
    updateStatusPill(updated?'Updated outside app':info.mode==='original'?'Detected':'Patched',updated?'warn':'success');
    $('restoreBtn').disabled=state.busy||!info.canRestore||updated;
    if(updated&&!silent)setMessage('Lunar was updated. Use a fresh compatibility check; old backups are preserved.','error');
  } catch(e) {
    state.info=null;
    $('statusTitle').textContent='Lunar Client not found';
    $('statusDetails').textContent='Check the installation folder below';
    updateStatusPill('Not found','warn');
    $('restoreBtn').disabled=true;
    if(!silent)setMessage(safeError(e),'error');
  }
}
function remember(label,data){
  state.log=label+'\n\n'+(typeof data==='string'?data:JSON.stringify(data,null,2));
  $('detailsText').textContent=state.log;
}
async function apply(){
  if(state.busy)return;
  state.folder=$('folderInput').value.trim();persist();
  setBusy(true,'Checking launcher compatibility…');
  try {
    const checked=await api.check(state.folder,state.mode,state.color);
    remember('Compatibility check',checked);
    setMessage('Applying '+MODE_INFO[state.mode].name+'…');
    const result=await api.apply(state.folder,state.mode,state.color);
    remember('Installed '+MODE_INFO[state.mode].name,result);
    setMessage('Changes applied. Open Lunar normally.','success');
  }catch(e){const text=safeError(e);remember('Error',text);setMessage(text,'error');}
  finally{setBusy(false);await refresh(true);}
}
async function restore(){
  if(state.busy)return;
  if(!confirm('Restore Lunar Client from the verified original backup?\n\nClose Lunar Client before continuing.'))return;
  setBusy(true,'Restoring the original launcher…');
  try{const result=await api.restore(state.folder);remember('Restore original',result);setMessage('Original files restored.','success');}
  catch(e){const text=safeError(e);remember('Restore error',text);setMessage(text,'error');}
  finally{setBusy(false);await refresh(true);}
}
function renderUpdate(u){
  if(!u)return;state.update=u;
  const phase=u.phase||'idle';
  $('updateTitle').textContent={idle:'Update status',checking:'Checking for updates',downloading:'Downloading update',ready:'Update ready',current:'You are up to date',development:'Development build',error:'Update check failed'}[phase]||'Update status';
  $('updateDescription').textContent=u.message||'Updates are distributed through GitHub Releases.';
  $('updateTrack').hidden=phase!=='downloading';
  $('updateProgress').style.width=Math.min(100,Math.max(0,Number(u.progress)||0))+'%';
  $('installUpdateBtn').hidden=phase!=='ready';
  $('updateBadge').hidden=phase!=='ready';
  $('checkUpdatesBtn').disabled=phase==='checking'||phase==='downloading';
}
function showToast(message){
  $('toast').textContent=message;$('toast').hidden=false;
}
function wire(){
  document.querySelectorAll('input[name="mode"]').forEach(input=>input.addEventListener('change',()=>setMode(input.value)));
  $('colorPicker').addEventListener('input',e=>setColor(e.target.value));
  $('folderInput').addEventListener('change',()=>{state.folder=$('folderInput').value.trim();persist();refresh();});
  $('browseBtn').addEventListener('click',async()=>{try{const path=await api.pickFolder();if(path){state.folder=path;$('folderInput').value=path;persist();refresh();}}catch(e){setMessage(safeError(e),'error');}});
  $('refreshBtn').addEventListener('click',()=>refresh());
  $('applyBtn').addEventListener('click',apply);
  $('restoreBtn').addEventListener('click',restore);
  $('updatesBtn').addEventListener('click',()=>{$('updatesDialog').showModal();});
  $('detailsBtn').addEventListener('click',()=>{$('detailsDialog').showModal();});
  $('githubBtn').addEventListener('click',()=>api.openRepository());
  $('checkUpdatesBtn').addEventListener('click',async()=>{renderUpdate({phase:'checking',message:'Checking for releases…'});try{renderUpdate(await api.checkUpdates());}catch(e){renderUpdate({phase:'error',message:safeError(e)});}});
  $('installUpdateBtn').addEventListener('click',async()=>{try{await api.installUpdate();}catch(e){showToast(safeError(e));}});
  document.querySelectorAll('[data-close]').forEach(b=>b.addEventListener('click',()=>$(b.dataset.close).close()));
  $('toast').addEventListener('click',()=>{$('toast').hidden=true;});
  api.onUpdate(renderUpdate);
}
async function init(){
  if(!api){setMessage('Desktop integration unavailable. Open this interface inside LunarPatcher.','error');return;}
  wire();
  try {
    const [pref,info]=await Promise.all([api.getPreferences(),api.getAppInfo()]);
    state.folder=pref.folder||info.defaultFolder||'';
    $('folderInput').value=state.folder;
    $('appVersion').textContent='v'+info.version;
    setColor(pref.color||'#E59CBC');
    setMode(MODE_INFO[pref.mode]?pref.mode:'sakura');
    if(info.update)renderUpdate(info.update);
    await refresh(true);
  } catch(e){setMessage(safeError(e),'error');}
}
document.addEventListener('DOMContentLoaded',init);
