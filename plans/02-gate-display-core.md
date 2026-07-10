# Milestone 2: Gate Display App — Core Display

## Objective

Build the gate display Electron app with the state machine, mock controller, and all visual components.

## Files

| # | File | Action |
|---|------|--------|
| 1 | `gate-display/package.json` | **New** |
| 2 | `gate-display/tsconfig.json` | **New** |
| 3 | `gate-display/tsconfig.main.json` | **New** |
| 4 | `gate-display/src/main.ts` | **New** — frameless fullscreen window |
| 5 | `gate-display/src/preload.ts` | **New** — read/write config, quit |
| 6 | `gate-display/src/renderer/index.html` | **New** |
| 7 | `gate-display/src/renderer/index.tsx` | **New** |
| 8 | `gate-display/src/renderer/App.tsx` | **New** — routes: registration vs display |
| 9 | `gate-display/src/renderer/App.css` | **New** — dark industrial CSS |
| 10 | `gate-display/src/renderer/lib/api.ts` | **New** — `fetchGateInfo()` |
| 11 | `gate-display/src/renderer/lib/gateMachine.ts` | **New** — pure state machine |
| 12 | `gate-display/src/renderer/lib/controller.ts` | **New** — `ControllerInterface` + `MockController` |
| 13 | `gate-display/src/renderer/hooks/useGateState.ts` | **New** — React hook |
| 14 | `gate-display/src/renderer/screens/GateDisplay.tsx` | **New** — assembles all components |
| 15 | `gate-display/src/renderer/screens/RegistrationScreen.tsx` | **New** — placeholder for M3 (just text) |
| 16 | `gate-display/src/renderer/components/Header.tsx` | **New** |
| 17 | `gate-display/src/renderer/components/WelcomeSign.tsx` | **New** |
| 18 | `gate-display/src/renderer/components/CameraFeed.tsx` | **New** |
| 19 | `gate-display/src/renderer/components/LoopIndicator.tsx` | **New** |
| 20 | `gate-display/src/renderer/components/TicketButton.tsx` | **New** |
| 21 | `gate-display/src/renderer/components/InstructionText.tsx` | **New** |
| 22 | `gate-display/src/renderer/components/RatesTable.tsx` | **New** |
| 23 | `gate-display/src/renderer/components/GateBarrier.tsx` | **New** |
| 24 | `gate-display/src/renderer/components/DebugPanel.tsx` | **New** |
| 25 | `gate-display/config.json` | **Create manually for testing** |

## 2.1 Project scaffold

### `gate-display/package.json`

```json
{
  "name": "parkir-gate-display",
  "version": "0.1.0",
  "private": true,
  "main": "dist/main.js",
  "engines": { "node": ">=20.0.0" },
  "scripts": {
    "dev": "npm run build && electron .",
    "build": "tsc --project tsconfig.main.json && npm run build-renderer && npm run copy-renderer",
    "build-renderer": "esbuild src/renderer/index.tsx --bundle --outfile=dist/renderer/renderer.js --minify --platform=browser --external:electron",
    "copy-renderer": "cp src/renderer/index.html dist/renderer/index.html && cp src/renderer/App.css dist/renderer/App.css",
    "start": "electron .",
    "test": "vitest run"
  },
  "dependencies": {
    "electron": "^31.0.0",
    "react": "^18.3.1",
    "react-dom": "^18.3.1"
  },
  "devDependencies": {
    "@types/node": "^20",
    "@types/react": "^18",
    "@types/react-dom": "^18",
    "esbuild": "^0.28.1",
    "typescript": "^5",
    "vitest": "^2.0.0"
  }
}
```

### `gate-display/tsconfig.main.json`

```json
{
  "compilerOptions": {
    "target": "ES2022",
    "module": "commonjs",
    "lib": ["ES2022"],
    "outDir": "dist",
    "rootDir": "src",
    "strict": true,
    "esModuleInterop": true,
    "skipLibCheck": true,
    "forceConsistentCasingInFileNames": true,
    "resolveJsonModule": true
  },
  "include": ["src/main.ts", "src/preload.ts"]
}
```

### `gate-display/tsconfig.json`

```json
{
  "compilerOptions": {
    "target": "ES2022",
    "module": "ES2022",
    "moduleResolution": "bundler",
    "lib": ["ES2022", "DOM", "DOM.Iterable"],
    "jsx": "react-jsx",
    "strict": true,
    "esModuleInterop": true,
    "skipLibCheck": true,
    "forceConsistentCasingInFileNames": true
  },
  "include": ["src/**/*.ts", "src/**/*.tsx"]
}
```

## 2.2 Main process — `gate-display/src/main.ts`

```typescript
import { app, BrowserWindow, ipcMain } from 'electron'
import * as path from 'path'
import * as fs from 'fs'

const CONFIG_PATH = path.join(__dirname, '..', 'config.json')

let mainWindow: BrowserWindow | null = null

function createWindow() {
  mainWindow = new BrowserWindow({
    fullscreen: true,
    frame: false,
    backgroundColor: '#1a1a1a',
    webPreferences: {
      preload: path.join(__dirname, 'preload.js'),
      contextIsolation: true,
      nodeIntegration: false,
    },
  })

  mainWindow.loadFile(path.join(__dirname, 'renderer', 'index.html'))

  mainWindow.webContents.on('before-input-event', (_event, input) => {
    if (input.key === 'F11') {
      mainWindow?.setFullScreen(!mainWindow?.isFullScreen())
    }
    if (input.key === 'F12') {
      mainWindow?.webContents.toggleDevTools()
    }
  })

  mainWindow.on('closed', () => { mainWindow = null })
}

app.whenReady().then(createWindow)

app.on('window-all-closed', () => {
  if (process.platform !== 'darwin') app.quit()
})

app.on('activate', () => {
  if (mainWindow === null) createWindow()
})

// IPC handlers
ipcMain.handle('read-config', () => {
  try {
    const raw = fs.readFileSync(CONFIG_PATH, 'utf-8')
    return JSON.parse(raw)
  } catch { return null }
})

ipcMain.handle('write-config', (_event, config: Record<string, unknown>) => {
  fs.writeFileSync(CONFIG_PATH, JSON.stringify(config, null, 2))
  return true
})

ipcMain.handle('quit-app', () => { app.quit() })
```

## 2.3 Preload — `gate-display/src/preload.ts`

```typescript
import { contextBridge, ipcRenderer } from 'electron'

contextBridge.exposeInMainWorld('electronAPI', {
  readConfig: () => ipcRenderer.invoke('read-config'),
  writeConfig: (config: unknown) => ipcRenderer.invoke('write-config', config),
  quitApp: () => ipcRenderer.invoke('quit-app'),
})
```

Declare global types:

```typescript
export interface ElectronAPI {
  readConfig: () => Promise<Record<string, unknown> | null>
  writeConfig: (config: unknown) => Promise<boolean>
  quitApp: () => Promise<void>
}

declare global {
  interface Window { electronAPI: ElectronAPI }
}
```

## 2.4 Renderer entry — `gate-display/src/renderer/index.html`

```html
<!DOCTYPE html>
<html lang="id">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>PARKIR Gate</title>
  <link rel="stylesheet" href="App.css">
</head>
<body>
  <div id="root"></div>
  <script src="renderer.js"></script>
</body>
</html>
```

### `gate-display/src/renderer/index.tsx`

```tsx
import { createRoot } from 'react-dom/client'
import App from './App'

createRoot(document.getElementById('root')!).render(<App />)
```

## 2.5 App — `gate-display/src/renderer/App.tsx`

```tsx
import { useEffect, useState } from 'react'
import RegistrationScreen from './screens/RegistrationScreen'
import GateDisplay from './screens/GateDisplay'

interface Config {
  api_url: string
  location_id: string
}

export default function App() {
  const [config, setConfig] = useState<Config | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    window.electronAPI.readConfig().then((cfg) => {
      if (cfg && typeof cfg.api_url === 'string' && typeof cfg.location_id === 'string') {
        setConfig(cfg as Config)
      }
      setLoading(false)
    })
  }, [])

  if (loading) return null
  if (!config) return <RegistrationScreen />
  return <GateDisplay apiUrl={config.api_url} locationId={config.location_id} />
}
```

## 2.6 State machine — `gate-display/src/renderer/lib/gateMachine.ts`

Pure functions — no React, testable with vitest.

### Types

```typescript
export type GateState =
  | 'IDLE'
  | 'VEHICLE_DETECTED'
  | 'TICKET_PRESSED'
  | 'TICKET_READY'
  | 'GATE_OPENING'
  | 'GATE_OPEN'
  | 'VEHICLE_EXITED'

export type ControllerEvent =
  | { type: 'LOOP_ON' }
  | { type: 'LOOP_OFF' }
  | { type: 'BUTTON_PRESSED' }
  | { type: 'DISPENSE_OK' }

export type TimeoutAction =
  | 'RESET_IDLE'
  | 'ADVANCE_GATE'
  | 'ADVANCE_GATE_DONE'
  | 'ADVANCE_EXIT'

export type ControllerCommand =
  | { type: 'DISPENSE_TICKET' }
  | { type: 'BARRIER_OPEN' }
  | { type: 'BARRIER_CLOSE' }

export interface GateMachineResult {
  state: GateState
  command: ControllerCommand | null
  timeout: TimeoutAction | null
}
```

### Transition function

```typescript
export function transition(
  current: GateState,
  event: ControllerEvent | TimeoutAction
): GateMachineResult {
  const key = `${current}__${typeof event === 'string' ? event : event.type}`

  switch (key) {
    // Hardware events
    case 'IDLE__LOOP_ON':
      return { state: 'VEHICLE_DETECTED', command: null, timeout: null }

    case 'VEHICLE_DETECTED__LOOP_OFF':
      return { state: 'IDLE', command: null, timeout: null }

    case 'VEHICLE_DETECTED__BUTTON_PRESSED':
      return { state: 'TICKET_PRESSED', command: { type: 'DISPENSE_TICKET' }, timeout: null }

    case 'TICKET_PRESSED__DISPENSE_OK':
      return { state: 'TICKET_READY', command: null, timeout: 'ADVANCE_GATE' }

    case 'GATE_OPEN__LOOP_OFF':
      return { state: 'VEHICLE_EXITED', command: { type: 'BARRIER_CLOSE' }, timeout: 'ADVANCE_EXIT' }

    // Timeouts
    case 'VEHICLE_DETECTED__RESET_IDLE':
      return { state: 'IDLE', command: null, timeout: null }

    case 'TICKET_READY__ADVANCE_GATE':
      return { state: 'GATE_OPENING', command: { type: 'BARRIER_OPEN' }, timeout: 'ADVANCE_GATE_DONE' }

    case 'GATE_OPENING__ADVANCE_GATE_DONE':
      return { state: 'GATE_OPEN', command: null, timeout: null }

    case 'VEHICLE_EXITED__ADVANCE_EXIT':
      return { state: 'IDLE', command: null, timeout: null }

    // No transition
    default:
      return { state: current, command: null, timeout: null }
  }
}
```

### Display state

```typescript
export interface DisplayState {
  instructionText: string
  loopIndicator: 'OFF' | 'AMBER'
  ticketIndicator: 'OFF' | 'GREEN'
  gateState: 'CLOSED' | 'OPENING' | 'OPEN'
  showRates: boolean
}

export function getDisplay(state: GateState): DisplayState {
  switch (state) {
    case 'IDLE':
      return { instructionText: '', loopIndicator: 'OFF', ticketIndicator: 'OFF', gateState: 'CLOSED', showRates: true }
    case 'VEHICLE_DETECTED':
      return { instructionText: 'Silakan tekan tombol untuk mengambil tiket', loopIndicator: 'AMBER', ticketIndicator: 'OFF', gateState: 'CLOSED', showRates: false }
    case 'TICKET_PRESSED':
      return { instructionText: 'Mencetak tiket...', loopIndicator: 'AMBER', ticketIndicator: 'GREEN', gateState: 'CLOSED', showRates: false }
    case 'TICKET_READY':
      return { instructionText: 'Silakan ambil tiket Anda', loopIndicator: 'AMBER', ticketIndicator: 'GREEN', gateState: 'CLOSED', showRates: false }
    case 'GATE_OPENING':
      return { instructionText: 'Pintu terbuka... Silakan masuk', loopIndicator: 'OFF', ticketIndicator: 'OFF', gateState: 'OPENING', showRates: false }
    case 'GATE_OPEN':
      return { instructionText: 'Selamat datang. Silakan masuk dengan hati-hati', loopIndicator: 'OFF', ticketIndicator: 'OFF', gateState: 'OPEN', showRates: false }
    case 'VEHICLE_EXITED':
      return { instructionText: 'Terima kasih', loopIndicator: 'OFF', ticketIndicator: 'OFF', gateState: 'OPEN', showRates: false }
  }
}
```

## 2.7 Controller — `gate-display/src/renderer/lib/controller.ts`

```typescript
export type { ControllerEvent, ControllerCommand }

export interface ControllerInterface {
  onEvent: (cb: (event: ControllerEvent) => void) => void
  sendCommand: (cmd: ControllerCommand) => Promise<void>
  connect: () => Promise<void>
  disconnect: () => void
}

export class MockController implements ControllerInterface {
  private listeners: Array<(event: ControllerEvent) => void> = []

  connect() { /* no-op for mock */ }
  disconnect() { this.listeners = [] }

  onEvent(cb: (event: ControllerEvent) => void) {
    this.listeners.push(cb)
  }

  sendCommand(cmd: ControllerCommand) {
    console.log('[MOCK CONTROLLER] command:', cmd)
    return Promise.resolve()
  }

  // Public trigger methods for DebugPanel
  trigger(event: ControllerEvent) {
    this.listeners.forEach((cb) => cb(event))
  }
}
```

## 2.8 useGateState hook — `gate-display/src/renderer/hooks/useGateState.ts`

```typescript
import { useEffect, useRef, useState, useCallback } from 'react'
import type { ControllerInterface, ControllerEvent, TimeoutAction } from '../lib/controller'
import { transition, getDisplay, type DisplayState, type GateState } from '../lib/gateMachine'

const TIMEOUT_DURATIONS: Record<string, number> = {
  RESET_IDLE: 30000,
  ADVANCE_GATE: 3000,
  ADVANCE_GATE_DONE: 2000,
  ADVANCE_EXIT: 2000,
}

export function useGateState(controller: ControllerInterface) {
  const [currentState, setCurrentState] = useState<GateState>('IDLE')
  const [display, setDisplay] = useState<DisplayState>(getDisplay('IDLE'))
  const timeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  const clearTimer = useCallback(() => {
    if (timeoutRef.current !== null) {
      clearTimeout(timeoutRef.current)
      timeoutRef.current = null
    }
  }, [])

  const dispatch = useCallback((event: ControllerEvent | TimeoutAction) => {
    clearTimer()
    setCurrentState((prev) => {
      const result = transition(prev, event)

      // Send command if any
      if (result.command) {
        controller.sendCommand(result.command)
      }

      // Set timeout if any
      if (result.timeout) {
        const ms = TIMEOUT_DURATIONS[result.timeout]
        const action = result.timeout
        timeoutRef.current = setTimeout(() => dispatch(action), ms)
      }

      setDisplay(getDisplay(result.state))
      return result.state
    })
  }, [controller, clearTimer])

  useEffect(() => {
    controller.onEvent((event) => dispatch(event))
    return () => clearTimer()
  }, [controller, dispatch, clearTimer])

  return display
}
```

## 2.9 Display components

### Header — `gate-display/src/renderer/components/Header.tsx`

```tsx
import { useEffect, useState } from 'react'

interface Props { locationName: string }

export default function Header({ locationName }: Props) {
  const [time, setTime] = useState(new Date())

  useEffect(() => {
    const id = setInterval(() => setTime(new Date()), 1000)
    return () => clearInterval(id)
  }, [])

  const fmt = time.toLocaleDateString('id-ID', {
    weekday: 'short', day: '2-digit', month: 'short', year: 'numeric',
    hour: '2-digit', minute: '2-digit', second: '2-digit',
    timeZone: 'Asia/Jakarta',
  })

  return (
    <div className="header">
      <span className="location-name">{locationName}</span>
      <span className="clock">{fmt}</span>
    </div>
  )
}
```

### WelcomeSign — `gate-display/src/renderer/components/WelcomeSign.tsx`

```tsx
export default function WelcomeSign() {
  return (
    <div style={{ textAlign: 'center' }}>
      <div className="welcome">WELCOME TO</div>
      <div className="welcome" style={{ marginTop: -8 }}>PARKIR</div>
    </div>
  )
}
```

### CameraFeed — `gate-display/src/renderer/components/CameraFeed.tsx`

```tsx
export default function CameraFeed() {
  return (
    <div className="camera-frame">
      <span>CAMERA</span>
    </div>
  )
}
```

### LoopIndicator — `gate-display/src/renderer/components/LoopIndicator.tsx`

```tsx
interface Props { state: 'OFF' | 'AMBER' }

export default function LoopIndicator({ state }: Props) {
  return (
    <div className="indicator">
      <span className={`indicator-dot ${state === 'AMBER' ? 'amber' : 'off'}`} />
      <span>Loop: {state === 'AMBER' ? 'TERDETEKSI' : '—'}</span>
    </div>
  )
}
```

### TicketButton — `gate-display/src/renderer/components/TicketButton.tsx`

```tsx
interface Props { state: 'OFF' | 'GREEN' }

export default function TicketButton({ state }: Props) {
  return (
    <div className="indicator">
      <span className={`indicator-dot ${state === 'GREEN' ? 'green' : 'off'}`} />
      <span>Tiket: {state === 'GREEN' ? 'DITEKAN' : '—'}</span>
    </div>
  )
}
```

### InstructionText — `gate-display/src/renderer/components/InstructionText.tsx`

```tsx
interface Props { text: string }

export default function InstructionText({ text }: Props) {
  if (!text) return null
  return <div className="instruction">{text}</div>
}
```

### RatesTable — `gate-display/src/renderer/components/RatesTable.tsx`

```tsx
interface RateRow {
  vehicle_type: string
  first_hour_rate: number
  subsequent_hourly_rate: number
  daily_flat_rate: number
}

interface Props { rates: RateRow[]; visible: boolean }

export default function RatesTable({ rates, visible }: Props) {
  if (!visible || rates.length === 0) return null

  return (
    <div className="rates-panel">
      <div className="rates-title">TARIF</div>
      <table className="rates-table">
        <tbody>
          {rates.map((r) => (
            <tr key={r.vehicle_type}>
              <td>{r.vehicle_type}</td>
              <td>Rp {r.first_hour_rate.toLocaleString('id-ID')} / jam</td>
              <td>Rp {r.daily_flat_rate.toLocaleString('id-ID')} / hari</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}
```

### GateBarrier — `gate-display/src/renderer/components/GateBarrier.tsx`

```tsx
interface Props { state: 'CLOSED' | 'OPENING' | 'OPEN' }

const LABELS: Record<string, string> = {
  CLOSED: 'TERTUTUP',
  OPENING: 'MEMBUKA',
  OPEN: 'TERBUKA',
}

export default function GateBarrier({ state }: Props) {
  return (
    <div className={`gate-barrier ${state.toLowerCase()}`}>
      {LABELS[state]}
    </div>
  )
}
```

### DebugPanel — `gate-display/src/renderer/components/DebugPanel.tsx`

```tsx
import type { MockController } from '../lib/controller'
import { useState } from 'react'

interface Props { controller: MockController }

export default function DebugPanel({ controller }: Props) {
  const [autoMode, setAutoMode] = useState(true)
  const autoRef = useState<ReturnType<typeof setInterval> | null>([null])[1]
  // ^ use ref instead

  const triggerLoop = () => controller.trigger({ type: 'LOOP_ON' })
  const triggerButton = () => controller.trigger({ type: 'BUTTON_PRESSED' })
  const triggerDispenseOK = () => controller.trigger({ type: 'DISPENSE_OK' })
  const triggerLoopOff = () => controller.trigger({ type: 'LOOP_OFF' })
  const resetIdle = () => { /* handled internally via timeout */ }

  const toggleAuto = () => {
    const next = !autoMode
    setAutoMode(next)
    // In auto mode, the mock auto-cycles through events
    // In manual, only triggered by buttons
  }

  return (
    <div className="debug-panel">
      <button onClick={triggerLoop}>Loop ON</button>
      <button onClick={triggerButton}>Ticket</button>
      <button onClick={triggerDispenseOK}>Dispense OK</button>
      <button onClick={triggerLoopOff}>Loop OFF</button>
      <span style={{ margin: '0 8px', color: '#888' }}>|</span>
      <button onClick={toggleAuto}>{autoMode ? 'Manual' : 'Auto'}</button>
    </div>
  )
}
```

## 2.10 GateDisplay screen — `gate-display/src/renderer/screens/GateDisplay.tsx`

```tsx
import { useEffect, useState, useMemo } from 'react'
import { fetchGateInfo } from '../lib/api'
import { MockController } from '../lib/controller'
import { useGateState } from '../hooks/useGateState'
import Header from '../components/Header'
import WelcomeSign from '../components/WelcomeSign'
import CameraFeed from '../components/CameraFeed'
import LoopIndicator from '../components/LoopIndicator'
import TicketButton from '../components/TicketButton'
import InstructionText from '../components/InstructionText'
import RatesTable from '../components/RatesTable'
import GateBarrier from '../components/GateBarrier'
import DebugPanel from '../components/DebugPanel'

interface Props {
  apiUrl: string
  locationId: string
}

interface GateInfo {
  location: { name: string }
  rates: Array<{
    vehicle_type: string
    first_hour_rate: number
    subsequent_hourly_rate: number
    daily_flat_rate: number
  }>
}

export default function GateDisplay({ apiUrl, locationId }: Props) {
  const controller = useMemo(() => new MockController(), [])
  const display = useGateState(controller)
  const [gateInfo, setGateInfo] = useState<GateInfo | null>(null)
  const [locationName, setLocationName] = useState('PARKIR')

  useEffect(() => {
    const load = async () => {
      try {
        const info = await fetchGateInfo(apiUrl, locationId)
        setGateInfo(info)
        setLocationName(info.location.name)
      } catch {
        // keep default "PARKIR" as fallback
      }
    }
    load()
    const interval = setInterval(load, 60000)
    return () => clearInterval(interval)
  }, [apiUrl, locationId])

  return (
    <div className="gate-display">
      <Header locationName={locationName} />
      <div className="main-content">
        <div className="left-panel">
          <CameraFeed />
          <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
            <LoopIndicator state={display.loopIndicator} />
            <TicketButton state={display.ticketIndicator} />
          </div>
        </div>
        <div className="right-panel">
          <WelcomeSign />
          <InstructionText text={display.instructionText} />
          <RatesTable rates={gateInfo?.rates || []} visible={display.showRates} />
        </div>
      </div>
      <GateBarrier state={display.gateState} />
      <DebugPanel controller={controller} />
    </div>
  )
}
```

## 2.11 Registration screen (placeholder) — `gate-display/src/renderer/screens/RegistrationScreen.tsx`

```tsx
export default function RegistrationScreen() {
  return (
    <div className="registration-screen">
      <h1>PARKIR GATE</h1>
      <div className="registration-card">
        <p className="waiting-text">Menunggu registrasi...</p>
        <p className="hint">Daftarkan gate ini dari aplikasi PARKIR Desktop.</p>
      </div>
    </div>
  )
}
```

## 2.12 API client — `gate-display/src/renderer/lib/api.ts`

```typescript
export async function fetchGateInfo(apiUrl: string, locationId: string) {
  const res = await fetch(`${apiUrl}/api/v1/gate/${encodeURIComponent(locationId)}/info`)
  if (!res.ok) throw new Error(`gate info: ${res.status}`)
  const body = await res.json()
  return body.data as {
    location: { name: string; code: string }
    rates: Array<{
      vehicle_type: string
      first_hour_rate: number
      subsequent_hourly_rate: number
      daily_flat_rate: number
    }>
    capacity: Record<string, number>
  }
}
```

## 2.13 CSS — `gate-display/src/renderer/App.css`

```css
:root {
  --bg-primary: #1a1a1a;
  --bg-panel: #2a2a2a;
  --border: #444;
  --text-primary: #ffffff;
  --text-muted: #888;
  --amber: #f59e0b;
  --green: #22c55e;
  --red: #dc2626;
  --font-mono: 'SF Mono', 'Fira Code', 'Consolas', monospace;
  --font-sans: system-ui, -apple-system, 'Segoe UI', sans-serif;
}

* { margin: 0; padding: 0; box-sizing: border-box; }

html, body, #root { height: 100%; background: var(--bg-primary); color: var(--text-primary); }

.gate-display {
  height: 100vh;
  display: flex;
  flex-direction: column;
  font-family: var(--font-sans);
  overflow: hidden;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 24px;
  background: var(--bg-panel);
  border-bottom: 1px solid var(--border);
  font-family: var(--font-mono);
  font-size: 14px;
}

.location-name { font-weight: 600; }
.clock { color: var(--text-muted); }

.main-content {
  flex: 1;
  display: flex;
  padding: 24px;
  gap: 24px;
  min-height: 0;
}

.left-panel {
  width: 400px;
  display: flex;
  flex-direction: column;
  gap: 16px;
  flex-shrink: 0;
}

.right-panel {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 24px;
}

.camera-frame {
  aspect-ratio: 4/3;
  background: #111;
  border: 1px solid var(--border);
  display: flex;
  align-items: center;
  justify-content: center;
}

.camera-frame span { color: #555; font-family: var(--font-mono); font-size: 14px; letter-spacing: 2px; }

.indicator {
  display: flex;
  align-items: center;
  gap: 8px;
  font-family: var(--font-mono);
  font-size: 16px;
}

.indicator-dot {
  width: 12px;
  height: 12px;
  border-radius: 50%;
  flex-shrink: 0;
}
.indicator-dot.off { background: #444; }
.indicator-dot.amber { background: var(--amber); }
.indicator-dot.green { background: var(--green); }

.welcome {
  font-size: 48px;
  font-weight: 700;
  text-align: center;
  letter-spacing: 4px;
}

.instruction {
  font-size: 24px;
  text-align: center;
  min-height: 36px;
}

.rates-panel {
  background: var(--bg-panel);
  border: 1px solid var(--border);
  padding: 16px;
  min-width: 320px;
}

.rates-title {
  font-size: 12px;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 2px;
  margin-bottom: 8px;
}

.rates-table { width: 100%; border-collapse: collapse; }
.rates-table td { padding: 6px 8px; font-family: var(--font-mono); font-size: 14px; }
.rates-table td:first-child { font-weight: 600; }
.rates-table td:last-child { text-align: right; color: var(--text-muted); }

.gate-barrier {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  font-weight: 700;
  letter-spacing: 4px;
  transition: background-color 0.3s;
}
.gate-barrier.closed { background: var(--red); }
.gate-barrier.opening { background: var(--amber); color: #000; }
.gate-barrier.open { background: var(--green); }

.debug-panel {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  height: 44px;
  background: #333;
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 0 16px;
  font-size: 12px;
  font-family: var(--font-mono);
  opacity: 0;
  transition: opacity 0.2s;
  z-index: 100;
}
.gate-display:hover .debug-panel { opacity: 1; }

.debug-panel button {
  background: #555;
  color: #fff;
  border: none;
  padding: 4px 12px;
  border-radius: 2px;
  cursor: pointer;
  font-size: 12px;
}
.debug-panel button:hover { background: #666; }

/* Registration screen */
.registration-screen {
  height: 100vh;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 32px;
  background: var(--bg-primary);
  color: var(--text-primary);
}

.registration-screen h1 { font-size: 36px; letter-spacing: 8px; }

.registration-card {
  background: var(--bg-panel);
  border: 1px solid var(--border);
  padding: 32px;
  text-align: center;
  max-width: 480px;
}

.waiting-text {
  font-size: 18px;
  color: var(--amber);
  margin-bottom: 24px;
  animation: pulse 2s infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

.hint { margin-top: 24px; font-size: 12px; color: var(--text-muted); }
```

## 2.14 Tests — `gate-display/src/renderer/lib/gateMachine.test.ts`

```typescript
import { describe, it, expect } from 'vitest'
import { transition, getDisplay, type GateState, type ControllerEvent, type TimeoutAction } from './gateMachine'

function t(state: GateState, event: ControllerEvent | TimeoutAction) {
  return transition(state, event)
}

describe('gateMachine transitions', () => {
  it('starts in IDLE with correct display', () => {
    const d = getDisplay('IDLE')
    expect(d.gateState).toBe('CLOSED')
    expect(d.showRates).toBe(true)
    expect(d.loopIndicator).toBe('OFF')
    expect(d.ticketIndicator).toBe('OFF')
    expect(d.instructionText).toBe('')
  })

  it('LOOP_ON: IDLE → VEHICLE_DETECTED', () => {
    const r = t('IDLE', { type: 'LOOP_ON' })
    expect(r.state).toBe('VEHICLE_DETECTED')
    expect(r.command).toBeNull()
  })

  it('LOOP_OFF in VEHICLE_DETECTED returns to IDLE', () => {
    const r = t('VEHICLE_DETECTED', { type: 'LOOP_OFF' })
    expect(r.state).toBe('IDLE')
  })

  it('BUTTON_PRESSED: VEHICLE_DETECTED → TICKET_PRESSED + DISPENSE_TICKET', () => {
    const r = t('VEHICLE_DETECTED', { type: 'BUTTON_PRESSED' })
    expect(r.state).toBe('TICKET_PRESSED')
    expect(r.command).toEqual({ type: 'DISPENSE_TICKET' })
  })

  it('DISPENSE_OK: TICKET_PRESSED → TICKET_READY + ADVANCE_GATE timeout', () => {
    const r = t('TICKET_PRESSED', { type: 'DISPENSE_OK' })
    expect(r.state).toBe('TICKET_READY')
    expect(r.timeout).toBe('ADVANCE_GATE')
  })

  it('ADVANCE_GATE timeout: TICKET_READY → GATE_OPENING + BARRIER_OPEN', () => {
    const r = t('TICKET_READY', 'ADVANCE_GATE')
    expect(r.state).toBe('GATE_OPENING')
    expect(r.command).toEqual({ type: 'BARRIER_OPEN' })
    expect(r.timeout).toBe('ADVANCE_GATE_DONE')
  })

  it('ADVANCE_GATE_DONE timeout: GATE_OPENING → GATE_OPEN', () => {
    const r = t('GATE_OPENING', 'ADVANCE_GATE_DONE')
    expect(r.state).toBe('GATE_OPEN')
  })

  it('LOOP_OFF: GATE_OPEN → VEHICLE_EXITED + BARRIER_CLOSE', () => {
    const r = t('GATE_OPEN', { type: 'LOOP_OFF' })
    expect(r.state).toBe('VEHICLE_EXITED')
    expect(r.command).toEqual({ type: 'BARRIER_CLOSE' })
    expect(r.timeout).toBe('ADVANCE_EXIT')
  })

  it('ADVANCE_EXIT timeout: VEHICLE_EXITED → IDLE', () => {
    const r = t('VEHICLE_EXITED', 'ADVANCE_EXIT')
    expect(r.state).toBe('IDLE')
  })

  it('30s timeout: VEHICLE_DETECTED → IDLE (no action)', () => {
    const r = t('VEHICLE_DETECTED', 'RESET_IDLE')
    expect(r.state).toBe('IDLE')
  })

  it('unexpected events are ignored', () => {
    const r = t('GATE_OPEN', { type: 'BUTTON_PRESSED' })
    expect(r.state).toBe('GATE_OPEN')
    expect(r.command).toBeNull()
  })
})

describe('getDisplay per state', () => {
  const cases: Array<[GateState, string, string, string]> = [
    ['IDLE', 'OFF', 'OFF', 'CLOSED'],
    ['VEHICLE_DETECTED', 'AMBER', 'OFF', 'CLOSED'],
    ['TICKET_PRESSED', 'AMBER', 'GREEN', 'CLOSED'],
    ['TICKET_READY', 'AMBER', 'GREEN', 'CLOSED'],
    ['GATE_OPENING', 'OFF', 'OFF', 'OPENING'],
    ['GATE_OPEN', 'OFF', 'OFF', 'OPEN'],
    ['VEHICLE_EXITED', 'OFF', 'OFF', 'OPEN'],
  ]

  it.each(cases)('%s → loop=%s ticket=%s gate=%s', (state, loop, ticket, gate) => {
    const d = getDisplay(state)
    expect(d.loopIndicator).toBe(loop)
    expect(d.ticketIndicator).toBe(ticket)
    expect(d.gateState).toBe(gate)
  })
})
```

## 2.15 Manual verification

```bash
# Create a test config
echo '{"api_url":"http://localhost:8080","location_id":"<real-location-uuid>"}' > gate-display/config.json

# Build and run
cd gate-display && npm install && npm run dev
```

- Window opens fullscreen, dark background, frameless
- Location name appears in header (or "PARKIR" fallback if API unreachable)
- Clock ticks in header
- Camera placeholder (dark 4:3 rectangle with "CAMERA" text)
- Loop indicator and ticket indicator show mock state
- Instruction text changes per state
- Rates visible in IDLE (fetched from API)
- Gate barrier bar at bottom (red/amber/green)
- Debug panel appears on hover near bottom, buttons trigger mock events

### Tests

```bash
cd gate-display && npx vitest run
# 14+ tests pass
```
