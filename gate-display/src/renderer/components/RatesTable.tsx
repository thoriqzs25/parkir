interface RateRow {
  vehicle_type: string
  first_hour_rate: number
  subsequent_hourly_rate: number
  daily_flat_rate: number
}

interface Props {
  rates: RateRow[]
  visible: boolean
}

export default function RatesTable({ rates, visible }: Props) {
  if (!visible || rates.length === 0) return null

  return (
    <div className="rates-panel">
      <div className="rates-title">TARIF</div>
      <table className="rates-table">
        <tbody>
          {rates.map((r) => (
            <tr key={r.vehicle_type}>
              <td>{r.vehicle_type}</td>
              <td>
                Rp {r.first_hour_rate.toLocaleString('id-ID')} / jam
              </td>
              <td>
                Rp {r.daily_flat_rate.toLocaleString('id-ID')} / hari
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}
