import { useEffect, useState, useMemo } from 'react'
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
        const res = await fetch(
          `${apiUrl}/api/v1/gate/${encodeURIComponent(locationId)}/info`,
        )
        if (!res.ok) throw new Error('fetch failed')
        const json = await res.json()
        const info: GateInfo = json.data
        setGateInfo(info)
        setLocationName(info.location.name)
      } catch {
        // keep fallback
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
