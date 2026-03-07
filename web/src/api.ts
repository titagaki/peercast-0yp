export interface Channel {
  id: string
  name: string
  genre: string
  desc: string
  url: string
  comment: string
  bitrate: number
  contentType: string
  track: { title: string; artist: string; album: string; contact: string }
  tracker: { ip: string; port: number; firewalled: boolean }
  numListeners: number
  numRelays: number
  upTime: number
}

export interface Session {
  id: number
  channelId: string
  channelName: string
  bitrate: number
  contentType: string
  genre: string
  url: string
  startedAt: string
  endedAt: string | null
  durationMin: number
}

export interface ActivityDay {
  date: string    // "YYYY-MM-DD"
  minutes: number
}

export interface TimelineRow {
  recordedAt: string
  listeners: number
  relays: number
  changed: boolean
  name?: string
  genre?: string
  description?: string
  url?: string
  comment?: string
  trackTitle?: string
  trackArtist?: string
}

async function get<T>(path: string): Promise<T> {
  const res = await fetch(path)
  if (!res.ok) throw new Error(`${res.status} ${res.statusText}`)
  return res.json()
}

export const api = {
  channels: () =>
    get<Channel[]>('/api/channels'),
  history: (limit = 50, offset = 0) =>
    get<Session[]>(`/api/history?limit=${limit}&offset=${offset}`),
  activity: (name: string) =>
    get<ActivityDay[]>(`/api/channels/activity?name=${encodeURIComponent(name)}`),
  timeline: (name: string, date: string) =>
    get<TimelineRow[]>(`/api/channels/timeline?name=${encodeURIComponent(name)}&date=${date}`),
}
