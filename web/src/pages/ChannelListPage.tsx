import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { History, Volume2 } from 'lucide-react'
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

  if (loading) return <p className="text-gray-400 text-base font-mono">loading...</p>
  if (error) return <p className="text-red-500 text-base font-mono">{error}</p>
  if (channels.length === 0) return (
    <div>
      <div className="mb-8">
        <p className="text-5xl font-black text-gray-900 tracking-tight">令和のYP 0yp</p>
        <p className="text-[1.8rem] text-gray-400 font-mono mt-1">れいわいぴー</p>
      </div>
      <p className="text-gray-400 text-base font-mono">現在放送中のチャンネルはありません。</p>
      <div className="mt-6">
        <Link to="/history" className="inline-flex items-center gap-2 border border-gray-900 px-4 py-2 font-mono text-sm font-bold text-gray-900 hover:bg-gray-900 hover:text-white transition-colors">
          <History size={16} />
          過去の配信
        </Link>
      </div>
    </div>
  )

  return (
    <div>
      <div className="mb-8">
        <p className="text-5xl font-black text-gray-900 tracking-tight">令和のYP 0yp</p>
        <p className="text-[1.8rem] text-gray-400 font-mono mt-1">れいわいぴー</p>
      </div>
      <div className="flex items-baseline gap-3 mb-4 border-b-2 border-gray-900 pb-2">
        <h1 className="font-black text-gray-900 uppercase tracking-tight text-xl">Live</h1>
        <span className="font-mono text-base text-gray-500">{channels.length} ch</span>
      </div>
      <div className="divide-y divide-gray-200 border border-gray-900">
        {channels.map(ch => (
          <div key={ch.id} className="flex items-start gap-4 px-4 py-5 hover:bg-gray-50 transition-colors">
            {/* live indicator */}
            <div className="pt-1 shrink-0 text-red-500">
              <Volume2 size={22} />
            </div>

            {/* main info */}
            <div className="flex-1 min-w-0">
              <Link
                to={`/channels/${encodeURIComponent(ch.name)}`}
                className="font-bold text-gray-900 hover:underline text-xl"
              >
                {ch.name}
              </Link>
              {(ch.genre || ch.desc) && (
                <p className="text-base text-gray-500 mt-0.5">
                  {[ch.genre, ch.desc].filter(Boolean).join(' ')}
                </p>
              )}
              {ch.comment && (
                <p className="text-base text-gray-400 mt-1">{ch.comment}</p>
              )}
              {ch.track.title && (
                <p className="text-base text-gray-500 mt-1">
                  {ch.track.artist ? `♪ ${ch.track.title} / ${ch.track.artist}` : ch.track.title}
                </p>
              )}
              {ch.url && (
                <p className="text-base mt-1 break-all">
                  <a href={ch.url} target="_blank" rel="noopener noreferrer"
                    className="text-gray-400 hover:text-gray-700 underline underline-offset-2"
                  >{ch.url}</a>
                </p>
              )}
            </div>

            {/* stats */}
            <div className="text-right shrink-0 font-mono text-base text-gray-500 space-y-1">
              <div className="tabular-nums">{ch.numListeners}/{ch.numRelays}</div>
              <div className="tabular-nums">{formatUpTime(ch.upTime)}</div>
            </div>

            {/* format */}
            <div className="text-right shrink-0 font-mono text-base text-gray-400 space-y-1 w-20">
              <div>{ch.contentType}</div>
              <div>{ch.bitrate}k</div>
            </div>
          </div>
        ))}
      </div>
      <div className="mt-6">
        <Link to="/history" className="inline-flex items-center gap-2 border border-gray-900 px-4 py-2 font-mono text-base font-bold text-gray-900 hover:bg-gray-900 hover:text-white transition-colors">
          <History size={18} />
          過去の配信
        </Link>
      </div>
      <div className="mt-8 text-right">
        <a href="https://github.com/titagaki/peercast-0yp" target="_blank" rel="noopener noreferrer"
          className="font-mono text-sm text-gray-400 hover:text-gray-900">
          github.com/titagaki/peercast-0yp
        </a>
      </div>
    </div>
  )
}
