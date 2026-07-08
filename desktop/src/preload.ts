import { contextBridge, ipcRenderer } from "electron";

export interface ElectronAPI {
  platform: string;
  getToken: () => Promise<string | null>;
  setToken: (token: string) => Promise<boolean>;
  clearToken: () => Promise<boolean>;
  printHtml: (html: string) => Promise<void>;
  isOnline: () => boolean;
  onOnlineStatusChange: (callback: (online: boolean) => void) => void;
}

declare global {
  interface Window {
    electronAPI: ElectronAPI;
  }
}

contextBridge.exposeInMainWorld("electronAPI", {
  platform: process.platform,
  getToken: () => ipcRenderer.invoke("get-token"),
  setToken: (token: string) => ipcRenderer.invoke("set-token", token),
  clearToken: () => ipcRenderer.invoke("clear-token"),
  printHtml: (html: string) => ipcRenderer.invoke("print-html", html),
  isOnline: () => window.navigator.onLine,
  onOnlineStatusChange: (callback: (online: boolean) => void) => {
    window.addEventListener("online", () => callback(true));
    window.addEventListener("offline", () => callback(false));
  },
});
