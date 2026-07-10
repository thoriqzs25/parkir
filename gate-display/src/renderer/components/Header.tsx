import { useEffect, useState } from 'react'

interface Props {
  locationName: string
}

export default function Header({ locationName }: Props) {
  const [time, setTime] = useState(new Date())

  useEffect(() => {
    const id = setInterval(() => setTime(new Date()), 1000)
    return () => clearInterval(id)
  }, [])

  const fmt = time.toLocaleDateString('id-ID', {
    weekday: 'short',
    day: '2-digit',
    month: 'short',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    timeZone: 'Asia/Jakarta',
  })

  return (
    <div className="header">
      <span className="location-name">{locationName}</span>
      <span className="clock">{fmt}</span>
    </div>
  )
}
