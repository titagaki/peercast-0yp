import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { api, Channel } from '../api'

function formatUpTime(seconds: number): string {
  const h = Math.floor(seconds / 3600)
  const m = Math.floor((seconds % 3600) / 60)
  return `${h}:${String(m).padStart(2, '0')}`
}

export default function ChannelListPage() {
  const [channels, setChannels] = useState<Channel[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const load = () =>
      api.channels()
        .then(setChannels)
        .catch(e => setError(String(e)))
        .finally(() => setLoading(false))
    load()
    const id = setInterval(load, 60_000)
    return () => clearInterval(id)
  }, [])

  if (loading) return <p className="text-gray-400 text-sm">読み込み中...</p>
  if (error) return <p className="text-red-500 text-sm">{error}</p>
  if (channels.length === 0) return <p className="text-gray-400 text-sm">現在放送中のチャンネルはありません。</p>

  return (
    <div className="space-y-3">
      <h1 className="text-sm font-semibold text-gray-500 uppercase tracking-wide">
        ライブ中 ({channels.length})
      </h1>
      {channels.map(ch => (
        <div key={ch.id} className="bg-white rounded-lg border border-gray-200 p-4">
          <div className="flex items-start justify-between gap-4">
            <div className="flex-1 min-w-0">
              <Link
                to={`/channels/${encodeURIComponent(ch.name)}`}
                className="font-semibold text-blue-600 hover:underline"
              >
                {ch.name}
              </Link>
              {(ch.genre || ch.desc) && (
                <p className="text-sm text-gray-500 mt-0.5">
                  {[ch.genre, ch.desc].filter(Boolean).join(' / ')}
                </p>
              )}
              {ch.track.title && (
                <p className="text-sm text-gray-600 mt-1">♪ {ch.track.title}</p>
              )}
              {ch.comment && (
                <p className="text-sm text-gray-400 mt-1">{ch.comment}</p>
              )}
            </div>
            <div className="text-right text-sm text-gray-500 shrink-0 space-y-0.5">
              <div>{ch.contentType} {ch.bitrate}kbps</div>
              <div>👥 {ch.numListeners}</div>
              <div className="tabular-nums">{formatUpTime(ch.upTime)}</div>
            </div>
          </div>
        </div>
      ))}
    </div>
  )
}
