import { ActivityDay } from '../api'

function cellColor(minutes: number, maxMin: number): string {
  if (minutes === 0) return 'bg-washi-surface'
  const ratio = minutes / maxMin
  if (ratio < 0.25) return 'bg-green-200'
  if (ratio < 0.5) return 'bg-green-400'
  if (ratio < 0.75) return 'bg-green-600'
  return 'bg-green-800'
}

export function ActivityHeatmap({ data, onDateClick }: { data: ActivityDay[]; onDateClick: (date: string) => void }) {
  const map = new Map(data.map(d => [d.date, d.minutes]))
  const maxMin = Math.max(...Array.from(map.values()), 1)

  const days: { date: string; minutes: number }[] = []
  for (let i = 364; i >= 0; i--) {
    const ms = Date.now() - i * 86400000
    const key = new Date(ms).toLocaleDateString('sv-SE', { timeZone: 'Asia/Tokyo' })
    days.push({ date: key, minutes: map.get(key) ?? 0 })
  }

  const weeks: typeof days[] = []
  for (let i = 0; i < days.length; i += 7) {
    weeks.push(days.slice(i, i + 7))
  }

  return (
    <div className="flex gap-0.5 pb-1">
      {weeks.map((week, wi) => (
        <div key={wi} className="flex flex-col gap-0.5 flex-1">
          {week.map(day => (
            <div
              key={day.date}
              title={`${day.date}  ${day.minutes}分`}
              onClick={() => day.minutes > 0 && onDateClick(day.date.replace(/-/g, ''))}
              className={`aspect-square rounded-sm ${cellColor(day.minutes, maxMin)} ${day.minutes > 0 ? 'cursor-pointer hover:ring-1 hover:ring-washi-header' : ''}`}
            />
          ))}
        </div>
      ))}
    </div>
  )
}
