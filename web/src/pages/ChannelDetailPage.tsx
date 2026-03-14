import { useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import { api, ActivityDay, Channel, TimelineRow } from '../api'
import { todayYYYYMMDD } from '../utils/date'
import { MonthCalendar } from '../components/MonthCalendar'
import { ActivityHeatmap } from '../components/ActivityHeatmap'

function formatDateDisplay(yyyymmdd: string): string {
  const y = yyyymmdd.slice(0, 4)
  const m = parseInt(yyyymmdd.slice(4, 6))
  const d = parseInt(yyyymmdd.slice(6, 8))
  return `${y}年${m}月${d}日`
}

function fmtTime(iso: string): string {
  return new Date(iso).toLocaleTimeString('ja-JP', { hour: '2-digit', minute: '2-digit' })
}

export default function ChannelDetailPage() {
  const { name } = useParams<{ name: string }>()
  const channelName = decodeURIComponent(name ?? '')

  const todayStr = todayYYYYMMDD()
  const [date, setDate] = useState(todayStr)
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
            {date !== todayStr && (
              <button
                onClick={() => setDate(todayStr)}
                className="font-mono text-xs px-2 py-0.5 border border-washi-header text-washi-header hover:bg-washi-header hover:text-white transition-colors"
              >
                今日に戻る
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
