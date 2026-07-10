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
    window.electronAPI.readConfig().then((cfg: Record<string, unknown> | null) => {
      if (cfg && typeof cfg.api_url === 'string' && typeof cfg.location_id === 'string') {
        setConfig(cfg as unknown as Config)
      }
      setLoading(false)
    })
  }, [])

  if (loading) return null
  if (!config) return <RegistrationScreen />
  return <GateDisplay apiUrl={config.api_url} locationId={config.location_id} />
}
