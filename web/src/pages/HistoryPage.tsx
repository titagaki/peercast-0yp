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
    <div className="space-y-4">
      <h1 className="text-sm font-semibold text-gray-500 uppercase tracking-wide">放送履歴</h1>

      {loading ? (
        <p className="text-gray-400 text-sm">読み込み中...</p>
      ) : (
        <>
          <table className="w-full text-sm">
            <thead>
              <tr className="text-left text-gray-500 border-b border-gray-200">
                <th className="py-1 pr-4 font-medium">チャンネル</th>
                <th className="py-1 pr-4 font-medium">開始</th>
                <th className="py-1 pr-4 font-medium">時間</th>
                <th className="py-1 font-medium">形式</th>
              </tr>
            </thead>
            <tbody>
              {sessions.map(s => (
                <tr key={s.id} className="border-b border-gray-100">
                  <td className="py-1.5 pr-4">
                    <Link
                      to={`/channels/${encodeURIComponent(s.channelName)}`}
                      className="text-blue-600 hover:underline"
                    >
                      {s.channelName}
                    </Link>
                    {s.genre && (
                      <span className="ml-2 text-gray-400">{s.genre}</span>
                    )}
                  </td>
                  <td className="py-1.5 pr-4 tabular-nums text-gray-500">
                    {fmtDate(s.startedAt)}
                  </td>
                  <td className="py-1.5 pr-4 tabular-nums">
                    {fmtDuration(s.durationMin)}
                  </td>
                  <td className="py-1.5 text-gray-500">
                    {s.contentType} {s.bitrate}kbps
                  </td>
                </tr>
              ))}
            </tbody>
          </table>

        </>
      )}
    </div>
  )
}
