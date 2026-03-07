import { useSearchParams } from 'react-router-dom'

export default function ChatPage() {
  const [params] = useSearchParams()
  const name = params.get('cn')

  return (
    <div className="text-center py-16 text-gray-500">
      {name && <p className="font-semibold text-gray-700 mb-2">{name}</p>}
      <p>0yp ではチャット機能を提供していません。</p>
    </div>
  )
}
