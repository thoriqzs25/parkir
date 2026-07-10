import type { ControllerEvent, ControllerCommand } from './gateMachine'

export type { ControllerEvent, ControllerCommand, TimeoutAction } from './gateMachine'

export interface ControllerInterface {
  onEvent: (cb: (event: ControllerEvent) => void) => void
  sendCommand: (cmd: ControllerCommand) => Promise<void>
  connect: () => Promise<void>
  disconnect: () => void
}

export class MockController implements ControllerInterface {
  private listeners: Array<(event: ControllerEvent) => void> = []

  connect() {
    return Promise.resolve()
  }

  disconnect() {
    this.listeners = []
  }

  onEvent(cb: (event: ControllerEvent) => void) {
    this.listeners.push(cb)
  }

  sendCommand(cmd: ControllerCommand) {
    console.log('[MOCK CONTROLLER] command:', cmd)
    return Promise.resolve()
  }

  trigger(event: ControllerEvent) {
    this.listeners.forEach((cb) => cb(event))
  }
}
