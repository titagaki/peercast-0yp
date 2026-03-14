import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { api, Session } from '../api'
import { PageHeading } from '../components/PageHeading'

const LIMIT = 200

function fmtDate(iso: string): string {
  return new Date(iso).toLocaleString('ja-JP', {
    month: 'numeric',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

function fmtDuration(min: number): string {
  return `${Math.floor(min / 60)}:${String(min % 60).padStart(2, '0')}`
}

export default function HistoryPage() {
  const [sessions, setSessions] = useState<Session[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    setLoading(true)
    api.history(LIMIT, 0)
      .then(setSessions)
      .finally(() => setLoading(false))
  }, [])

  return (
    <div>
      <PageHeading>
        <h1 className="font-black text-washi-text uppercase tracking-tight text-xl">History</h1>
      </PageHeading>

      {loading ? (
        <p className="text-washi-muted text-base">読み込み中...</p>
      ) : (
        <table className="w-full border border-washi-border">
          <thead>
            <tr className="text-left border-b-2 border-washi-header bg-washi-surface">
              <th className="py-2 px-4 text-xs font-bold uppercase tracking-wider text-washi-muted">チャンネル</th>
              <th className="py-2 px-4 text-xs font-bold uppercase tracking-wider text-washi-muted">開始</th>
              <th className="py-2 px-4 text-xs font-bold uppercase tracking-wider text-washi-muted">時間</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-washi-border">
            {sessions.map(s => (
              <tr key={s.id} className="hover:bg-washi-surface transition-colors">
                <td className="py-2.5 px-4">
                  <Link
                    to={`/channels/${encodeURIComponent(s.channelName)}`}
                    className="font-bold text-washi-accent hover:underline underline-offset-2"
                  >
                    {s.channelName}
                  </Link>
                  {s.genre && (
                    <span className="ml-2 text-sm text-washi-muted">{s.genre}</span>
                  )}
                </td>
                <td className="py-2.5 px-4 font-mono text-sm text-washi-muted tabular-nums whitespace-nowrap">
                  {fmtDate(s.startedAt)}
                </td>
                <td className="py-2.5 px-4 font-mono text-sm tabular-nums">
                  {fmtDuration(s.durationMin)}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  )
}
