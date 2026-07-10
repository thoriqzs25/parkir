import { app, BrowserWindow, ipcMain } from 'electron'
import * as path from 'path'
import * as fs from 'fs'
import * as os from 'os'
import * as http from 'http'
import * as crypto from 'crypto'
import makeMdns = require('multicast-dns')

const CONFIG_PATH = path.join(__dirname, '..', 'config.json')
const APP_DATA_DIR = path.join(app.getPath('userData'), 'parkir-gate')
const DEVICE_ID_PATH = path.join(APP_DATA_DIR, 'device-id')

let mainWindow: BrowserWindow | null = null

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

const deviceId = loadOrCreateDeviceId()

function startMDNS() {
  try {
    const mdns = makeMdns()

    mdns.on('query', (query: any) => {
      const hasQuestion = query.questions?.some(
        (q: any) => q.name === '_parkir-gate._tcp.local',
      )
      if (!hasQuestion) return

      const localIP = getLocalIP()
      mdns.respond({
        answers: [
          {
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
          },
          {
            name: '_parkir-gate._tcp.local',
            type: 'TXT',
            class: 'IN',
            ttl: 120,
            data: Buffer.from(`device_id=${deviceId}`),
          },
          {
            name: os.hostname() + '.local',
            type: 'A',
            class: 'IN',
            ttl: 120,
            data: localIP,
          },
        ],
      })
    })

    app.on('will-quit', () => {
      mdns.destroy()
    })
  } catch {
    console.warn('mDNS not available')
  }
}

function startHTTPServer() {
  const server = http.createServer((req, res) => {
    res.setHeader('Access-Control-Allow-Origin', '*')
    res.setHeader('Access-Control-Allow-Methods', 'GET, POST, OPTIONS')
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
            res.writeHead(400, { 'Content-Type': 'application/json' })
            res.end(JSON.stringify({ error: 'location_id and api_url required' }))
            return
          }

          const config = { location_id, api_url }
          fs.writeFileSync(CONFIG_PATH, JSON.stringify(config, null, 2))

          res.writeHead(200, { 'Content-Type': 'application/json' })
          res.end(JSON.stringify({ status: 'ok' }))

          app.relaunch()
          app.exit()
        } catch {
          res.writeHead(400, { 'Content-Type': 'application/json' })
          res.end(JSON.stringify({ error: 'invalid JSON' }))
        }
      })
      return
    }

    if (req.method === 'GET' && req.url === '/info') {
      res.writeHead(200, { 'Content-Type': 'application/json' })
      res.end(
        JSON.stringify({
          device_id: deviceId,
          ip: getLocalIP(),
          hostname: os.hostname(),
          registered: fs.existsSync(CONFIG_PATH),
        }),
      )
      return
    }

    res.writeHead(404)
    res.end('not found')
  })

  server.listen(9800, () => {
    console.log(`gate registration server on port 9800 (device: ${deviceId})`)
  })

  app.on('will-quit', () => {
    server.close()
  })
}

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

app.whenReady().then(() => {
  startMDNS()
  startHTTPServer()
  createWindow()
})

app.on('window-all-closed', () => {
  if (process.platform !== 'darwin') app.quit()
})

app.on('activate', () => {
  if (mainWindow === null) createWindow()
})

ipcMain.handle('read-config', () => {
  try {
    const raw = fs.readFileSync(CONFIG_PATH, 'utf-8')
    return JSON.parse(raw)
  } catch {
    return null
  }
})

ipcMain.handle('write-config', (_event, config: Record<string, unknown>) => {
  fs.writeFileSync(CONFIG_PATH, JSON.stringify(config, null, 2))
  return true
})

ipcMain.handle('quit-app', () => { app.quit() })

ipcMain.handle('get-device-id', () => deviceId)

ipcMain.handle('get-ip', () => getLocalIP())
