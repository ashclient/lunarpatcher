'use strict';
// Desktop Roblox controls. All OS requests go through a constrained Electron IPC bridge.
(function () {
  const api=window.lunarPatcher;
  const $=id=>document.getElementById(id);
  const names={'2753915549':'Blox Fruits','6516141723':'DOORS','4924922222':'Brookhaven RP','17625359962':'RIVALS'};
  let prefs={launchMode:'direct',fps:'default',layout:'comfortable',accent:'#E5A1BB',showDetails:true,keepHistory:true,confirmLaunch:false,favorites:[],recent:[]};
  let lastStatus=null;
  const note=(s,isError=false)=>{const el=$('robloxNotice');el.textContent=s;el.className='roblox-notice '+(isError?'error':'success');};
  const log=message=>{$('robloxLaunchLog').textContent=new Date().toLocaleTimeString()+'  '+message+'\n'+$('robloxLaunchLog').textContent.slice(0,2200)};
  function idFrom(value){
    const str=String(value||'').trim();if(/^[1-9]\d{0,17}$/.test(str))return str;
    try{const u=new URL(str);if(!['www.roblox.com','roblox.com'].includes(u.hostname.toLowerCase())||u.protocol!=='https:')return null;
      const m=u.pathname.match(/^\/games\/([1-9]\d{0,17})(?:\/|$)/);return m?m[1]:null;
    }catch{return null;}
  }
  async function save(){try{prefs=await api.saveRobloxPreferences(prefs)}catch(e){note('Settings could not be saved: '+e.message,true)}}
  function renderList(id,items,{removable=false}={}){
    const el=$(id);el.replaceChildren();
    if(!items.length){const empty=document.createElement('div');empty.className='roblox-list-empty';empty.textContent=id==='robloxFavorites'?'No saved games. Paste a game link and choose Save.':'No launch requests yet.';el.append(empty);return}
    items.forEach(game=>{
      const row=document.createElement('div');row.className='saved-row';
      const name=document.createElement('strong');name.textContent=game.name;
      const meta=document.createElement('small');meta.textContent=game.placeId;
      const launch=document.createElement('button');launch.className='button primary';launch.type='button';launch.textContent='Play';launch.addEventListener('click',()=>launchGame(game.placeId));
      row.append(name,meta,launch);
      if(removable){const remove=document.createElement('button');remove.type='button';remove.className='button subtle';remove.textContent='Remove';remove.addEventListener('click',()=>{prefs.favorites=prefs.favorites.filter(f=>f.placeId!==game.placeId);renderGames();save()});row.append(remove)}
      el.append(row);
    });
  }
  function renderGames(){renderList('robloxFavorites',prefs.favorites,{removable:true});renderList('robloxRecent',prefs.recent)}
  function updatePrefsUI(){
    $('robloxFps').value=prefs.fps;$('robloxLayout').value=prefs.layout;$('robloxAccent').value=prefs.accent;
    $('robloxMode').value=prefs.launchMode;$('robloxShowDetails').checked=prefs.showDetails;
    $('robloxKeepHistory').checked=prefs.keepHistory;$('robloxConfirmLaunch').checked=prefs.confirmLaunch;
    $('robloxFpsGuide').textContent=prefs.fps==='default'?'Roblox manages its own frame rate. You can change it in Escape → Settings → Maximum Frame Rate.':`Preferred target: ${prefs.fps} FPS. In a Roblox experience, open Escape → Settings → Maximum Frame Rate and choose ${prefs.fps} FPS. LunarPatcher does not override this setting.`;
    const preview=$('robloxLayoutPreview');preview.className='roblox-layout-preview '+prefs.layout;preview.style.setProperty('--rb-accent',prefs.accent);
  }
  async function refreshStatus(){
    try{lastStatus=await api.robloxStatus();
      $('rbProtocol').textContent=lastStatus.protocol?'Registered':'Not registered';
      $('rbPlayerProtocol').textContent=lastStatus.playerProtocol?'Registered':'Not registered';
      $('rbExe').textContent=lastStatus.playerExe||'Not detected';
      $('rbReason').textContent=lastStatus.reason||'';
      $('robloxHealth').textContent=lastStatus.protocol?'Player handler ready':'Check Roblox installation';
    }catch(e){lastStatus=null;$('rbReason').textContent=e.message;$('robloxHealth').textContent='Detection failed';}
  }
  async function launchGame(value,app=false){
    if(prefs.confirmLaunch && !window.confirm(app?'Open Roblox Player?':`Open ${names[idFrom(value)]||'this experience'} in Roblox?`))return;
    try{
      const result=await api.launchRoblox({target:value,app,mode:prefs.launchMode});
      note(result.message||'Launch request accepted by Windows.');
      log((result.method||'unknown')+' → '+(result.url||'')+' ('+(result.message||'')+')');
      const id=idFrom(value);
      if(!app&&id&&prefs.keepHistory){prefs.recent=[{placeId:id,name:names[id]||prefs.favorites.find(f=>f.placeId===id)?.name||`Experience ${id}`,time:Date.now()},...prefs.recent.filter(x=>x.placeId!==id)].slice(0,16);renderGames();await save()}
    }catch(e){note('Launch failed: '+e.message,true);log('FAILED: '+e.message);}
  }
  async function saveFavorite(){const id=idFrom($('robloxTarget').value);
    if(!id){note('Paste a valid Roblox game URL or place ID before saving.',true);return}
    if(prefs.favorites.some(x=>x.placeId===id)){note('This experience is already saved.');return}
    prefs.favorites.unshift({placeId:id,name:names[id]||`Experience ${id}`});prefs.favorites=prefs.favorites.slice(0,40);renderGames();await save();note('Added to your library.');
  }
  function wire(){
    $('robloxLaunchForm').addEventListener('submit',e=>{e.preventDefault();launchGame($('robloxTarget').value)});
    $('robloxSaveGame').addEventListener('click',saveFavorite);
    $('openRobloxApp').addEventListener('click',()=>launchGame('',true));
    $('robloxOpenSettings').addEventListener('click',()=>launchGame('',true));
    $('robloxRecheck').addEventListener('click',refreshStatus);
    $('robloxTestLaunch').addEventListener('click',()=>launchGame('',true));
    $('robloxWebsite').addEventListener('click',()=>api.launchRoblox({app:true,mode:'browser'}).then(x=>note(x.message)).catch(e=>note(e.message,true)));
    document.querySelectorAll('[data-roblox-place]').forEach(el=>el.addEventListener('click',()=>launchGame(el.dataset.robloxPlace)));
    document.querySelectorAll('[data-roblox-panel]').forEach(btn=>btn.addEventListener('click',()=>{
      const name=btn.dataset.robloxPanel;
      document.querySelectorAll('[data-roblox-panel]').forEach(x=>{x.classList.toggle('active',x===btn);x.setAttribute('aria-selected',String(x===btn))});
      for(const n of ['games','performance','customize','diagnostics'])$('roblox'+n[0].toUpperCase()+n.slice(1)).hidden=name!==n;
    }));
    for(const [id,key] of [['robloxFps','fps'],['robloxLayout','layout'],['robloxMode','launchMode'],['robloxAccent','accent']]){
      $(id).addEventListener('change',()=>{prefs[key]=$(id).value;updatePrefsUI();save()});
    }
    for(const [id,key] of [['robloxShowDetails','showDetails'],['robloxKeepHistory','keepHistory'],['robloxConfirmLaunch','confirmLaunch']]){
      $(id).addEventListener('change',()=>{prefs[key]=$(id).checked;save()});
    }
  }
  async function init(){
    if(!api||!api.launchRoblox){note('Requires LunarPatcher desktop application.',true);return}
    wire();
    try {prefs=await api.robloxPreferences();updatePrefsUI();renderGames();await refreshStatus();}catch(e){note('Could not load Roblox preferences: '+e.message,true)}
  }
  document.addEventListener('DOMContentLoaded',init);
})();
