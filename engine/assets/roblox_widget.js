// Injected into Lunar's renderer. Account bookmarks contain names only; no cookies, passwords or tokens.
lunarPatcherRobloxIcon=props=>(0,__JSX__.jsx)(`svg`,{viewBox:`0 0 24 24`,width:props?.size||23,height:props?.size||23,fill:`currentColor`,"aria-hidden":!0,children:(0,__JSX__.jsx)(`path`,{fillRule:`evenodd`,d:`M7.56 1.15 22.85 7.56 16.44 22.85 1.15 16.44 7.56 1.15ZM10.23 8.45 8.45 13.77 13.77 15.55 15.55 10.23 10.23 8.45Z`})}),
lunarPatcherRobloxPage=()=>{
 const h=(type,props,...children)=>(0,__JSX__.jsx)(type,{...(props||{}),children:children.length===1?children[0]:children});
 const KEY=`lunarpatcher.roblox.`;
 const get=(key,def)=>{try{const v=JSON.parse(localStorage.getItem(KEY+key)||`null`);return v??def}catch{return def}};
 const put=(key,v)=>{try{localStorage.setItem(KEY+key,JSON.stringify(v));return true}catch{return false}};
 const validId=id=>/^[1-9]\d{0,17}$/.test(String(id||``));
 const validUser=name=>/^[A-Za-z0-9_]{3,20}$/.test(name)&&!name.startsWith(`_`)&&!name.endsWith(`_`);
 const cleanGame=g=>g&&validId(g.id)&&typeof g.name===`string`?{id:String(g.id),name:g.name.slice(0,48)||`Game ${g.id}`} :null;
 const cleanAccount=a=>a&&typeof a.username===`string`&&validUser(a.username)?{username:a.username,label:String(a.label||a.username).slice(0,36)}:null;
 const initialSettings={launchMode:`direct`,fallback:true,confirm:false,showPopular:true,keepHistory:true};
 const [tab,setTab]=(0,__REACT__.useState)(`play`);
 const [entry,setEntry]=(0,__REACT__.useState)(``);
 const [notice,setNotice]=(0,__REACT__.useState)(``);
 const [favorites,setFavorites]=(0,__REACT__.useState)(()=>{const a=get(`favorites.v1`,[]);return Array.isArray(a)?a.map(cleanGame).filter(Boolean).slice(0,32):[]});
 const [recent,setRecent]=(0,__REACT__.useState)(()=>{const a=get(`recent.v1`,[]);return Array.isArray(a)?a.map(cleanGame).filter(Boolean).slice(0,12):[]});
 const [accounts,setAccounts]=(0,__REACT__.useState)(()=>{const a=get(`accounts.v1`,[]);return Array.isArray(a)?a.map(cleanAccount).filter(Boolean).slice(0,5):[]});
 const [preferred,setPreferred]=(0,__REACT__.useState)(()=>String(get(`preferred.v1`,``)));
 const [username,setUsername]=(0,__REACT__.useState)(``);
 const [accountLabel,setAccountLabel]=(0,__REACT__.useState)(``);
 const [settings,setSettings]=(0,__REACT__.useState)(()=>{const s=get(`settings.v1`,{});return{...initialSettings,...(s&&typeof s===`object`?s:{})}});
 const [renaming,setRenaming]=(0,__REACT__.useState)(``);
 const [newGameName,setNewGameName]=(0,__REACT__.useState)(``);
 const popular=[{id:`4924922222`,name:`Brookhaven RP`,type:`Roleplay`},{id:`2753915549`,name:`Blox Fruits`,type:`Adventure`},{id:`6516141723`,name:`DOORS`,type:`Horror`},{id:`920587237`,name:`Adopt Me!`,type:`Roleplay`}];
 const persist=(key,data,setter)=>{setter(data);if(!put(key,data))setNotice(`Could not save locally. Check launcher storage permissions.`)};
 const updateSetting=(name,value)=>{const next={...settings,[name]:value};persist(`settings.v1`,next,setSettings)};
 const open=url=>Promise.resolve(window.lunar.misc.openExternalLink({url,initiator:__INITIATOR__}));
 const openWebsite=(id)=>open(`https://www.roblox.com/games/${id}/`);
 const titleFor=id=>favorites.find(g=>g.id===id)?.name||recent.find(g=>g.id===id)?.name||popular.find(g=>g.id===id)?.name||`Game ${id}`;
 const parseTarget=value=>{
  const raw=String(value||``).trim();
  if(validId(raw))return{kind:`game`,id:raw};
  if(/^roblox:\/\/experiences\/start\?placeId=([1-9]\d{0,17})$/i.test(raw))return{kind:`game`,id:raw.split(`=`)[1]};
  try{
   const parsed=new URL(raw);
   if(parsed.protocol!==`https:`||!([`roblox.com`,`www.roblox.com`].includes(parsed.hostname.toLowerCase())))return null;
   const match=parsed.pathname.match(/^\/games\/([1-9]\d{0,17})(?:\/[^?#]*)?\/?$/i);
   if(match)return{kind:`game`,id:match[1]};
   if(parsed.pathname===`/share`&&parsed.searchParams.get(`type`)===`Server`){
    const code=parsed.searchParams.get(`code`);
    if(code&&/^[a-zA-Z0-9_+\/-]{3,128}$/.test(code))return{kind:`server`,url:`https://www.roblox.com/share?code=${encodeURIComponent(code)}&type=Server`};
   }
  }catch{}
  return null;
 };
 const remember=id=>{if(!settings.keepHistory)return;const next=[{id,name:titleFor(id)},...recent.filter(g=>g.id!==id)].slice(0,12);persist(`recent.v1`,next,setRecent)};
 const launch=(target)=>{
  if(!target){setNotice(`Enter a valid Roblox game, place ID or private-server invite.`);return}
  if(target.kind===`server`){setNotice(`Opening the private server invitation in Roblox.`);open(target.url).catch(()=>setNotice(`Could not open the server link.`));return}
  const id=target.id;
  if(settings.confirm&&typeof window.confirm===`function`&&!window.confirm(`Launch ${titleFor(id)} in Roblox?`))return;
  remember(id);
  if(settings.launchMode===`website`){setNotice(`Opening the experience on Roblox.`);openWebsite(id).catch(()=>setNotice(`Could not open Roblox in your browser.`));return}
  setNotice(`Requested Roblox to open ${titleFor(id)}. The official Roblox Player handles launch.`);
  try{open(`roblox://experiences/start?placeId=${id}`).then(result=>{if(result===false&&settings.fallback)openWebsite(id)}).catch(()=>{if(settings.fallback)openWebsite(id);else setNotice(`Roblox Player did not accept the launch request.`)})}
  catch{if(settings.fallback)openWebsite(id);else setNotice(`Unable to open the Roblox link.`)}
 };
 const saveGame=()=>{
  const target=parseTarget(entry);
  if(!target){setNotice(`Enter a valid Roblox experience first.`);return}
  if(target.kind!==`game`){setNotice(`Private server invites are not saved because they may grant access to private sessions.`);return}
  if(favorites.some(g=>g.id===target.id)){setNotice(`This game is already saved.`);return}
  persist(`favorites.v1`,[...favorites,{id:target.id,name:titleFor(target.id)}].slice(-32),setFavorites);setNotice(`Shortcut saved.`)
 };
 const removeGame=id=>{persist(`favorites.v1`,favorites.filter(g=>g.id!==id),setFavorites);if(renaming===id)setRenaming(``)};
 const saveRename=id=>{
  const label=newGameName.trim().slice(0,48);
  if(!label){setNotice(`Enter a name first.`);return}
  persist(`favorites.v1`,favorites.map(g=>g.id===id?{...g,name:label}:g),setFavorites);setRenaming(``);setNotice(`Shortcut renamed.`)
 };
 const addAccount=()=>{
  const name=username.trim();
  if(!validUser(name)){setNotice(`Enter a valid Roblox username (3–20 letters, numbers or underscores).`);return}
  if(accounts.some(a=>a.username.toLowerCase()===name.toLowerCase())){setNotice(`This profile is already saved.`);return}
  if(accounts.length>=5){setNotice(`The list supports up to 5 profiles, like Roblox's official account switcher.`);return}
  const next=[...accounts,{username:name,label:accountLabel.trim().slice(0,36)||name}];persist(`accounts.v1`,next,setAccounts);setUsername(``);setAccountLabel(``);setNotice(`Profile bookmarked. Sign-in remains on the official Roblox website.`)
 };
 const removeAccount=name=>{persist(`accounts.v1`,accounts.filter(a=>a.username!==name),setAccounts);if(preferred===name)persist(`preferred.v1`,``,setPreferred)};
 const accountSwitcher=()=>{setNotice(`On Roblox Home, open Settings (gear) → Switch Account. The bookmark here does not change your login.`);open(`https://www.roblox.com/home`).catch(()=>setNotice(`Could not open the official account switcher page.`))};
 const card=(g,saved)=>h(`div`,{className:`lunar-patcher-roblox-card`,key:g.id},
   h(`div`,{className:`lunar-patcher-roblox-tile`},h(lunarPatcherRobloxIcon,{size:22})),
   h(`div`,{className:`lunar-patcher-roblox-card-copy`},saved&&renaming===g.id?
    h(`div`,{className:`lunar-patcher-roblox-inline`},h(`input`,{type:`text`,value:newGameName,maxLength:48,onChange:e=>setNewGameName(e.target.value),onKeyDown:e=>e.key===`Enter`&&saveRename(g.id),"aria-label":`New game name`}),h(`button`,{className:`lunar-patcher-roblox-textbtn`,onClick:()=>saveRename(g.id)},`Save`)):
    h(`strong`,{},g.name),h(`span`,{},g.type||`Experience · ${g.id}`)),
   saved?h(`button`,{className:`lunar-patcher-roblox-iconbtn`,title:`Rename`,"aria-label":`Rename ${g.name}`,onClick:()=>{setRenaming(g.id);setNewGameName(g.name)}},`Edit`):null,
   saved?h(`button`,{className:`lunar-patcher-roblox-iconbtn`,title:`Remove`,"aria-label":`Remove ${g.name}`,onClick:()=>removeGame(g.id)},`Remove`):null,
   h(`button`,{className:`lunar-patcher-roblox-secondary`,onClick:()=>openWebsite(g.id).catch(()=>setNotice(`Could not open Roblox website.`))},`View`),
   h(`button`,{className:`lunar-patcher-roblox-play`,onClick:()=>launch({kind:`game`,id:g.id})},`Play`));
 const section=(title,body,action)=>h(`section`,{className:`lunar-patcher-roblox-section`},h(`div`,{className:`lunar-patcher-roblox-heading`},h(`h2`,{},title),action||null),body);
 const settingRow=(title,detail,checked,onChange)=>h(`label`,{className:`lunar-patcher-roblox-setting`},h(`span`,{},h(`strong`,{},title),h(`small`,{},detail)),h(`input`,{type:`checkbox`,checked,onChange:e=>onChange(e.target.checked)}));
 const button=(label,onClick,kind=`secondary`)=>h(`button`,{type:`button`,className:`lunar-patcher-roblox-${kind}`,onClick},label);
 const gamesSections=[
  section(`Launch a game`, h(`div`,{},
   h(`div`,{className:`lunar-patcher-roblox-inputrow`},
    h(`input`,{type:`text`,value:entry,onChange:e=>setEntry(e.target.value.slice(0,300)),onKeyDown:e=>e.key===`Enter`&&launch(parseTarget(entry)),placeholder:`Game URL, place ID or private server invite`,"aria-label":`Roblox game URL or place ID`}),
    button(`Play`,()=>launch(parseTarget(entry)),`play`),button(`Save`,saveGame)),
   h(`p`,{className:`lunar-patcher-roblox-hint`},`Uses the installed Roblox Player. Private server invitations open on Roblox's website.`)))
 ];
 if(favorites.length)gamesSections.push(section(`Saved games`,h(`div`,{className:`lunar-patcher-roblox-grid`},...favorites.map(g=>card(g,true))),button(`Clear`,()=>persist(`favorites.v1`,[],setFavorites))));
 if(recent.length)gamesSections.push(section(`Recently launched`,h(`div`,{className:`lunar-patcher-roblox-grid`},...recent.map(g=>card(g,false))),button(`Clear`,()=>persist(`recent.v1`,[],setRecent))));
 if(settings.showPopular)gamesSections.push(section(`Discover`,h(`div`,{className:`lunar-patcher-roblox-grid`},...popular.map(g=>card(g,false))),button(`Browse Roblox`,()=>open(`https://www.roblox.com/discover`))));
 const gamesPage=h(`div`,{className:`lunar-patcher-roblox-body`},...gamesSections);
 const accountForm=h(`div`,{},
  h(`p`,{className:`lunar-patcher-roblox-hint`},`Organize up to five usernames here. Changing the active login happens through Roblox's official website. No passwords or authentication cookies are saved.`),
  h(`div`,{className:`lunar-patcher-roblox-inputrow`},
   h(`input`,{type:`text`,value:username,maxLength:20,onChange:e=>setUsername(e.target.value),placeholder:`Roblox username`,"aria-label":`Roblox username`}),
   h(`input`,{type:`text`,value:accountLabel,maxLength:36,onChange:e=>setAccountLabel(e.target.value),placeholder:`Optional label`,"aria-label":`Profile label`}),
   button(`Add profile`,addAccount,`play`)));
 const accountSections=[section(`Account profiles`,accountForm,button(`Open Roblox account switcher`,accountSwitcher))];
 if(accounts.length){
  const accountCards=accounts.map(a=>h(`div`,{className:`lunar-patcher-roblox-account`,key:a.username},
   h(`span`,{className:`lunar-patcher-roblox-avatar`},a.username[0].toUpperCase()),
   h(`div`,{className:`lunar-patcher-roblox-account-copy`},h(`strong`,{},a.label),h(`span`,{},`@${a.username}${preferred===a.username?` · Preferred label`:``}`)),
   button(preferred===a.username?`Selected`:`Prefer`,()=>persist(`preferred.v1`,a.username,setPreferred)),
   button(`Switch on Roblox`,accountSwitcher),button(`Remove`,()=>removeAccount(a.username))));
  accountSections.push(section(`Saved profiles`,h(`div`,{className:`lunar-patcher-roblox-account-list`},...accountCards)));
 }
 accountSections.push(section(`Account switching`,h(`p`,{className:`lunar-patcher-roblox-hint`},`Roblox's Account Switcher supports up to five signed-in accounts on this device for eligible users. Bookmarks in Lunar do not auto-login, modify sessions or launch under a selected account. Switch on Roblox before launching a game.`)));
 const accountsPage=h(`div`,{className:`lunar-patcher-roblox-body`},...accountSections);
 const launchSettings=h(`div`,{},
  h(`label`,{className:`lunar-patcher-roblox-setting`},h(`span`,{},h(`strong`,{},`Launch method`),h(`small`,{},`Choose how games open.`)),h(`select`,{value:settings.launchMode,onChange:e=>updateSetting(`launchMode`,e.target.value)},h(`option`,{value:`direct`},`Open in Roblox Player`),h(`option`,{value:`website`},`Open Roblox website`))),
  settingRow(`Fall back to website`,`Open the game page if the direct launch is rejected.`,!!settings.fallback,v=>updateSetting(`fallback`,v)),
  settingRow(`Confirm before launching`,`Ask before sending a game launch request.`,!!settings.confirm,v=>updateSetting(`confirm`,v)));
 const librarySettings=h(`div`,{},
  settingRow(`Show recommended games`,`Display a small set of popular experiences.`,!!settings.showPopular,v=>updateSetting(`showPopular`,v)),
  settingRow(`Remember recent games`,`Store recent place IDs locally in this launcher.`,!!settings.keepHistory,v=>updateSetting(`keepHistory`,v)));
 const robloxSettings=h(`div`,{},
  h(`p`,{className:`lunar-patcher-roblox-hint`},`Graphics, volume, camera and frame-rate settings are managed inside Roblox. LunarPatcher does not modify Roblox client files or unsupported FastFlags.`),
  h(`div`,{className:`lunar-patcher-roblox-actions`},
   button(`Account settings`,()=>open(`https://www.roblox.com/my/account`)),
   button(`Install or update Roblox`,()=>open(`https://www.roblox.com/download`)),
   button(`Roblox support`,()=>open(`https://en.help.roblox.com/`))));
 const settingsPage=h(`div`,{className:`lunar-patcher-roblox-body`},
  section(`Launch behavior`,launchSettings),section(`Library`,librarySettings),section(`Roblox controls`,robloxSettings));
 const header=h(`header`,{className:`lunar-patcher-roblox-header`},
  h(`div`,{},h(`div`,{className:`lunar-patcher-roblox-eyebrow`},`Lunar Client · Roblox`),h(`h1`,{},`Roblox`),h(`p`,{},`Games, accounts and launch preferences in one place.`)),
  button(`Roblox website`,()=>open(`https://www.roblox.com/home`)));
 const tabs=h(`nav`,{className:`lunar-patcher-roblox-tabs`,role:`tablist`},
  ...[[`play`,`Games`],[`accounts`,`Accounts`],[`settings`,`Settings`]].map(([id,label])=>h(`button`,{type:`button`,role:`tab`,"aria-selected":tab===id,className:tab===id?`is-active`:``,key:id,onClick:()=>{setTab(id);setNotice(``)}},label)));
 return h(`main`,{className:`lunar-patcher-roblox`},header,tabs,
  h(`div`,{className:`lunar-patcher-roblox-scroll`},tab===`accounts`?accountsPage:tab===`settings`?settingsPage:gamesPage),
  h(`div`,{className:`lunar-patcher-roblox-notice`,role:`status`},notice||`Uses Roblox's official client and website. LunarPatcher never asks for passwords or cookies.`));
},
