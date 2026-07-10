interface ElectronAPI {
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

export {}
