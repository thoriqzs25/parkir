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
