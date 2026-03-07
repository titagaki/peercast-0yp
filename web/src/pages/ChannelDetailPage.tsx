import { useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import { api, ActivityDay, TimelineRow } from '../api'

function todayYYYYMMDD(): string {
  const d = new Date()
  return `${d.getFullYear()}${String(d.getMonth() + 1).padStart(2, '0')}${String(d.getDate()).padStart(2, '0')}`
}

function toYYYYMMDD(d: Date): string {
  return `${d.getFullYear()}${String(d.getMonth() + 1).padStart(2, '0')}${String(d.getDate()).padStart(2, '0')}`
}

function MonthCalendar({ activityMap, selected, onSelect }: {
  activityMap: Map<string, number>
  selected: string
  onSelect: (date: string) => void
}) {
  const today = new Date()
  const [year, setYear] = useState(
    parseInt(selected.slice(0, 4)),
  )
  const [month, setMonth] = useState(
    parseInt(selected.slice(4, 6)) - 1,
  )

  const prevMonth = () => { if (month === 0) { setYear(y => y - 1); setMonth(11) } else setMonth(m => m - 1) }
  const nextMonth = () => { if (month === 11) { setYear(y => y + 1); setMonth(0) } else setMonth(m => m + 1) }

  const firstDay = new Date(year, month, 1).getDay()
  const daysInMonth = new Date(year, month + 1, 0).getDate()
  const cells: (number | null)[] = [...Array(firstDay).fill(null), ...Array.from({ length: daysInMonth }, (_, i) => i + 1)]

  return (
    <div className="border border-gray-900 w-56">
      <div className="flex items-center justify-between px-2 py-1 border-b border-gray-900 bg-gray-50">
        <button onClick={prevMonth} className="font-mono text-base px-1 hover:bg-gray-200">←</button>
        <span className="font-mono text-base font-bold">{year}/{String(month + 1).padStart(2, '0')}</span>
        <button onClick={nextMonth} className="font-mono text-base px-1 hover:bg-gray-200">→</button>
      </div>
      <div className="grid grid-cols-7 text-center">
        {['日','月','火','水','木','金','土'].map(d => (
          <div key={d} className="font-mono text-base text-gray-400 py-0.5">{d}</div>
        ))}
        {cells.map((day, i) => {
          if (!day) return <div key={i} />
          const dateStr = `${year}${String(month + 1).padStart(2, '0')}${String(day).padStart(2, '0')}`
          const hasActivity = (activityMap.get(dateStr.slice(0,4) + '-' + dateStr.slice(4,6) + '-' + dateStr.slice(6,8)) ?? 0) > 0
          const isSelected = dateStr === selected
          const isFuture = new Date(year, month, day) > today
          return (
            <button
              key={i}
              disabled={!hasActivity || isFuture}
              onClick={() => onSelect(dateStr)}
              className={`font-mono text-base py-1 transition-colors
                ${isSelected ? 'bg-gray-900 text-white' : ''}
                ${!isSelected && hasActivity ? 'bg-green-100 hover:bg-green-300 cursor-pointer font-bold' : ''}
                ${!hasActivity || isFuture ? 'text-gray-300 cursor-default' : 'text-gray-900'}
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

function ActivityHeatmap({ data, onDateClick }: { data: ActivityDay[]; onDateClick: (date: string) => void }) {
  const map = new Map(data.map(d => [d.date, d.minutes]))
  const maxMin = Math.max(...Array.from(map.values()), 1)

  // Build 365 days ending today
  const today = new Date()
  const days: { date: string; minutes: number }[] = []
  for (let i = 364; i >= 0; i--) {
    const d = new Date(today)
    d.setDate(d.getDate() - i)
    const key = d.toISOString().slice(0, 10)
    days.push({ date: key, minutes: map.get(key) ?? 0 })
  }

  // Group into columns of 7 (weeks)
  const weeks: typeof days[] = []
  for (let i = 0; i < days.length; i += 7) {
    weeks.push(days.slice(i, i + 7))
  }

  function cellColor(minutes: number): string {
    if (minutes === 0) return 'bg-gray-100'
    const ratio = minutes / maxMin
    if (ratio < 0.25) return 'bg-green-200'
    if (ratio < 0.5) return 'bg-green-400'
    if (ratio < 0.75) return 'bg-green-600'
    return 'bg-green-800'
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
              className={`aspect-square rounded-sm ${cellColor(day.minutes)} ${day.minutes > 0 ? 'cursor-pointer hover:ring-1 hover:ring-gray-900' : ''}`}
            />
          ))}
        </div>
      ))}
    </div>
  )
}

export default function ChannelDetailPage() {
  const { name } = useParams<{ name: string }>()
  const channelName = decodeURIComponent(name ?? '')

  const [date, setDate] = useState(todayYYYYMMDD())
  const [activity, setActivity] = useState<ActivityDay[]>([])
  const activityMap = new Map(activity.map(d => [d.date, d.minutes]))
  const [timeline, setTimeline] = useState<TimelineRow[]>([])
  const [loadingTimeline, setLoadingTimeline] = useState(false)

  useEffect(() => {
    if (!channelName) return
    api.activity(channelName).then(setActivity).catch(() => {})
  }, [channelName])

  useEffect(() => {
    if (!channelName) return
    setLoadingTimeline(true)
    api.timeline(channelName, date)
      .then(setTimeline)
      .catch(() => setTimeline([]))
      .finally(() => setLoadingTimeline(false))
  }, [channelName, date])

  function fmtTime(iso: string): string {
    return new Date(iso).toLocaleTimeString('ja-JP', { hour: '2-digit', minute: '2-digit' })
  }

  // Only show detail text on changed rows
  const rows = timeline.map(row => ({
    ...row,
    detail: row.changed ? (row.trackTitle || row.description || '') : '',
  }))

  return (
    <div className="space-y-8">
      <div className="border-b-2 border-gray-900 pb-2">
        <h1 className="font-black text-gray-900 text-xl tracking-tight">{channelName}</h1>
      </div>

      <section>
        <div className="border-b-2 border-gray-900 pb-3 mb-3 space-y-2">
          <h2 className="font-mono text-xl uppercase tracking-widest text-gray-900 font-bold">
            Timeline
          </h2>
          <MonthCalendar activityMap={activityMap} selected={date} onSelect={setDate} />
        </div>

        {loadingTimeline ? (
          <p className="text-gray-400 text-xl font-mono">loading...</p>
        ) : rows.length === 0 ? (
          <p className="text-gray-400 text-xl font-mono">データがありません。</p>
        ) : (
          <table className="w-full text-xl border border-gray-900">
            <thead>
              <tr className="text-left border-b-2 border-gray-900 bg-gray-50">
                <th className="py-2 px-3 font-mono text-base uppercase tracking-wider text-gray-600">時刻</th>
                <th className="py-2 px-3 font-mono text-base uppercase tracking-wider text-gray-600">リスナー/リレー</th>
                <th className="py-2 px-3 font-mono text-base uppercase tracking-wider text-gray-600">詳細</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200">
              {rows.map((row, i) => (
                <tr
                  key={i}
                  className={`hover:bg-gray-50 transition-colors ${row.changed ? '' : 'opacity-40'}`}
                >
                  <td className="py-1.5 px-3 font-mono text-base tabular-nums">{fmtTime(row.recordedAt)}</td>
                  <td className="py-1.5 px-3 font-mono text-base tabular-nums">{row.listeners}/{row.relays}</td>
                  <td className="py-1.5 px-3 text-base text-gray-700">{row.detail}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </section>

      <section>
        <h2 className="font-mono text-xl uppercase tracking-widest text-gray-500 mb-3">
          過去365日の放送
        </h2>
        <ActivityHeatmap data={activity} onDateClick={setDate} />
      </section>
    </div>
  )
}
