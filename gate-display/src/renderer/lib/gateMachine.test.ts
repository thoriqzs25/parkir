import { describe, it, expect } from 'vitest'
import { transition, getDisplay, type GateState, type ControllerEvent, type TimeoutAction } from './gateMachine'

function t(state: GateState, event: ControllerEvent | TimeoutAction) {
  return transition(state, event)
}

describe('gateMachine transitions', () => {
  it('IDLE: default display', () => {
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

  it('ADVANCE_GATE: TICKET_READY → GATE_OPENING + BARRIER_OPEN', () => {
    const r = t('TICKET_READY', 'ADVANCE_GATE')
    expect(r.state).toBe('GATE_OPENING')
    expect(r.command).toEqual({ type: 'BARRIER_OPEN' })
    expect(r.timeout).toBe('ADVANCE_GATE_DONE')
  })

  it('ADVANCE_GATE_DONE: GATE_OPENING → GATE_OPEN', () => {
    const r = t('GATE_OPENING', 'ADVANCE_GATE_DONE')
    expect(r.state).toBe('GATE_OPEN')
  })

  it('LOOP_OFF: GATE_OPEN → VEHICLE_EXITED + BARRIER_CLOSE', () => {
    const r = t('GATE_OPEN', { type: 'LOOP_OFF' })
    expect(r.state).toBe('VEHICLE_EXITED')
    expect(r.command).toEqual({ type: 'BARRIER_CLOSE' })
    expect(r.timeout).toBe('ADVANCE_EXIT')
  })

  it('ADVANCE_EXIT: VEHICLE_EXITED → IDLE', () => {
    const r = t('VEHICLE_EXITED', 'ADVANCE_EXIT')
    expect(r.state).toBe('IDLE')
  })

  it('RESET_IDLE: VEHICLE_DETECTED → IDLE (30s timeout)', () => {
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
