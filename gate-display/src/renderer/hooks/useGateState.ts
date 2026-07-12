import { useEffect, useRef, useState, useCallback } from 'react'
import type { ControllerInterface, ControllerEvent, TimeoutAction } from '../lib/controller'
import { transition, getDisplay, type DisplayState, type GateState } from '../lib/gateMachine'

const TIMEOUT_MS: Record<string, number> = {
  RESET_IDLE: 30000,
  ADVANCE_GATE: 3000,
  ADVANCE_GATE_DONE: 2000,
  ADVANCE_EXIT: 2000,
}

export function useGateState(controller: ControllerInterface, onTransition?: (msg: string) => void): DisplayState {
  const [currentState, setCurrentState] = useState<GateState>('IDLE')
  const [display, setDisplay] = useState<DisplayState>(getDisplay('IDLE'))
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  const clearTimer = useCallback(() => {
    if (timerRef.current !== null) {
      clearTimeout(timerRef.current)
      timerRef.current = null
    }
  }, [])

  const dispatch = useCallback(
    (event: ControllerEvent | TimeoutAction) => {
      const eventLabel = typeof event === 'string' ? event : event.type
      clearTimer()
      setCurrentState((prev) => {
        const result = transition(prev, event)
        const msg = `[GATE] ${prev} → ${eventLabel} → ${result.state}`
        console.log(msg)
        onTransition?.(msg)

        if (result.command) {
          const cmdMsg = `[GATE] command: ${result.command.type}`
          console.log(cmdMsg)
          onTransition?.(cmdMsg)
          controller.sendCommand(result.command)
        }

        if (result.timeout) {
          const ms = TIMEOUT_MS[result.timeout]
          const action = result.timeout as TimeoutAction
          const tmMsg = `[GATE] timeout ${result.timeout} in ${ms}ms`
          console.log(tmMsg)
          onTransition?.(tmMsg)
          timerRef.current = setTimeout(() => dispatch(action), ms)
        }

        setDisplay(getDisplay(result.state))
        return result.state
      })
    },
    [controller, clearTimer],
  )

  useEffect(() => {
    controller.onEvent((event) => dispatch(event))
    return () => clearTimer()
  }, [controller, dispatch, clearTimer])

  return display
}
