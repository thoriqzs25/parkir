import { useEffect, useRef, useState, useCallback } from 'react'
import type { ControllerInterface, ControllerEvent, TimeoutAction } from '../lib/controller'
import { transition, getDisplay, type DisplayState, type GateState } from '../lib/gateMachine'

const TIMEOUT_MS: Record<string, number> = {
  RESET_IDLE: 30000,
  ADVANCE_GATE: 3000,
  ADVANCE_GATE_DONE: 2000,
  ADVANCE_EXIT: 2000,
}

export function useGateState(controller: ControllerInterface): DisplayState {
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
      clearTimer()
      setCurrentState((prev) => {
        const result = transition(prev, event)

        if (result.command) {
          controller.sendCommand(result.command)
        }

        if (result.timeout) {
          const ms = TIMEOUT_MS[result.timeout]
          const action = result.timeout as TimeoutAction
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
