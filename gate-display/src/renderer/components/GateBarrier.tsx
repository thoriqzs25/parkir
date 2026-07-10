interface Props {
  state: 'CLOSED' | 'OPENING' | 'OPEN'
}

const LABELS: Record<string, string> = {
  CLOSED: 'TERTUTUP',
  OPENING: 'MEMBUKA',
  OPEN: 'TERBUKA',
}

export default function GateBarrier({ state }: Props) {
  return <div className={`gate-barrier ${state.toLowerCase()}`}>{LABELS[state]}</div>
}
