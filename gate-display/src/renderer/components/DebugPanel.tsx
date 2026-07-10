import type { MockController, ControllerEvent } from '../lib/controller'

interface Props {
  controller: MockController
}

export default function DebugPanel({ controller }: Props) {
  const trigger = (event: ControllerEvent) => {
    controller.trigger(event)
  }

  return (
    <div className="debug-panel">
      <button onClick={() => trigger({ type: 'LOOP_ON' })}>Loop ON</button>
      <button onClick={() => trigger({ type: 'LOOP_OFF' })}>Loop OFF</button>
      <button onClick={() => trigger({ type: 'BUTTON_PRESSED' })}>Ticket</button>
      <button onClick={() => trigger({ type: 'DISPENSE_OK' })}>Dispense OK</button>
    </div>
  )
}
