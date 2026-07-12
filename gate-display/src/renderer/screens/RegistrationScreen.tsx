import { useEffect, useState } from 'react'

interface Props {
  onBypass: (apiUrl: string, locationId: string) => void
}

const VERSION = 'v0.1.0'

export default function RegistrationScreen({ onBypass }: Props) {
  const [deviceId, setDeviceId] = useState('')
  const [ip, setIp] = useState('')
  const [clickCount, setClickCount] = useState(0)

  useEffect(() => {
    window.electronAPI.getDeviceId().then(setDeviceId)
    window.electronAPI.getIP().then(setIp)
  }, [])

  const handleVersionClick = () => {
    const next = clickCount + 1
    setClickCount(next)
    if (next >= 5) {
      onBypass('http://localhost:8080', 'bypass')
    }
  }

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
      <span className="version-text" onClick={handleVersionClick}>
        {VERSION}
      </span>
    </div>
  )
}
