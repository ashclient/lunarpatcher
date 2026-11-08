'use strict';
const api = window.lunarPatcher;
const $ = id => document.getElementById(id);
const MODES = Object.freeze({
  ads: {title:'Ads Only', kicker:'Original Lunar styling', description:'The original launcher appearance with recognized advertisements removed.'},
  layout: {title:'Clean Layout', kicker:'A simpler homepage', description:'Full friends list, Quick Play, shortcuts and a separate Updates page.'},
  sakura: {title:'Sakura', kicker:'Sakura · fixed theme', description:'Cherry-blossom artwork, soft pink panels and all Clean Layout features.'},
  color: {title:'Custom Color', kicker:'Your own palette', description:'Custom background and matching accents with all Clean Layout features.'}
});
const state = {mode:'sakura',color:'#E59CBC',folder:'',robloxEnabled:false,busy:false,info:null,update:null,log:'No operation yet.',refreshSeq:0};
const friendlyError = e => String(e?.message || e || 'Unknown error').replace(/^ERROR:\s*/i,'').slice(0,1200);
function showMessage(text,kind='') { const el=$('actionMessage');el.textContent=text;el.className='action-copy'+(kind?' '+kind:''); }
function persist() { if(api) api.savePreferences({folder:state.folder,mode:state.mode,color:state.color,robloxEnabled:state.robloxEnabled}).catch(()=>{}); }
function showView(view) {
  if(!['appearance','roblox','updates','about'].includes(view))return;
  document.querySelectorAll('.nav-item[data-view]').forEach(button=>{
    const active=button.dataset.view===view;
    button.classList.toggle('current',active);
    if(active)button.setAttribute('aria-current','page');else button.removeAttribute('aria-current');
  });
  for(const name of ['appearance','roblox','updates','about'])$(name+'Page').hidden=name!==view;
  $('actionBar').hidden=view!=='appearance';
  $('topbarTitle').textContent={appearance:'Appearance',roblox:'Roblox',updates:'Updates',about:'About'}[view];
  $('viewport').scrollTop=0;
}
function chooseMode(mode) {
  if(!MODES[mode] || state.busy)return;
  state.mode=mode;
  document.querySelectorAll('input[name="mode"]').forEach(input=>{
    input.checked=input.value===mode;
    input.closest('.mode-option').classList.toggle('selected',input.checked);
  });
  const preview=$('themePreview');preview.className='theme-preview '+mode;
  preview.style.setProperty('--preview-color',state.color);
  preview.setAttribute('aria-label',MODES[mode].kicker+' preview');
  $('previewCaption').textContent=MODES[mode].title;
  $('previewKicker').textContent=MODES[mode].kicker;
  $('modeDescription').textContent=MODES[mode].description;
  $('colorRow').hidden=mode!=='color';
  persist();
}
function chooseColor(value) {
  if(!/^#[0-9a-f]{6}$/i.test(value))return;
  state.color=value.toUpperCase();
  $('colorPicker').value=value;
  $('colorHex').textContent=state.color;
  $('themePreview').style.setProperty('--preview-color',state.color);
  persist();
}
function setBusy(value,description) {
  state.busy=!!value;
  for(const id of ['applyBtn','browseBtn','refreshBtn','folderInput','colorPicker','robloxToggle'])$(id).disabled=state.busy;
  document.querySelectorAll('input[name="mode"]').forEach(input=>{input.disabled=state.busy;});
  $('restoreBtn').disabled=state.busy || !state.info?.canRestore || !!state.info?.changedSincePatch;
  $('applyBtn').textContent=state.busy?'Applying…':'Apply changes';
  if(description)showMessage(description);
}
function setStatus(label,kind='') {
  $('statusLabel').textContent=label;
  $('statusPill').className='state-tag'+(kind?' '+kind:'');
}
async function refresh(silent=false) {
  const request=++state.refreshSeq;
  state.folder=$('folderInput').value.trim();
  if(!api||!state.folder){state.info=null;$('statusDetails').textContent='Select the installation folder below';setStatus('Not found','warn');$('restoreBtn').disabled=true;return;}
  try {
    const status=await api.getStatus(state.folder);
    if(request!==state.refreshSeq)return;
    state.info=status;
    const patched=status.mode && status.mode!=='original';
    $('installationHeading').textContent='Lunar Client '+(status.version||'');
    $('statusDetails').textContent=status.changedSincePatch?'Lunar files changed outside LunarPatcher':patched?('Installed: '+(MODES[status.mode]?.title||status.mode)+(status.robloxEnabled?' · Roblox enabled':'')):'Original launcher detected';
    setStatus(status.changedSincePatch?'Modified':patched?'Customized':'Detected',status.changedSincePatch?'warn':'success');
    $('restoreBtn').disabled=state.busy||!status.canRestore||!!status.changedSincePatch;
    if(status.changedSincePatch&&!silent)showMessage('Launcher files changed. Inspect compatibility before applying again.','error');
  } catch(e) {
    if(request!==state.refreshSeq)return;
    state.info=null;
    $('installationHeading').textContent='Lunar Client';
    $('statusDetails').textContent='Not found at this location';
    setStatus('Not found','warn');$('restoreBtn').disabled=true;
    if(!silent)showMessage(friendlyError(e),'error');
  }
}
function record(label,value) {
  state.log=label+'\n\n'+(typeof value==='string'?value:JSON.stringify(value,null,2));
  $('detailsText').textContent=state.log;
}
async function applyPatch() {
  if(state.busy)return;
  state.folder=$('folderInput').value.trim();persist();
  if(!state.folder){showMessage('Choose the Lunar Client installation folder first.','error');return;}
  setBusy(true,'Checking compatibility…');
  try {
    const check=await api.check(state.folder,state.mode,state.color,state.robloxEnabled);
    record('Compatibility check',check);
    showMessage('Applying '+MODES[state.mode].title+'…');
    const result=await api.apply(state.folder,state.mode,state.color,state.robloxEnabled);
    record('Applied '+MODES[state.mode].title,result);
    showMessage('Changes applied. You can open Lunar Client normally.','success');
  }catch(e){const message=friendlyError(e);record('Patch failed',message);showMessage(message,'error');}
  finally{setBusy(false);await refresh(true);}
}
async function restoreOriginal() {
  if(state.busy)return;
  $('restoreDialog').close();
  setBusy(true,'Restoring verified backups…');
  try {const result=await api.restore(state.folder);record('Restored original',result);showMessage('Original launcher restored.','success');}
  catch(e){const message=friendlyError(e);record('Restore failed',message);showMessage(message,'error');}
  finally{setBusy(false);await refresh(true);}
}
function renderUpdate(update) {
  if(!update)return;
  state.update=update;
  const phase=update.phase||'idle';
  $('updateTitle').textContent={idle:'Automatic updates are on',checking:'Checking for updates',downloading:'Downloading update',ready:'Update ready to install',current:'You’re up to date',development:'Development build',error:'Could not check for updates'}[phase]||'Update status';
  $('updateDescription').textContent=update.message||'Updates are downloaded within LunarPatcher.';
  $('updateTopline').textContent=phase==='ready'?'Update ready · '+(update.version||'new version'):phase==='downloading'?'Downloading update…':'Automatic updates enabled';
  $('updateIndicator').hidden=phase!=='ready';
  $('updateTrack').hidden=phase!=='downloading';
  $('updateProgress').style.width=Math.min(100,Math.max(0,Number(update.progress)||0))+'%';
  $('installUpdateBtn').hidden=phase!=='ready';
  $('checkUpdatesBtn').disabled=phase==='checking'||phase==='downloading';
}
function showToast(message) { $('toast').textContent=message;$('toast').hidden=false; }
function wireEvents() {
  document.querySelectorAll('[data-view]').forEach(button=>button.addEventListener('click',()=>showView(button.dataset.view)));
  document.querySelectorAll('input[name="mode"]').forEach(input=>input.addEventListener('change',()=>chooseMode(input.value)));
  $('colorPicker').addEventListener('input',event=>chooseColor(event.target.value));
  $('robloxToggle').addEventListener('change',event=>{state.robloxEnabled=event.target.checked;persist();});
  $('folderInput').addEventListener('change',()=>{state.folder=$('folderInput').value.trim();persist();refresh();});
  $('browseBtn').addEventListener('click',async()=>{try{const folder=await api.pickFolder();if(folder){state.folder=folder;$('folderInput').value=folder;persist();await refresh();}}catch(e){showMessage(friendlyError(e),'error');}});
  $('refreshBtn').addEventListener('click',()=>refresh());
  $('applyBtn').addEventListener('click',applyPatch);
  $('restoreBtn').addEventListener('click',()=>{if(!state.busy&&state.info?.canRestore)$('restoreDialog').showModal();});
  $('confirmRestoreBtn').addEventListener('click',restoreOriginal);
  $('detailsBtn').addEventListener('click',()=>$('detailsDialog').showModal());
  $('githubBtn').addEventListener('click',()=>api.openRepository());
  $('aboutGithubBtn').addEventListener('click',()=>api.openRepository());
  $('checkUpdatesBtn').addEventListener('click',async()=>{
    renderUpdate({phase:'checking',message:'Checking GitHub Releases…'});
    try{renderUpdate(await api.checkUpdates());}catch(e){renderUpdate({phase:'error',message:friendlyError(e)});}
  });
  $('installUpdateBtn').addEventListener('click',async()=>{try{await api.installUpdate();}catch(e){showToast(friendlyError(e));}});
  document.querySelectorAll('[data-close]').forEach(button=>button.addEventListener('click',()=>$(button.dataset.close).close()));
  $('toast').addEventListener('click',()=>$('toast').hidden=true);
  api.onUpdate(renderUpdate);
}
async function initialize() {
  if(!api){showMessage('Open this window in the installed LunarPatcher application.','error');return;}
  wireEvents();
  try {
    const [prefs,info]=await Promise.all([api.getPreferences(),api.getAppInfo()]);
    state.folder=prefs.folder||info.defaultFolder||'';
    $('folderInput').value=state.folder;
    $('appVersion').textContent='v'+info.version;
    $('aboutVersion').textContent=info.version;
    chooseColor(prefs.color||'#E59CBC');
    chooseMode(MODES[prefs.mode]?prefs.mode:'sakura');
    state.robloxEnabled=prefs.robloxEnabled===true;
    $('robloxToggle').checked=state.robloxEnabled;
    renderUpdate(info.update);
    await refresh(true);
  }catch(e){showMessage(friendlyError(e),'error');}
}
document.addEventListener('DOMContentLoaded',initialize);
