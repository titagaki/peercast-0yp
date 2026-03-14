import { useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import { api, ActivityDay, Channel, TimelineRow } from '../api'

function todayYYYYMMDD(): string {
  // sv-SE locale produces YYYY-MM-DD; use Asia/Tokyo to avoid UTC date shift
  return new Date().toLocaleDateString('sv-SE', { timeZone: 'Asia/Tokyo' }).replace(/-/g, '')
}

function formatDateDisplay(yyyymmdd: string): string {
  const y = yyyymmdd.slice(0, 4)
  const m = parseInt(yyyymmdd.slice(4, 6))
  const d = parseInt(yyyymmdd.slice(6, 8))
  return `${y}年${m}月${d}日`
}

function MonthCalendar({ activityMap, selected, onSelect }: {
  activityMap: Map<string, number>
  selected: string
  onSelect: (date: string) => void
}) {
  const todayStr = todayYYYYMMDD()
  const today = new Date(todayStr.slice(0,4) + '-' + todayStr.slice(4,6) + '-' + todayStr.slice(6,8) + 'T00:00:00+09:00')
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
          const isFuture = new Date(year, month, day) > today
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

function ActivityHeatmap({ data, onDateClick }: { data: ActivityDay[]; onDateClick: (date: string) => void }) {
  const map = new Map(data.map(d => [d.date, d.minutes]))
  const maxMin = Math.max(...Array.from(map.values()), 1)

  const todayStr = todayYYYYMMDD()
  const today = new Date(todayStr.slice(0,4) + '-' + todayStr.slice(4,6) + '-' + todayStr.slice(6,8) + 'T00:00:00+09:00')
  const days: { date: string; minutes: number }[] = []
  for (let i = 364; i >= 0; i--) {
    const d = new Date(today)
    d.setDate(d.getDate() - i)
    const key = d.toLocaleDateString('sv-SE', { timeZone: 'Asia/Tokyo' })
    days.push({ date: key, minutes: map.get(key) ?? 0 })
  }

  const weeks: typeof days[] = []
  for (let i = 0; i < days.length; i += 7) {
    weeks.push(days.slice(i, i + 7))
  }

  function cellColor(minutes: number): string {
    if (minutes === 0) return 'bg-washi-surface'
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
              className={`aspect-square rounded-sm ${cellColor(day.minutes)} ${day.minutes > 0 ? 'cursor-pointer hover:ring-1 hover:ring-washi-header' : ''}`}
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
  const [liveChannel, setLiveChannel] = useState<Channel | null>(null)

  useEffect(() => {
    if (!channelName) return
    api.activity(channelName).then(setActivity).catch(() => {})
    api.channels().then(chs => {
      setLiveChannel(chs.find(c => c.name === channelName) ?? null)
    }).catch(() => {})
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

  const rows = timeline.map(row => ({
    ...row,
    detail: row.changed ? [[row.genre, row.trackTitle || row.description].filter(Boolean).join(' - '), row.comment].filter(Boolean).join(' - ') : '',
  }))

  return (
    <div className="space-y-8">
      <div className="border-b-2 border-washi-header pb-3">
        <h1 className="font-black text-washi-text text-xl tracking-tight">{channelName}</h1>
        {liveChannel?.comment && (
          <p className="mt-1 text-sm text-washi-text">{liveChannel.comment}</p>
        )}
      </div>

      <section>
        <div className="border-b-2 border-washi-header pb-3 mb-4 space-y-3">
          <div className="flex items-baseline gap-3 flex-wrap">
            <h2 className="font-black text-xl uppercase tracking-tight text-washi-text">Timeline</h2>
            <span className="font-mono text-base font-bold text-washi-header">{formatDateDisplay(date)}</span>
            {date !== todayYYYYMMDD() && (
              <button
                onClick={() => setDate(todayYYYYMMDD())}
                className="font-mono text-xs px-2 py-0.5 border border-washi-header text-washi-header hover:bg-washi-header hover:text-white transition-colors"
              >
                今日
              </button>
            )}
          </div>
          <MonthCalendar activityMap={activityMap} selected={date} onSelect={setDate} />
        </div>

        {loadingTimeline ? (
          <p className="text-washi-muted text-base">読み込み中...</p>
        ) : rows.length === 0 ? (
          <p className="text-washi-muted text-base">データがありません。</p>
        ) : (
          <table className="w-full border border-washi-border">
            <thead>
              <tr className="text-left border-b-2 border-washi-header bg-washi-surface">
                <th className="py-2 px-4 text-xs font-bold uppercase tracking-wider text-washi-muted">時刻</th>
                <th className="py-2 px-4 text-xs font-bold uppercase tracking-wider text-washi-muted">リスナー/リレー</th>
                <th className="py-2 px-4 text-xs font-bold uppercase tracking-wider text-washi-muted">詳細</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-washi-border">
              {rows.map((row, i) => (
                <tr
                  key={i}
                  className={`hover:bg-washi-surface transition-colors ${row.changed ? '' : 'opacity-40'}`}
                >
                  <td className="py-2 px-4 font-mono text-sm tabular-nums">{fmtTime(row.recordedAt)}</td>
                  <td className="py-2 px-4 font-mono text-sm tabular-nums">{row.listeners}/{row.relays}</td>
                  <td className="py-2 px-4 text-sm text-washi-text">{row.detail}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </section>

      <section>
        <h2 className="font-bold text-sm uppercase tracking-wider text-washi-muted mb-3">
          過去365日の配信
        </h2>
        <ActivityHeatmap data={activity} onDateClick={setDate} />
      </section>
    </div>
  )
}
