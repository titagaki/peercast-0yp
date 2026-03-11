import { useEffect, useState } from 'react'
import { api, SiteConfig } from '../api'

function CodeBlock({ value, placeholder }: { value?: string; placeholder: string }) {
  return (
    <div className={`rounded px-4 py-3 font-mono text-xs break-all select-all ${value ? 'bg-gray-100' : 'bg-gray-50 text-gray-400 italic'}`}>
      {value || placeholder}
    </div>
  )
}

export default function HowToPage() {
  const [config, setConfig] = useState<SiteConfig | null>(null)
  useEffect(() => { api.config().then(setConfig).catch(() => {}) }, [])

  return (
    <div className="max-w-2xl space-y-10 text-sm text-gray-700">
      <h1 className="text-xl font-bold text-gray-900">使い方</h1>

      <section className="space-y-4">
        <h2 className="text-base font-semibold text-gray-900">視聴方法</h2>
        <p>
          PeerCast の YP（Yellow Page）対応クライアントのチャンネル一覧 URL に以下を登録してください。
        </p>
        <CodeBlock value={config?.ypIndexURL} placeholder="（未設定）" />
      </section>

      <section className="space-y-4">
        <h2 className="text-base font-semibold text-gray-900">掲載方法</h2>
        <p>
          PeerCastStation の YellowPage 設定の「配信掲載 URL」に以下を登録してください。
        </p>
        <CodeBlock value={config?.pcpAddress} placeholder="（未設定）" />
        <div className="space-y-2">
          <p className="font-medium text-gray-800">ジャンルの設定</p>
          <p className="text-gray-600">
            ジャンル欄の先頭に <code className="bg-gray-100 px-1 rounded">yp</code> を付けてください。
            付いていないチャンネルはこの YP には掲載されません。
          </p>
          <table className="w-full text-xs border-collapse">
            <thead>
              <tr className="text-left text-gray-500 border-b border-gray-200">
                <th className="py-1 pr-4 font-medium">設定例</th>
                <th className="py-1 font-medium">効果</th>
              </tr>
            </thead>
            <tbody className="text-gray-600">
              <tr className="border-b border-gray-100">
                <td className="py-1 pr-4 font-mono">ypゲーム</td>
                <td className="py-1">ジャンル「ゲーム」で掲載</td>
              </tr>
              <tr className="border-b border-gray-100">
                <td className="py-1 pr-4 font-mono">yp?ゲーム</td>
                <td className="py-1">リスナー数を非表示にして掲載</td>
              </tr>
            </tbody>
          </table>
        </div>
        <div className="space-y-1 bg-amber-50 border border-amber-200 rounded px-4 py-3">
          <p className="font-medium text-amber-800">掲載にあたってのお願い</p>
          <ul className="list-disc list-inside space-y-1 text-amber-700">
            <li>著作権を侵害するコンテンツの配信はお控えください</li>
            <li>
              掲載には<a href="/terms" className="underline hover:text-amber-900">利用規約</a>への同意が必要です
            </li>
          </ul>
        </div>
      </section>
    </div>
  )
}
