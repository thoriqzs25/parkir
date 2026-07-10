interface Props {
  state: 'OFF' | 'GREEN'
}

export default function TicketButton({ state }: Props) {
  return (
    <div className="indicator">
      <span className={`indicator-dot ${state === 'GREEN' ? 'green' : 'off'}`} />
      <span>Tiket: {state === 'GREEN' ? 'DITEKAN' : '—'}</span>
    </div>
  )
}
