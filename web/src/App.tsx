import { BrowserRouter, Routes, Route, Navigate, useSearchParams, Link, useLocation } from 'react-router-dom'
import ChannelListPage from './pages/ChannelListPage'
import ChannelDetailPage from './pages/ChannelDetailPage'
import HistoryPage from './pages/HistoryPage'
import ChatPage from './pages/ChatPage'
import TermsPage from './pages/TermsPage'
import HowToPage from './pages/HowToPage'

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
      className={`text-base font-mono uppercase tracking-widest px-1 py-0.5 ${pathname === to ? 'text-gray-900 border-b-2 border-gray-900' : 'text-gray-400 hover:text-gray-900'}`}
    >
      {label}
    </Link>
  )
  return (
    <nav className="bg-white border-b-2 border-gray-900 py-3">
      <div className="max-w-4xl mx-auto px-4 flex items-center gap-6">
        <Link to="/" className="font-black text-gray-900 tracking-tighter text-xl">0yp</Link>
        {link('/howto', '使い方')}
        {link('/terms', '利用規約')}
      </div>
    </nav>
  )
}

function Layout({ children }: { children: React.ReactNode }) {
  return (
    <div className="min-h-screen bg-white">
      <Nav />
      <main className="max-w-4xl mx-auto px-4 py-8">{children}</main>
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
        <Route path="/howto" element={<Layout><HowToPage /></Layout>} />
        <Route path="/terms" element={<Layout><TermsPage /></Layout>} />
        <Route path="/getgmt.php" element={<GetgmtRedirect />} />
        <Route path="/chat.php" element={<Layout><ChatPage /></Layout>} />
      </Routes>
    </BrowserRouter>
  )
}
