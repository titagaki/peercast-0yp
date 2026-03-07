import { useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import { api, ActivityDay, TimelineRow } from '../api'

function todayYYYYMMDD(): string {
  const d = new Date()
  return `${d.getFullYear()}${String(d.getMonth() + 1).padStart(2, '0')}${String(d.getDate()).padStart(2, '0')}`
}

function yyyymmddToInputValue(s: string): string {
  return `${s.slice(0, 4)}-${s.slice(4, 6)}-${s.slice(6, 8)}`
}

function inputValueToYYYYMMDD(s: string): string {
  return s.replace(/-/g, '')
}

function ActivityHeatmap({ data }: { data: ActivityDay[] }) {
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
    if (ratio < 0.25) return 'bg-blue-200'
    if (ratio < 0.5) return 'bg-blue-400'
    if (ratio < 0.75) return 'bg-blue-600'
    return 'bg-blue-800'
  }

  return (
    <div className="flex gap-0.5 overflow-x-auto pb-1">
      {weeks.map((week, wi) => (
        <div key={wi} className="flex flex-col gap-0.5">
          {week.map(day => (
            <div
              key={day.date}
              title={`${day.date}  ${day.minutes}分`}
              className={`w-3 h-3 rounded-sm ${cellColor(day.minutes)}`}
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
      <h1 className="text-xl font-bold text-gray-900">{channelName}</h1>

      <section>
        <h2 className="text-sm font-semibold text-gray-500 uppercase tracking-wide mb-3">
          過去365日の放送
        </h2>
        <ActivityHeatmap data={activity} />
      </section>

      <section>
        <div className="flex items-center gap-3 mb-3">
          <h2 className="text-sm font-semibold text-gray-500 uppercase tracking-wide">
            タイムライン
          </h2>
          <input
            type="date"
            value={yyyymmddToInputValue(date)}
            onChange={e => setDate(inputValueToYYYYMMDD(e.target.value))}
            className="text-sm border border-gray-300 rounded px-2 py-0.5"
          />
        </div>

        {loadingTimeline ? (
          <p className="text-gray-400 text-sm">読み込み中...</p>
        ) : rows.length === 0 ? (
          <p className="text-gray-400 text-sm">データがありません。</p>
        ) : (
          <table className="w-full text-sm">
            <thead>
              <tr className="text-left text-gray-500 border-b border-gray-200">
                <th className="py-1 pr-4 font-medium">時刻</th>
                <th className="py-1 pr-4 font-medium">リスナー</th>
                <th className="py-1 font-medium">詳細</th>
              </tr>
            </thead>
            <tbody>
              {rows.map((row, i) => (
                <tr
                  key={i}
                  className={`border-b border-gray-50 ${row.changed ? '' : 'text-gray-400'}`}
                >
                  <td className="py-0.5 pr-4 tabular-nums">{fmtTime(row.recordedAt)}</td>
                  <td className="py-0.5 pr-4 tabular-nums">{row.listeners}</td>
                  <td className="py-0.5">{row.detail}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </section>
    </div>
  )
}
