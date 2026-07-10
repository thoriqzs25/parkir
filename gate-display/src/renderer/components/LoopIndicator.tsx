interface Props {
  state: 'OFF' | 'AMBER'
}

export default function LoopIndicator({ state }: Props) {
  return (
    <div className="indicator">
      <span className={`indicator-dot ${state === 'AMBER' ? 'amber' : 'off'}`} />
      <span>Loop: {state === 'AMBER' ? 'TERDETEKSI' : '—'}</span>
    </div>
  )
}
