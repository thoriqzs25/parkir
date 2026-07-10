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

export function transition(
  current: GateState,
  event: ControllerEvent | TimeoutAction,
): GateMachineResult {
  const key = `${current}__${typeof event === 'string' ? event : event.type}`

  switch (key) {
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

    case 'VEHICLE_DETECTED__RESET_IDLE':
      return { state: 'IDLE', command: null, timeout: null }

    case 'TICKET_READY__ADVANCE_GATE':
      return { state: 'GATE_OPENING', command: { type: 'BARRIER_OPEN' }, timeout: 'ADVANCE_GATE_DONE' }

    case 'GATE_OPENING__ADVANCE_GATE_DONE':
      return { state: 'GATE_OPEN', command: null, timeout: null }

    case 'VEHICLE_EXITED__ADVANCE_EXIT':
      return { state: 'IDLE', command: null, timeout: null }

    default:
      return { state: current, command: null, timeout: null }
  }
}

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
