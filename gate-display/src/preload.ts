import { contextBridge, ipcRenderer } from 'electron'

export interface ElectronAPI {
  readConfig: () => Promise<Record<string, unknown> | null>
  writeConfig: (config: unknown) => Promise<boolean>
  quitApp: () => Promise<void>
  getDeviceId: () => Promise<string>
  getIP: () => Promise<string>
}

declare global {
  interface Window {
    electronAPI: ElectronAPI
  }
}

contextBridge.exposeInMainWorld('electronAPI', {
  readConfig: () => ipcRenderer.invoke('read-config'),
  writeConfig: (config: unknown) => ipcRenderer.invoke('write-config', config),
  quitApp: () => ipcRenderer.invoke('quit-app'),
  getDeviceId: () => ipcRenderer.invoke('get-device-id'),
  getIP: () => ipcRenderer.invoke('get-ip'),
})
