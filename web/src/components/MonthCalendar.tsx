import { useState } from 'react'
import { todayYYYYMMDD } from '../utils/date'

export function MonthCalendar({ activityMap, selected, onSelect }: {
  activityMap: Map<string, number>
  selected: string
  onSelect: (date: string) => void
}) {
  const todayStr = todayYYYYMMDD()
  const [year, setYear] = useState(parseInt(selected.slice(0, 4)))
  const [month, setMonth] = useState(parseInt(selected.slice(4, 6)) - 1)

  const prevMonth = () => { if (month === 0) { setYear(y => y - 1); setMonth(11) } else setMonth(m => m - 1) }
  const nextMonth = () => { if (month === 11) { setYear(y => y + 1); setMonth(0) } else setMonth(m => m + 1) }

  const firstDay = new Date(year, month, 1).getDay()
  const daysInMonth = new Date(year, month + 1, 0).getDate()
  const cells: (number | null)[] = [...Array(firstDay).fill(null), ...Array.from({ length: daysInMonth }, (_, i) => i + 1)]

  return (
    <div className="border border-washi-header w-56">
      <div className="flex items-center justify-between px-2 py-1.5 border-b border-washi-border bg-washi-surface">
        <button onClick={prevMonth} className="font-mono text-sm px-2 py-0.5 hover:bg-washi-border transition-colors">←</button>
        <span className="font-mono text-sm font-bold">{year}/{String(month + 1).padStart(2, '0')}</span>
        <button onClick={nextMonth} className="font-mono text-sm px-2 py-0.5 hover:bg-washi-border transition-colors">→</button>
      </div>
      <div className="grid grid-cols-7 text-center">
        {['日','月','火','水','木','金','土'].map(d => (
          <div key={d} className="text-xs text-washi-muted py-1">{d}</div>
        ))}
        {cells.map((day, i) => {
          if (!day) return <div key={i} />
          const dateStr = `${year}${String(month + 1).padStart(2, '0')}${String(day).padStart(2, '0')}`
          const hasActivity = (activityMap.get(dateStr.slice(0,4) + '-' + dateStr.slice(4,6) + '-' + dateStr.slice(6,8)) ?? 0) > 0
          const isSelected = dateStr === selected
          const isFuture = dateStr > todayStr
          return (
            <button
              key={i}
              disabled={!hasActivity || isFuture}
              onClick={() => onSelect(dateStr)}
              className={`font-mono text-xs py-1 transition-colors
                ${isSelected ? 'bg-washi-header text-white' : ''}
                ${!isSelected && hasActivity ? 'bg-green-100 hover:bg-green-300 cursor-pointer font-bold' : ''}
                ${!hasActivity || isFuture ? 'text-washi-border cursor-default' : 'text-washi-text'}
              `}
            >
              {day}
            </button>
          )
        })}
      </div>
    </div>
  )
}
