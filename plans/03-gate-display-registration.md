# Milestone 3: Gate Display App — Registration Flow

## Objective

Add LAN discovery (mDNS), registration HTTP server, and the registration screen so the gate display can be configured from the desktop app.

## Files

| # | File | Action |
|---|------|--------|
| 1 | `gate-display/src/main.ts` | **Edit** — add mDNS + HTTP server on port 9800 |
| 2 | `gate-display/src/preload.ts` | **Edit** — add `getDeviceId`, `getIP` |
| 3 | `gate-display/src/renderer/screens/RegistrationScreen.tsx` | **Edit** — show device ID + IP, replace placeholder |

## Dependency

Requires `multicast-dns` npm package:
```
cd gate-display && npm install multicast-dns
cd gate-display && npm install --save-dev @types/multicast-dns
```

## 3.1 Device identity

A persistent device ID is generated once and stored in the Electron app data directory.

```typescript
// In main.ts
import * as os from 'os'
import { v4 as uuidv4 } from 'uuid'
// Or use crypto.randomUUID()

const APP_DATA_DIR = path.join(app.getPath('userData'), 'gate-data')
const DEVICE_ID_PATH = path.join(APP_DATA_DIR, 'device-id')

function loadOrCreateDeviceId(): string {
  try {
    return fs.readFileSync(DEVICE_ID_PATH, 'utf-8').trim()
  } catch {
    const id = crypto.randomUUID()
    fs.mkdirSync(APP_DATA_DIR, { recursive: true })
    fs.writeFileSync(DEVICE_ID_PATH, id)
    return id
  }
}
```

## 3.2 mDNS announcement

```typescript
import * as multicastdns from 'multicast-dns'
import * as os from 'os'

const mdns = multicastdns()
const deviceId = loadOrCreateDeviceId()

function getLocalIP(): string {
  const interfaces = os.networkInterfaces()
  for (const iface of Object.values(interfaces)) {
    if (!iface) continue
    for (const addr of iface) {
      if (addr.family === 'IPv4' && !addr.internal) return addr.address
    }
  }
  return '127.0.0.1'
}

mdns.on('query', (query) => {
  if (query.questions?.some(q => q.name === '_parkir-gate._tcp.local')) {
    mdns.respond({
      answers: [{
        name: '_parkir-gate._tcp.local',
        type: 'SRV',
        class: 'IN',
        ttl: 120,
        data: {
          priority: 0,
          weight: 0,
          port: 9800,
          target: os.hostname() + '.local',
        },
      }, {
        name: '_parkir-gate._tcp.local',
        type: 'TXT',
        class: 'IN',
        ttl: 120,
        data: Buffer.from(`device_id=${deviceId}`),
      }, {
        name: os.hostname() + '.local',
        type: 'A',
        class: 'IN',
        ttl: 120,
        data: getLocalIP(),
      }],
    })
  }
})
```

## 3.3 HTTP registration server

Built into main.ts using Node's built-in `http` module.

```typescript
import * as http from 'http'

const CONFIG_PATH = path.join(__dirname, '..', 'config.json')

const server = http.createServer((req, res) => {
  // CORS for desktop app
  res.setHeader('Access-Control-Allow-Origin', '*')
  res.setHeader('Access-Control-Allow-Methods', 'POST, OPTIONS')
  res.setHeader('Access-Control-Allow-Headers', 'Content-Type')

  if (req.method === 'OPTIONS') {
    res.writeHead(204)
    res.end()
    return
  }

  if (req.method === 'POST' && req.url === '/register') {
    let body = ''
    req.on('data', (chunk) => (body += chunk))
    req.on('end', () => {
      try {
        const { location_id, api_url } = JSON.parse(body)
        if (!location_id || !api_url) {
          res.writeHead(400)
          res.end(JSON.stringify({ error: 'location_id and api_url required' }))
          return
        }

        const config = { location_id, api_url }
        fs.writeFileSync(CONFIG_PATH, JSON.stringify(config, null, 2))

        res.writeHead(200, { 'Content-Type': 'application/json' })
        res.end(JSON.stringify({ status: 'ok' }))

        // Restart into display mode
        app.relaunch()
        app.exit()
      } catch {
        res.writeHead(400)
        res.end(JSON.stringify({ error: 'invalid JSON' }))
      }
    })
    return
  }

  // GET /info — for desktop app to discover
  if (req.method === 'GET' && req.url === '/info') {
    res.writeHead(200, { 'Content-Type': 'application/json' })
    res.end(JSON.stringify({
      device_id: deviceId,
      ip: getLocalIP(),
      hostname: os.hostname(),
      registered: fs.existsSync(CONFIG_PATH),
    }))
    return
  }

  res.writeHead(404)
  res.end('not found')
})

server.listen(9800, () => {
  console.log(`Gate registration server on port 9800 (device: ${deviceId})`)
})

app.on('will-quit', () => {
  server.close()
  mdns.destroy()
})
```

## 3.4 Preload additions

```typescript
import { contextBridge, ipcRenderer } from 'electron'

contextBridge.exposeInMainWorld('electronAPI', {
  readConfig: () => ipcRenderer.invoke('read-config'),
  writeConfig: (config: unknown) => ipcRenderer.invoke('write-config', config),
  quitApp: () => ipcRenderer.invoke('quit-app'),
  getDeviceId: () => ipcRenderer.invoke('get-device-id'),
  getIP: () => ipcRenderer.invoke('get-ip'),
})
```

IPC handlers in main.ts:

```typescript
ipcMain.handle('get-device-id', () => deviceId)
ipcMain.handle('get-ip', () => getLocalIP())
```

## 3.5 Registration screen — updated

Replace the placeholder with the full version:

```tsx
import { useEffect, useState } from 'react'

export default function RegistrationScreen() {
  const [deviceId, setDeviceId] = useState('')
  const [ip, setIp] = useState('')

  useEffect(() => {
    window.electronAPI.getDeviceId().then(setDeviceId)
    window.electronAPI.getIP().then(setIp)
  }, [])

  return (
    <div className="registration-screen">
      <h1>PARKIR GATE</h1>
      <div className="registration-card">
        <p className="waiting-text">Menunggu registrasi...</p>
        <div className="info-row">
          <span>Device ID</span>
          <span className="mono">{deviceId}</span>
        </div>
        <div className="info-row">
          <span>IP Address</span>
          <span className="mono">{ip}</span>
        </div>
        <p className="hint">
          Daftarkan gate ini dari aplikasi PARKIR Desktop atau Dashboard.
        </p>
      </div>
    </div>
  )
}
```

CSS additions to App.css:

```css
.info-row {
  display: flex;
  justify-content: space-between;
  gap: 24px;
  padding: 8px 0;
  font-size: 14px;
  border-bottom: 1px solid var(--border);
  font-family: var(--font-sans);
}

.info-row .mono { font-family: var(--font-mono); color: var(--text-muted); }
```

## 3.6 Manual verification

```bash
cd gate-display && npm install multicast-dns && npm run dev
```

**Step 1:** Registration screen shows with device ID + IP (config.json deleted)

**Step 2:** From another terminal:
```bash
curl http://localhost:9800/info
# → { "device_id": "...", "ip": "192.168.1.x", "hostname": "...", "registered": false }

curl -X POST http://localhost:9800/register \
  -H "Content-Type: application/json" \
  -d '{"location_id":"<real-uuid>","api_url":"http://localhost:8080"}'
# → { "status": "ok" }
# Gate display restarts into display mode
```

**Step 3:** Verify `config.json` was created with correct content.

**Step 4:** Delete `config.json`, restart → back to registration screen.
