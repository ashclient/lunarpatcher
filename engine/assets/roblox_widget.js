// LunarPatcher Roblox Hub. Native UI only. No Roblox credentials, cookies or token handling.
lunarPatcherRobloxIcon=props=>(0,__JSX__.jsx)(`svg`,{viewBox:`0 0 24 24`,width:props?.size||22,height:props?.size||22,fill:`currentColor`,"aria-hidden":!0,children:(0,__JSX__.jsx)(`path`,{fillRule:`evenodd`,d:`M7.56 1.15 22.85 7.56 16.44 22.85 1.15 16.44 7.56 1.15ZM10.23 8.45 8.45 13.77 13.77 15.55 15.55 10.23 10.23 8.45Z`})}),
lunarPatcherRobloxPage=()=>{
 // React void elements MUST NOT receive children, even an empty children array.
 const h=(type,props,...children)=>{const p={...(props||{})};if(children.length===1)p.children=children[0];else if(children.length>1)p.children=children;return (0,__JSX__.jsx)(type,p)};
 const KEY=`lunarpatcher.roblox.`;
 const get=(key,def)=>{try{const v=JSON.parse(localStorage.getItem(KEY+key)||`null`);return v??def}catch{return def}};
 const put=(key,v)=>{try{localStorage.setItem(KEY+key,JSON.stringify(v));return true}catch{return false}};
 const validId=id=>/^[1-9]\d{0,17}$/.test(String(id||``));
 const validUser=name=>/^[A-Za-z0-9_]{3,20}$/.test(name)&&!name.startsWith(`_`)&&!name.endsWith(`_`);
 const cleanGame=g=>g&&validId(g.id)&&typeof g.name===`string`?{id:String(g.id),name:g.name.slice(0,60)||`Experience ${g.id}`,genre:String(g.genre||`Experience`).slice(0,24)}:null;
 const cleanAccount=a=>a&&typeof a.username===`string`&&validUser(a.username)?{username:a.username,label:String(a.label||a.username).slice(0,36)}:null;
 const ART=__ROBLOX_ART__; // Locally cached, official Roblox game icons (when available)
 const imageFor=id=>ART[String(id)]||`https://www.roblox.com/asset-thumbnail/image?assetId=${encodeURIComponent(id)}&width=420&height=420&format=png`;
 const CATALOG=[
  {id:`4924922222`,name:`Brookhaven RP`,genre:`Roleplay`,subtitle:`Your own world, your own story`,art:`brookhaven`},
  {id:`2753915549`,name:`Blox Fruits`,genre:`Adventure`,subtitle:`Explore an ocean of adventure`,art:`blox`},
  {id:`6516141723`,name:`DOORS`,genre:`Horror`,subtitle:`Every door hides something`,art:`doors`},
  {id:`920587237`,name:`Adopt Me!`,genre:`Roleplay`,subtitle:`Collect, raise, and trade pets`,art:`adopt`},
  {id:`142823291`,name:`Murder Mystery 2`,genre:`Mystery`,subtitle:`Trust nobody`,art:`mystery`},
  {id:`1962086868`,name:`Tower of Hell`,genre:`Obby`,subtitle:`Race to the top`,art:`tower`},
  {id:`189707`,name:`Natural Disaster Survival`,genre:`Survival`,subtitle:`Survive the unexpected`,art:`disaster`},
  {id:`17625359962`,name:`RIVALS`,genre:`Action`,subtitle:`Fast-paced competition`,art:`rivals`}
 ];
 const DEFAULT={launchMode:`desktop`,fallback:true,confirm:false,showPopular:true,keepHistory:true,hubLayout:`comfortable`,hubColor:`#E5A1BB`,fpsTarget:`default`,reduceMotion:false};
 const [page,setPage]=(0,__REACT__.useState)(`home`);
 const [query,setQuery]=(0,__REACT__.useState)(``);
 const [category,setCategory]=(0,__REACT__.useState)(`All`);
 const [entry,setEntry]=(0,__REACT__.useState)(``);
 const [notice,setNotice]=(0,__REACT__.useState)(``);
 const [detail,setDetail]=(0,__REACT__.useState)(null);
 const [favorites,setFavorites]=(0,__REACT__.useState)(()=>{const v=get(`favorites.v1`,[]);return Array.isArray(v)?v.map(cleanGame).filter(Boolean).slice(0,48):[]});
 const [recent,setRecent]=(0,__REACT__.useState)(()=>{const v=get(`recent.v1`,[]);return Array.isArray(v)?v.map(cleanGame).filter(Boolean).slice(0,16):[]});
 const [accounts,setAccounts]=(0,__REACT__.useState)(()=>{const v=get(`accounts.v1`,[]);return Array.isArray(v)?v.map(cleanAccount).filter(Boolean).slice(0,5):[]});
 const [preferred,setPreferred]=(0,__REACT__.useState)(()=>String(get(`preferred.v1`,``)));
 const [username,setUsername]=(0,__REACT__.useState)(``);
 const [accountLabel,setAccountLabel]=(0,__REACT__.useState)(``);
 const [settings,setSettings]=(0,__REACT__.useState)(()=>{const v=get(`settings.v1`,{});return{...DEFAULT,...(v&&typeof v===`object`&&!Array.isArray(v)?v:{})}});
 const [renameId,setRenameId]=(0,__REACT__.useState)(``);
 const [renameText,setRenameText]=(0,__REACT__.useState)(``);
 const persist=(key,data,set)=>{set(data);if(!put(key,data))setNotice(`Could not save this change locally.`)};
 const updateSetting=(key,value)=>persist(`settings.v1`,{...settings,[key]:value},setSettings);
 const navTo=id=>{setDetail(null);setPage(id);setNotice(``)};
 const external=url=>Promise.resolve(window.lunar.misc.openExternalLink({url,initiator:__INITIATOR__}));
 const browse=url=>external(url).catch(()=>setNotice(`Couldn't open Roblox. Try opening Roblox in your browser.`));
 const gameUrl=id=>`https://www.roblox.com/games/${id}`;
 const titleFor=id=>favorites.find(g=>g.id===id)?.name||recent.find(g=>g.id===id)?.name||CATALOG.find(g=>g.id===id)?.name||`Experience ${id}`;
 const parseTarget=value=>{
  const raw=String(value||``).trim();if(validId(raw))return{kind:`game`,id:raw};
  try{const u=new URL(raw);
   if(u.protocol===`roblox:`&&u.hostname.toLowerCase()===`experiences`&&u.pathname===`/start`){const id=u.searchParams.get(`placeId`);return validId(id)?{kind:`game`,id}:null}
   if(u.protocol!==`https:`||!([`roblox.com`,`www.roblox.com`].includes(u.hostname.toLowerCase())))return null;
   const m=u.pathname.match(/^\/games\/([1-9]\d{0,17})(?:\/[^?#]*)?\/?$/i);if(m)return{kind:`game`,id:m[1]};
   if(u.pathname===`/share`&&u.searchParams.get(`type`)===`Server`){const code=u.searchParams.get(`code`);if(code&&/^[A-Za-z0-9_+\/-]{3,128}$/.test(code))return{kind:`server`,url:`https://www.roblox.com/share?code=${encodeURIComponent(code)}&type=Server`}}
  }catch{}return null;
 };
 const launch=target=>{
  if(!target){setNotice(`Enter a Roblox game link or place ID.`);return}
  if(target.kind===`server`){setNotice(`Opening the official Roblox private-server invitation.`);browse(target.url);return}
  const id=String(target.id);if(!validId(id)){setNotice(`Invalid place ID.`);return}
  if(settings.confirm&&typeof window.confirm===`function`&&!window.confirm(`Launch ${titleFor(id)} in Roblox?`))return;
  if(settings.keepHistory)persist(`recent.v1`,[{id,name:titleFor(id)},...recent.filter(g=>g.id!==id)].slice(0,16),setRecent);
  if(settings.launchMode===`website`){setNotice(`Opening game page on Roblox.`);browse(gameUrl(id));return}
  setNotice(`Opening ${titleFor(id)} in Roblox Player…`);
  const desktop=`lunarpatcher://launch?placeId=${id}`;
  const direct=`roblox://experiences/start?placeId=${id}`;
  // Lunar's external-link bridge can reject custom schemes. Try the next supported
  // handler only on explicit failure; never fire two accepted launches at once.
  const urls=settings.launchMode===`desktop`?[desktop,direct]:[direct];
  const tryLaunch=async index=>{
   if(index>=urls.length){if(settings.fallback)return browse(gameUrl(id));setNotice(`Could not hand the request to Windows. Open Roblox through your browser.`);return}
   try{const result=await external(urls[index]);if(result===false)throw Error(`Link rejected`);
    setNotice(`Launch link sent to Windows. If Roblox didn't open, try the Roblox website or the LunarPatcher desktop app.`);
   }catch(e){return tryLaunch(index+1)}
  };
  void tryLaunch(0);
 };
 const saveGame=g=>{
  if(!g||!validId(g.id))return;
  if(favorites.some(x=>x.id===String(g.id))){setNotice(`Already saved.`);return}
  persist(`favorites.v1`,[{id:String(g.id),name:g.name||titleFor(g.id),genre:g.genre||`Experience`},...favorites].slice(0,48),setFavorites);setNotice(`Saved to your library.`);
 };
 const saveFromInput=()=>{const target=parseTarget(entry);if(!target){setNotice(`Paste a valid Roblox URL or place ID.`);return}if(target.kind===`server`){setNotice(`Private-server invites are opened, not saved.`);return}saveGame({id:target.id,name:titleFor(target.id)})};
 const openGame=g=>{setDetail({...g,id:String(g.id),name:g.name||titleFor(g.id)})};
 const games=[...CATALOG];
 const genres=[`All`,`Roleplay`,`Adventure`,`Action`,`Horror`,`Obby`,`Survival`,`Mystery`];
 const needle=query.trim().toLowerCase();
 const visible=games.filter(g=>(category===`All`||g.genre===category)&&(!needle||`${g.name} ${g.genre} ${g.id}`.toLowerCase().includes(needle)));
 const btn=(label,onClick,variant=`secondary`,extra={})=>h(`button`,{type:`button`,className:`lpr-btn lpr-${variant}`,onClick,...extra},label);
 const icon=(name)=>({home:`⌂`,discover:`▦`,library:`▤`,friends:`◌`,avatar:`◉`,marketplace:`◇`,accounts:`♙`,settings:`⚙`})[name]||`•`;
 const tile=(g,compact=false)=>h(`article`,{className:`lpr-game ${compact?`lpr-game-small`:``}`,key:g.id},
   h(`button`,{type:`button`,className:`lpr-game-art lpr-art-${g.art||`custom`}`,onClick:()=>openGame(g),"aria-label":`View ${g.name}`},
     h(`span`,{className:`lpr-game-glyph`},(g.name||`R`).slice(0,1).toUpperCase()),
     h(`img`,{loading:`lazy`,alt:`${g.name} game artwork`,src:imageFor(g.id),onError:e=>{e.currentTarget.style.display=`none`}})),
   h(`div`,{className:`lpr-game-meta`},h(`button`,{type:`button`,className:`lpr-game-name`,onClick:()=>openGame(g)},g.name),h(`small`,{},g.genre||`Experience`)),
   h(`button`,{type:`button`,className:`lpr-game-quick`,onClick:()=>launch({kind:`game`,id:g.id}),"aria-label":`Play ${g.name}`},`▶`));
 const rail=(title,items,action)=>h(`section`,{className:`lpr-rail`},h(`div`,{className:`lpr-rail-head`},h(`h2`,{},title),action||null),items.length?h(`div`,{className:`lpr-game-grid`},...items.map(g=>tile(g))):h(`div`,{className:`lpr-empty`},`Nothing here yet. Browse experiences to add some.`));
 const catalogSaved=id=>favorites.some(g=>g.id===id);
 const primary=CATALOG[1];
 const launchRow=h(`form`,{className:`lpr-launch-row`,onSubmit:e=>{e.preventDefault();launch(parseTarget(entry))}},
  h(`input`,{type:`text`,value:entry,onChange:e=>setEntry(e.target.value.slice(0,350)),placeholder:`Paste a Roblox game link or place ID`,"aria-label":`Roblox game link or place ID`}),
  btn(`Play`,()=>launch(parseTarget(entry)),`primary`),btn(`Save`,saveFromInput));
 const home=h(`div`,{className:`lpr-content`},
   h(`div`,{className:`lpr-welcome`},h(`div`,{},h(`div`,{className:`lpr-eyebrow`},`LUNAR / ROBLOX`),h(`h1`,{},`Home`),h(`p`,{},`Your games, right where you left off.`)),btn(`Open Roblox`,()=>browse(`https://www.roblox.com/home`))),
   h(`section`,{className:`lpr-feature lpr-art-blox`},h(`img`,{className:`lpr-feature-image`,src:imageFor(primary.id),alt:``,onError:e=>{e.currentTarget.style.display=`none`}}),h(`div`,{className:`lpr-feature-copy`},h(`span`,{className:`lpr-feature-label`},`FEATURED EXPERIENCE`),h(`h2`,{},primary.name),h(`p`,{},primary.subtitle),h(`div`,{className:`lpr-button-row`},btn(`▶  Play`,()=>launch({kind:`game`,id:primary.id}),`primary`),btn(`Game details`,()=>openGame(primary),`glass`)))),
   recent.length?rail(`Continue playing`,recent.slice(0,6),btn(`View history`,()=>navTo(`library`),`text`)):rail(`Recently popular`,CATALOG.slice(0,5),btn(`See all`,()=>navTo(`discover`),`text`)),
   favorites.length?rail(`Your saved experiences`,favorites.slice(0,6),btn(`Library`,()=>navTo(`library`),`text`)):null,
   settings.showPopular?rail(`Explore something new`,CATALOG.slice(3,8),btn(`Explore`,()=>navTo(`discover`),`text`)):null,
   h(`div`,{className:`lpr-inline-launch`},h(`strong`,{},`Have a game link?`),launchRow));
 const discover=h(`div`,{className:`lpr-content`},
   h(`div`,{className:`lpr-heading`},h(`div`,{},h(`h1`,{},`Discover`),h(`p`,{},`Browse a starter catalog, or explore everything on Roblox.`)),btn(`All Roblox charts ↗`,()=>browse(`https://www.roblox.com/charts`))),
   h(`div`,{className:`lpr-chips`},...genres.map(g=>btn(g,()=>setCategory(g),category===g?`chip-active`:`chip`,{"aria-pressed":category===g,key:g}))),
   rail(`${visible.length} experiences`,visible,btn(`Browse on Roblox`,()=>browse(`https://www.roblox.com/charts`),`text`)),
   needle&&visible.length===0?h(`div`,{className:`lpr-empty`},`No matches in the local catalog. Use Roblox search to browse more games.`,btn(`Search on Roblox ↗`,()=>browse(`https://www.roblox.com/charts`),`primary`)):null);
 const library=h(`div`,{className:`lpr-content`},h(`div`,{className:`lpr-heading`},h(`div`,{},h(`h1`,{},`Library`),h(`p`,{},`Saved and recently played experiences are stored locally.`))),
  h(`div`,{className:`lpr-inline-launch`},h(`strong`,{},`Add a game`),launchRow),
  rail(`Saved experiences`,favorites, favorites.length?btn(`Clear saved`,()=>{persist(`favorites.v1`,[],setFavorites);setNotice(`Saved games cleared.`)},`text`):null),
  rail(`Recently played`,recent,recent.length?btn(`Clear history`,()=>{persist(`recent.v1`,[],setRecent);setNotice(`Recent games cleared.`)},`text`):null));
 const gate=(title,description,links)=>h(`div`,{className:`lpr-content`},h(`div`,{className:`lpr-heading`},h(`div`,{},h(`h1`,{},title),h(`p`,{},description))),
   h(`section`,{className:`lpr-account-gate`},h(`div`,{className:`lpr-gate-mark`},title===`Avatar`?`◉`:title===`Friends`?`◌`:`◇`),h(`h2`,{},`Continue on Roblox`),
    h(`p`,{},`This area needs your signed-in Roblox account. The official site handles your session; LunarPatcher does not collect credentials or cookies.`),
    h(`div`,{className:`lpr-button-row`},...links.map(x=>btn(x[0],()=>browse(x[1]),x[2]||`secondary`)))));
 const friends=gate(`Friends`,`Manage your friends and connections on Roblox.`,[[`Open friends ↗`,`https://www.roblox.com/users/friends`,`primary`],[`Find people ↗`,`https://www.roblox.com/search/users`]]);
 const avatar=gate(`Avatar`,`Customize your character and manage outfits.`,[[`Open Avatar Editor ↗`,`https://www.roblox.com/my/avatar`,`primary`],[`Browse marketplace ↗`,`https://www.roblox.com/catalog`]]);
 const marketplace=gate(`Marketplace`,`Browse clothing, accessories, and avatar items.`,[[`Browse Marketplace ↗`,`https://www.roblox.com/catalog`,`primary`],[`Creator Store ↗`,`https://create.roblox.com/store/models`]]);
 const addAccount=()=>{const name=username.trim();if(!validUser(name)){setNotice(`Enter a valid Roblox username (3–20 letters, numbers or underscores).`);return}if(accounts.some(a=>a.username.toLowerCase()===name.toLowerCase())){setNotice(`Already saved.`);return}if(accounts.length>=5){setNotice(`Up to five account bookmarks are supported.`);return}persist(`accounts.v1`,[...accounts,{username:name,label:accountLabel.trim().slice(0,36)||name}],setAccounts);setUsername(``);setAccountLabel(``);setNotice(`Account bookmark saved. Switch the active Roblox login on the official website.`)};
 const accountSwitcher=()=>{setNotice(`On Roblox Home, use the gear menu to switch accounts. Bookmarks don't change the active sign-in.`);browse(`https://www.roblox.com/home`)};
 const accountsPage=h(`div`,{className:`lpr-content`},h(`div`,{className:`lpr-heading`},h(`div`,{},h(`h1`,{},`Accounts`),h(`p`,{},`Bookmarks for quick reference. Sign-in stays with Roblox.`)),btn(`Official account switcher ↗`,accountSwitcher)),
  h(`section`,{className:`lpr-panel`},h(`h2`,{},`Add a profile`),h(`p`,{},`Store a username and optional label. No passwords, tokens or sessions are saved.`),
   h(`div`,{className:`lpr-account-form`},h(`input`,{type:`text`,value:username,maxLength:20,onChange:e=>setUsername(e.target.value),placeholder:`Roblox username`,"aria-label":`Roblox username`}),h(`input`,{type:`text`,value:accountLabel,maxLength:36,onChange:e=>setAccountLabel(e.target.value),placeholder:`Optional label`,"aria-label":`Profile label`}),btn(`Add`,addAccount,`primary`))),
  h(`section`,{className:`lpr-panel`},h(`h2`,{},`Saved profiles`),accounts.length?
   h(`div`,{className:`lpr-profile-list`},...accounts.map(a=>h(`div`,{className:`lpr-profile`,key:a.username},h(`div`,{className:`lpr-profile-avatar`},a.username[0].toUpperCase()),
    h(`div`,{className:`lpr-profile-name`},h(`strong`,{},a.label),h(`small`,{},`@${a.username}`)),
    preferred===a.username?h(`span`,{className:`lpr-pill`},`Preferred`):null,
    btn(preferred===a.username?`Selected`:`Prefer`,()=>persist(`preferred.v1`,a.username,setPreferred)),btn(`Switch`,accountSwitcher),btn(`Remove`,()=>{persist(`accounts.v1`,accounts.filter(x=>x.username!==a.username),setAccounts);if(preferred===a.username)persist(`preferred.v1`,``,setPreferred)},`text`)))):
   h(`p`,{className:`lpr-muted`},`No saved profiles yet.`)));
 const switchRow=(title,desc,value,set)=>h(`label`,{className:`lpr-setting`},h(`span`,{},h(`strong`,{},title),h(`small`,{},desc)),h(`input`,{type:`checkbox`,checked:value,onChange:e=>set(e.target.checked)}));
 const settingsPage=h(`div`,{className:`lpr-content`},h(`div`,{className:`lpr-heading`},h(`div`,{},h(`h1`,{},`Settings`),h(`p`,{},`Roblox launch preferences and local storage.`))),
  h(`section`,{className:`lpr-panel`},h(`h2`,{},`Launching`),h(`label`,{className:`lpr-setting`},h(`span`,{},h(`strong`,{},`Launch method`),h(`small`,{},`Use the installed Roblox Player or view the game on Roblox.com.`)),h(`select`,{value:settings.launchMode,onChange:e=>updateSetting(`launchMode`,e.target.value)},h(`option`,{value:`desktop`},`LunarPatcher desktop handler (recommended)`),h(`option`,{value:`direct`},`Roblox protocol handler`),h(`option`,{value:`website`},`Roblox website`))),
   switchRow(`Fallback to browser`,`Open the game website if direct launching fails.`,!!settings.fallback,v=>updateSetting(`fallback`,v)),
   switchRow(`Confirm before playing`,`Show a confirmation prompt first.`,!!settings.confirm,v=>updateSetting(`confirm`,v))),
  h(`section`,{className:`lpr-panel`},h(`h2`,{},`Performance`),
   h(`p`,{},`Preferred FPS is a reminder only. Apply it inside Roblox: Escape → Settings → Maximum Frame Rate.`),
   h(`label`,{className:`lpr-setting`},h(`span`,{},h(`strong`,{},`Preferred maximum FPS`),h(`small`,{},`Roblox in-game setting; no unsupported FPS unlocker.`)),
    h(`select`,{value:settings.fpsTarget||`default`,onChange:e=>updateSetting(`fpsTarget`,e.target.value)},
     ...[`default`,`60`,`120`,`144`,`240`].map(v=>h(`option`,{value:v,key:v},v===`default`?`Roblox default`:`${v} FPS`)))),
   h(`p`,{},`To display Roblox performance statistics, press Shift + F5 while playing. Use Roblox's Graphics Mode and Fullscreen controls for graphical tuning.`)),
  h(`section`,{className:`lpr-panel`},h(`h2`,{},`Hub appearance`),
   h(`p`,{},`These settings customize the Roblox page inside Lunar, not the game's protected in-game GUI.`),
   h(`label`,{className:`lpr-setting`},h(`span`,{},h(`strong`,{},`Layout density`),h(`small`,{},`Control spacing and card sizes.`)),
    h(`select`,{value:settings.hubLayout||`comfortable`,onChange:e=>updateSetting(`hubLayout`,e.target.value)},
     ...[`compact`,`comfortable`,`spacious`].map(v=>h(`option`,{value:v,key:v},v[0].toUpperCase()+v.slice(1))))),
   h(`label`,{className:`lpr-setting`},h(`span`,{},h(`strong`,{},`Hub accent color`),h(`small`,{},`Independent of LunarPatcher's Sakura theme.`)),
    h(`input`,{type:`color`,value:/^#[0-9a-f]{6}$/i.test(settings.hubColor||``)?settings.hubColor:`#E5A1BB`,onChange:e=>updateSetting(`hubColor`,e.target.value)})),
   switchRow(`Reduce animations`,`Use more subtle transitions in the Roblox hub.`,!!settings.reduceMotion,v=>updateSetting(`reduceMotion`,v))),
  h(`section`,{className:`lpr-panel`},h(`h2`,{},`Library`),switchRow(`Show suggested experiences`,`Display a curated selection on Home.`,!!settings.showPopular,v=>updateSetting(`showPopular`,v)),switchRow(`Save recent games`,`Only place IDs and names are stored locally.`,!!settings.keepHistory,v=>updateSetting(`keepHistory`,v))),
  h(`section`,{className:`lpr-panel`},h(`h2`,{},`Roblox Player`),h(`p`,{},`Roblox supports an in-game maximum FPS setting up to 240 on Windows. This hub doesn't bypass Roblox's restrictions or modify Roblox client files.`),
   h(`div`,{className:`lpr-button-row`},btn(`Install Roblox ↗`,()=>browse(`https://www.roblox.com/download`)),btn(`Account settings ↗`,()=>browse(`https://www.roblox.com/my/account`)),btn(`Support ↗`,()=>browse(`https://en.help.roblox.com/`)))));
 const activeGame=detail&&validId(detail.id)?detail:null;
 const details=activeGame?h(`div`,{className:`lpr-detail`},
  btn(`← Back`,()=>setDetail(null),`text`),
  h(`div`,{className:`lpr-detail-hero lpr-art-${activeGame.art||`custom`}`},h(`img`,{className:`lpr-detail-image`,alt:`${activeGame.name} game artwork`,src:imageFor(activeGame.id),onError:e=>{e.currentTarget.style.display=`none`}}),h(`span`,{className:`lpr-detail-glyph`},(activeGame.name||`R`)[0].toUpperCase())),
  h(`div`,{className:`lpr-detail-info`},h(`div`,{},h(`div`,{className:`lpr-eyebrow`},`EXPERIENCE`),h(`h1`,{},activeGame.name),h(`p`,{},activeGame.subtitle||activeGame.genre||`Roblox experience`),h(`small`,{},`Place ID ${activeGame.id}`)),
   h(`div`,{className:`lpr-button-row`},btn(`▶ Play`,()=>launch({kind:`game`,id:activeGame.id}),`primary`),
    btn(catalogSaved(activeGame.id)?`Remove from library`:`Save to library`,()=>catalogSaved(activeGame.id)?persist(`favorites.v1`,favorites.filter(g=>g.id!==activeGame.id),setFavorites):saveGame(activeGame)),
    btn(`Open on Roblox ↗`,()=>browse(gameUrl(activeGame.id))))),
  h(`div`,{className:`lpr-panel`},h(`h2`,{},`About this experience`),h(`p`,{},`For live player counts, creator updates, servers, reviews and purchases, use the official Roblox game page.`),btn(`Visit game page ↗`,()=>browse(gameUrl(activeGame.id)),`secondary`))):null;
 const items=[[`home`,`Home`],[`discover`,`Discover`],[`library`,`Library`],[`friends`,`Friends`],[`avatar`,`Avatar`],[`marketplace`,`Marketplace`],[`accounts`,`Accounts`],[`settings`,`Settings`]];
 const header=h(`header`,{className:`lpr-topbar`},h(`div`,{className:`lpr-wordmark`},h(lunarPatcherRobloxIcon,{size:22}),h(`strong`,{},`Roblox`)),
  h(`form`,{className:`lpr-search`,onSubmit:e=>{e.preventDefault();navTo(`discover`)}},h(`span`,{},`⌕`),h(`input`,{type:`search`,value:query,onChange:e=>setQuery(e.target.value.slice(0,90)),placeholder:`Search experiences`,"aria-label":`Search Roblox experiences`}),btn(`Search`,()=>navTo(`discover`),`primary`)),
  btn(`Open Roblox ↗`,()=>browse(`https://www.roblox.com/home`)));
 const navigation=h(`nav`,{className:`lpr-navigation`,"aria-label":`Roblox sections`},...items.map(([id,label])=>h(`button`,{type:`button`,className:`lpr-nav-item ${page===id?`is-active`:``}`,onClick:()=>navTo(id),"aria-label":label,"aria-current":page===id?`page`:void 0,key:id},h(`span`,{className:`lpr-nav-icon`},icon(id)),label)));
 return h(`main`,{className:`lunar-patcher-roblox lpr-density-${[`compact`,`comfortable`,`spacious`].includes(settings.hubLayout)?settings.hubLayout:`comfortable`} ${settings.reduceMotion?`lpr-reduced-motion`:``}`,style:{"--lpr-custom-accent":/^#[0-9a-f]{6}$/i.test(settings.hubColor||``)?settings.hubColor:`#E5A1BB`}},header,navigation,
  h(`div`,{className:`lpr-scroll`},details||({home,discover,library,friends,avatar,marketplace,accounts:accountsPage,settings:settingsPage}[page]||home)),
  h(`div`,{className:`lpr-status`,role:`status`},notice||`Launches through the official Roblox Player. Account features use Roblox.com.`));
},
