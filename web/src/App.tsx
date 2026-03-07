import { BrowserRouter, Routes, Route, Navigate, useSearchParams, Link, useLocation } from 'react-router-dom'
import ChannelListPage from './pages/ChannelListPage'
import ChannelDetailPage from './pages/ChannelDetailPage'
import HistoryPage from './pages/HistoryPage'
import ChatPage from './pages/ChatPage'

function GetgmtRedirect() {
  const [params] = useSearchParams()
  const name = params.get('cn') ?? ''
  return <Navigate to={`/channels/${encodeURIComponent(name)}`} replace />
}

function Nav() {
  const { pathname } = useLocation()
  const link = (to: string, label: string) => (
    <Link
      to={to}
      className={`text-sm px-1 ${pathname === to ? 'text-gray-900 font-semibold' : 'text-gray-500 hover:text-gray-900'}`}
    >
      {label}
    </Link>
  )
  return (
    <nav className="bg-white border-b border-gray-200 px-4 py-3">
      <div className="max-w-4xl mx-auto flex items-center gap-6">
        <Link to="/" className="font-bold text-gray-900 tracking-tight">0yp</Link>
        {link('/', 'ライブ')}
        {link('/history', '履歴')}
      </div>
    </nav>
  )
}

function Layout({ children }: { children: React.ReactNode }) {
  return (
    <div className="min-h-screen bg-gray-50">
      <Nav />
      <main className="max-w-4xl mx-auto px-4 py-6">{children}</main>
    </div>
  )
}

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Layout><ChannelListPage /></Layout>} />
        <Route path="/history" element={<Layout><HistoryPage /></Layout>} />
        <Route path="/channels/:name" element={<Layout><ChannelDetailPage /></Layout>} />
        <Route path="/getgmt.php" element={<GetgmtRedirect />} />
        <Route path="/chat.php" element={<Layout><ChatPage /></Layout>} />
      </Routes>
    </BrowserRouter>
  )
}
