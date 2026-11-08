'use strict';
const { contextBridge, ipcRenderer } = require('electron');
const ALLOWED = new Set(['status','check','apply','restore']);
contextBridge.exposeInMainWorld('lunarPatcher', Object.freeze({
  getAppInfo: () => ipcRenderer.invoke('app:info'),
  pickFolder: () => ipcRenderer.invoke('app:browse'),
  getStatus: (folder) => ipcRenderer.invoke('engine:run', { operation:'status', folder }),
  check: (folder, mode, color, robloxEnabled=false) => ipcRenderer.invoke('engine:run', {operation:'check', folder, mode, color, robloxEnabled}),
  apply: (folder, mode, color, robloxEnabled=false) => ipcRenderer.invoke('engine:run', {operation:'apply', folder, mode, color, robloxEnabled}),
  restore: (folder) => ipcRenderer.invoke('engine:run', {operation:'restore', folder}),
  getPreferences: () => ipcRenderer.invoke('prefs:read'),
  savePreferences: (prefs) => ipcRenderer.invoke('prefs:save', prefs),
  checkUpdates: () => ipcRenderer.invoke('updates:check'),
  installUpdate: () => ipcRenderer.invoke('updates:install'),
  openRepository: () => ipcRenderer.invoke('app:repo'),
  launchLunar: (folder) => ipcRenderer.invoke('app:launch', folder),
  robloxStatus: () => ipcRenderer.invoke('roblox:status'),
  robloxPreferences: () => ipcRenderer.invoke('roblox:prefs-read'),
  saveRobloxPreferences: (prefs) => ipcRenderer.invoke('roblox:prefs-save', prefs),
  launchRoblox: (request) => ipcRenderer.invoke('roblox:launch', request),
  onUpdate: (callback) => {
    if (typeof callback !== 'function') return () => {};
    const listener = (_event, payload) => callback(payload);
    ipcRenderer.on('updates:status', listener);
    return () => ipcRenderer.removeListener('updates:status', listener);
  }
}));
