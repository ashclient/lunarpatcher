/* Expose the smallest possible bridge only to Lunar's trusted local renderer. */
;(() => {
  'use strict';
  const {contextBridge, ipcRenderer} = require('electron');
  const invoke = (action,...args) => ipcRenderer.invoke('lunarpatcher:roblox:'+action,...args);
  contextBridge.exposeInMainWorld('lunarPatcherRobloxBrowser', Object.freeze({
    show: bounds => invoke('show',bounds),
    bounds: bounds => invoke('bounds',bounds),
    navigate: url => invoke('navigate',url),
    back: () => invoke('back'),
    forward: () => invoke('forward'),
    reload: () => invoke('reload'),
    home: () => invoke('home'),
    hide: () => invoke('hide'),
    external: url => invoke('external',url),
    onState: callback => {
      if (typeof callback !== 'function') return () => {};
      const handler = (_event,data) => callback(data);
      ipcRenderer.on('lunarpatcher:roblox:state', handler);
      return () => ipcRenderer.removeListener('lunarpatcher:roblox:state', handler);
    }
  }));
})();
