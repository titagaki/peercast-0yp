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

  if (loading) return <p className="text-washi-muted text-base">読み込み中...</p>
  if (error) return <p className="text-red-500 text-base">{error}</p>
  const hero = (
    <div className="mb-8 flex items-center gap-6">
      <img src="/yp/logo.jpeg" alt="0yp" className="h-40 w-auto rounded-xl" />
      <div>
        <p className="text-5xl font-black text-washi-text tracking-tight">令和のYP 0yp</p>
        <p className="text-2xl text-washi-muted mt-1">れいわいぴー</p>
      </div>
    </div>
  )

  if (channels.length === 0) return (
    <div>
      {hero}
      <p className="text-washi-muted text-base">現在放送中のチャンネルはありません。</p>
      <div className="mt-6">
        <Link to="/history" className="inline-flex items-center gap-2 border border-washi-accent px-4 py-2 text-sm font-bold text-washi-accent hover:bg-washi-accent hover:text-white transition-colors">
          <History size={16} />
          過去の配信
        </Link>
      </div>
    </div>
  )

  return (
    <div>
      {hero}
      <div className="flex items-baseline gap-3 mb-4 border-b-2 border-washi-header pb-3">
        <h1 className="font-black text-washi-text uppercase tracking-tight text-xl">Live</h1>
        <span className="font-mono text-sm text-washi-muted">{channels.length} ch</span>
      </div>
      <div className="divide-y divide-washi-border border border-washi-border">
        {channels.map(ch => (
          <div key={ch.id} className="flex items-start gap-4 px-4 py-4 hover:bg-washi-surface transition-colors">
            {/* live indicator */}
            <div className="pt-1 shrink-0 text-washi-accent">
              <Volume2 size={20} />
            </div>

            {/* main info */}
            <div className="flex-1 min-w-0">
              <Link
                to={`/channels/${encodeURIComponent(ch.name)}`}
                className="font-bold text-washi-accent hover:underline text-lg"
              >
                {ch.name}
              </Link>
              {(ch.genre || ch.desc) && (
                <p className="text-sm text-washi-muted mt-0.5">
                  {[ch.genre, ch.desc].filter(Boolean).join(' ')}
                </p>
              )}
              {ch.comment && (
                <p className="text-sm text-washi-muted mt-0.5">{ch.comment}</p>
              )}
              {ch.track.title && (
                <p className="text-sm text-washi-muted mt-0.5">
                  {ch.track.artist ? `♪ ${ch.track.title} / ${ch.track.artist}` : ch.track.title}
                </p>
              )}
              {ch.url && (
                <p className="text-sm mt-0.5 break-all">
                  <a href={ch.url} target="_blank" rel="noopener noreferrer"
                    className="text-washi-muted hover:text-washi-text underline underline-offset-2 font-mono"
                  >{ch.url}</a>
                </p>
              )}
            </div>

            {/* stats */}
            <div className="text-right shrink-0 font-mono text-sm text-washi-muted space-y-1">
              <div className="tabular-nums">{ch.numListeners}/{ch.numRelays}</div>
              <div className="tabular-nums">{formatUpTime(ch.upTime)}</div>
            </div>

            {/* format */}
            <div className="text-right shrink-0 font-mono text-sm text-washi-muted space-y-1 w-16">
              <div>{ch.contentType}</div>
              <div>{ch.bitrate}k</div>
            </div>
          </div>
        ))}
      </div>
      <div className="mt-6">
        <Link to="/history" className="inline-flex items-center gap-2 border border-washi-accent px-4 py-2 text-sm font-bold text-washi-accent hover:bg-washi-accent hover:text-white transition-colors">
          <History size={16} />
          過去の配信
        </Link>
      </div>
      <div className="mt-8 text-right">
        <a href="https://github.com/titagaki/peercast-0yp" target="_blank" rel="noopener noreferrer"
          className="font-mono text-xs text-washi-muted hover:text-washi-text">
          github.com/titagaki/peercast-0yp
        </a>
      </div>
    </div>
  )
}
