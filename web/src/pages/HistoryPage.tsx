import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { api, Session } from '../api'

const LIMIT = 200  // 7日分を一括取得（サーバー側で7日に絞っているため上限不要）

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
      <div className="flex items-baseline gap-3 mb-4 border-b-2 border-gray-900 pb-2">
        <h1 className="font-black text-gray-900 uppercase tracking-tight text-xl">History</h1>
      </div>

      {loading ? (
        <p className="text-gray-400 text-xl font-mono">loading...</p>
      ) : (
        <table className="w-full text-xl border border-gray-900">
          <thead>
            <tr className="text-left border-b-2 border-gray-900 bg-gray-50">
              <th className="py-2 px-3 font-mono text-base uppercase tracking-wider text-gray-600">チャンネル</th>
              <th className="py-2 px-3 font-mono text-base uppercase tracking-wider text-gray-600">開始</th>
              <th className="py-2 px-3 font-mono text-base uppercase tracking-wider text-gray-600">時間</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-200">
            {sessions.map(s => (
              <tr key={s.id} className="hover:bg-gray-50 transition-colors">
                <td className="py-2 px-3">
                  <Link
                    to={`/channels/${encodeURIComponent(s.channelName)}`}
                    className="font-bold text-gray-900 hover:underline underline-offset-2"
                  >
                    {s.channelName}
                  </Link>
                  {s.genre && (
                    <span className="ml-2 text-base text-gray-400 font-mono">{s.genre}</span>
                  )}
                </td>
                <td className="py-2 px-3 font-mono text-base text-gray-500 tabular-nums whitespace-nowrap">
                  {fmtDate(s.startedAt)}
                </td>
                <td className="py-2 px-3 font-mono text-base tabular-nums">
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
