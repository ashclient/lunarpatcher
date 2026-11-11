/* LunarPatcher Roblox browser: isolated official-site BrowserView inside Lunar's window.
 * No login handling, cookie access, or unsafe Electron webview flags.
 * This code runs only in Lunar's main process after the user enables the optional tab.
 */
;(() => {
  'use strict';
  const {ipcMain, BrowserWindow, BrowserView, shell, session} = require('electron');
  const records = new Map();
  const channels = ['show','bounds','navigate','back','forward','reload','home','hide','external'];
  const isOfficial = input => {
    if (typeof input !== 'string' || input.length > 2048) return false;
    try {
      const url = new URL(input);
      const hostname = url.hostname.toLowerCase();
      return url.protocol === 'https:' && (hostname === 'roblox.com' || hostname.endsWith('.roblox.com')) && !url.username && !url.password;
    } catch { return false; }
  };
  // Roblox's OS-installed player handles this link. Never run child processes here.
  const isPlayerLink = value => {
    if (typeof value !== 'string' || value.length > 2048 || /[\r\n\x00-\x1f]/.test(value)) return false;
    if (/^roblox-player:1\+launchmode:(?:play|app)(?:\+|$)/i.test(value)) return true;
    try {
      const uri = new URL(value);
      return uri.protocol === 'roblox:' && uri.hostname === 'experiences' && uri.pathname === '/start'
        && /^[1-9][0-9]{0,19}$/.test(uri.searchParams.get('placeId') || '');
    } catch { return false; }
  };
  // A normal Electron BrowserView is not recognized as an installed Roblox
  // browser integration, even when Roblox Player is installed on Windows.
  // Only intercept a deliberate Play click on an official /games/{placeId} page.
  // The fixed, minimal hook does not inspect cookies, tokens, login forms or APIs.
  const playHookSource = String.raw`(() => {
    if (window.__lunarPatcherDirectPlay) return;
    window.__lunarPatcherDirectPlay = true;
    const onGamePage = () => /^\/games\/([1-9][0-9]{0,19})(?:\/|$)/.exec(location.pathname);
    const isPlayControl = element => {
      if (!element || typeof element.closest !== 'function') return false;
      const button = element.closest('button,[role="button"],a,#game-detail-play-button,.game-play-button');
      if (!button || button.disabled || button.closest('[role="dialog"],[aria-modal="true"]')) return false;
      const id = (button.id || '').toLowerCase();
      const testid = (button.getAttribute('data-testid') || '').toLowerCase();
      const cls = (typeof button.className === 'string' ? button.className : '').toLowerCase();
      const label = (button.getAttribute('aria-label') || button.getAttribute('title') || '').toLowerCase().trim();
      const text = (button.textContent || '').toLowerCase().trim();
      return id === 'game-detail-play-button' || /(?:^|[-_])(?:game[-_])?play(?:[-_]button)?(?:$|[-_])/.test(testid)
        || /(?:^|\s)(?:game-detail-play-button|game-play-button|btn-play)(?:$|\s)/.test(cls)
        || label === 'play' || text === 'play'
        || !!button.querySelector('[class*="icon-play"],[data-testid="play-icon"],svg[aria-label="Play"]');
    };
    document.addEventListener('click', event => {
      const match = onGamePage();
      if (!match || !isPlayControl(event.target) || !event.isTrusted) return;
      event.preventDefault();
      event.stopImmediatePropagation();
      // Hand the selected public place to the OS-installed Roblox Player.
      // No auth tickets, cookies, account data or backend calls are copied.
      window.open('roblox://experiences/start?placeId=' + match[1], '_blank');
    }, true);
  })();`;
  const owner = event => {
    const sender = event.sender;
    const url = sender && sender.getURL();
    if (typeof url !== 'string' || !url.startsWith('file:') || !/\/dist\/html\/main\.html(?:$|[?#])/.test(url)) throw new Error('Roblox browser IPC denied: not the Lunar launcher');
    const win = BrowserWindow.fromWebContents(sender);
    if (!win || win.isDestroyed()) throw new Error('Lunar window unavailable');
    return win;
  };
  const emit = (win, state) => {
    if (win && !win.isDestroyed() && !win.webContents.isDestroyed()) {
      win.webContents.send('lunarpatcher:roblox:state', state);
    }
  };
  const get = win => records.get(win.id);
  const validBounds = input => {
    if (!input || typeof input !== 'object') throw new Error('Missing browser bounds');
    const all = ['x','y','width','height'].map(key => Number(input[key]));
    if (all.some(x => !Number.isFinite(x))) throw new Error('Invalid browser bounds');
    const [x,y,width,height] = all.map(Math.round);
    if (width < 100 || height < 100 || width > 10000 || height > 10000 || Math.abs(x) > 20000 || Math.abs(y) > 20000) throw new Error('Browser bounds out of range');
    return {x:Math.max(0,x), y:Math.max(0,y), width, height};
  };
  function close(win) {
    const row = get(win);
    if (!row) return;
    records.delete(win.id);
    try { win.removeBrowserView(row.view); } catch {}
    try { if (!row.view.webContents.isDestroyed()) row.view.webContents.destroy(); } catch {}
  }
  function notifyNavigation(win, row) {
    try {
      const web = row.view.webContents;
      const url = web.getURL();
      emit(win, {phase:'ready',url:isOfficial(url)?url:'https://www.roblox.com/home',canGoBack:web.canGoBack(),canGoForward:web.canGoForward()});
    } catch {}
  }
  async function handleExternal(win, value) {
    if (!isPlayerLink(value)) return false;
    try {
      await shell.openExternal(value);
      emit(win, {phase:'launch-requested',message:'Sent Roblox Player link to Windows. Check Roblox Player to confirm the game joined.'});
      return true;
    } catch (error) {
      // Missing/broken OS protocol association: use the normal browser, which
      // can invoke its own Roblox Player integration. Never auto-install anything.
      if (value.startsWith('roblox://experiences/start?')) {
        const placeId = new URL(value).searchParams.get('placeId');
        try {
          await shell.openExternal('https://www.roblox.com/games/' + placeId);
          emit(win, {phase:'error',message:'Roblox protocol could not open; the game was opened in your browser instead.'});
          return false;
        } catch {}
      }
      emit(win, {phase:'error',message:'Roblox Player could not be opened: '+String(error.message || error)});
      return false;
    }
  }
  function create(win) {
    const view = new BrowserView({
      webPreferences: {
        nodeIntegration:false, contextIsolation:true, sandbox:true,
        webSecurity:true, allowRunningInsecureContent:false,
        partition:'persist:lunarpatcher-roblox'
      }
    });
    const row = {view};
    records.set(win.id,row);
    const web = view.webContents;
    // Only the official site loads; all device permissions default to denied.
    const dedicatedSession = session.fromPartition('persist:lunarpatcher-roblox');
    dedicatedSession.setPermissionRequestHandler((_contents,_permission,callback)=>callback(false));
    web.setWindowOpenHandler(({url}) => {
      if (isOfficial(url)) { void web.loadURL(url).catch(error=>emit(win,{phase:'error',message:String(error.message||error)})); }
      else if (isPlayerLink(url)) { void handleExternal(win,url); }
      return {action:'deny'};
    });
    web.on('will-navigate',(event,url)=>{
      if (isOfficial(url)) return;
      event.preventDefault();
      if (isPlayerLink(url)) void handleExternal(win,url);
      else emit(win,{phase:'blocked',message:'Only roblox.com pages can open inside Lunar.'});
    });
    web.on('will-redirect',(event,url)=>{
      if (isOfficial(url)) return;
      event.preventDefault();
      if (isPlayerLink(url)) void handleExternal(win,url);
    });
    web.on('did-navigate',()=>notifyNavigation(win,row));
    web.on('did-navigate-in-page',()=>notifyNavigation(win,row));
    web.on('dom-ready',()=>{
      // Install on official Roblox.com (including Home), so it survives
      // single-page navigation to a game. Only genuine game Play clicks act.
      const location = web.getURL();
      if (!/^https:\/\/www\.roblox\.com(?:\/|$)/.test(location)) return;
      void web.executeJavaScript(playHookSource, true).catch(error => emit(win, {
        phase:'error', message:'Could not enable direct Roblox Play: '+String(error.message || error)
      }));
    });
    web.on('did-fail-load',(_event,errorCode,description,_url,isMainFrame)=>{
      if (isMainFrame !== false && errorCode !== -3) emit(win,{phase:'error',message:'Roblox page did not load: '+description});
    });
    web.on('did-start-loading',()=>emit(win,{phase:'loading'}));
    win.on('closed',()=>close(win));
    win.addBrowserView(view);
    // Roblox login cookies are maintained only by Electron's isolated session.
    void web.loadURL('https://www.roblox.com/home').catch(error=>emit(win,{phase:'error',message:String(error.message||error)}));
    return row;
  }
  ipcMain.handle('lunarpatcher:roblox:show',(event,bounds)=>{
    const win = owner(event);
    const row = get(win) || create(win);
    row.view.setBounds(validBounds(bounds));
    if (typeof win.setTopBrowserView === 'function') win.setTopBrowserView(row.view);
    return {ok:true};
  });
  ipcMain.handle('lunarpatcher:roblox:bounds',(event,bounds)=>{
    const win=owner(event),row=get(win);
    if (!row) return false;
    row.view.setBounds(validBounds(bounds));
    return true;
  });
  ipcMain.handle('lunarpatcher:roblox:navigate',async(event,url)=>{
    const win=owner(event),row=get(win);
    if (!row || !isOfficial(url)) throw new Error('Only https://roblox.com pages are allowed');
    await row.view.webContents.loadURL(url);
    return true;
  });
  ipcMain.handle('lunarpatcher:roblox:back',event=>{
    const win=owner(event),row=get(win);
    if (row && row.view.webContents.canGoBack()) row.view.webContents.goBack();
    return !!row;
  });
  ipcMain.handle('lunarpatcher:roblox:forward',event=>{
    const win=owner(event),row=get(win);
    if (row && row.view.webContents.canGoForward()) row.view.webContents.goForward();
    return !!row;
  });
  ipcMain.handle('lunarpatcher:roblox:reload',event=>{
    const win=owner(event),row=get(win); if(row)row.view.webContents.reload();return !!row;
  });
  ipcMain.handle('lunarpatcher:roblox:home',async event=>{
    const win=owner(event),row=get(win);
    if (!row) return false;
    await row.view.webContents.loadURL('https://www.roblox.com/home'); return true;
  });
  ipcMain.handle('lunarpatcher:roblox:external',async(event,url)=>{
    owner(event);
    if (!isOfficial(url)) throw new Error('External navigation restricted to roblox.com');
    await shell.openExternal(url);
    return true;
  });
  ipcMain.handle('lunarpatcher:roblox:hide',event=>{close(owner(event));return true;});
  // On launcher quit, Electron disposes the view and encrypted browser-session cookies.
})();
