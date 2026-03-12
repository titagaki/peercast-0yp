export default function TermsPage() {
  return (
    <div className="max-w-2xl space-y-6 text-base text-washi-text">
      <div>
        <h1 className="text-xl font-black text-washi-text">利用規約</h1>
        <p className="text-sm text-washi-muted mt-1">最終更新日：2026年3月13日</p>
      </div>

      <section className="space-y-2">
        <h2 className="font-bold text-washi-text">第1条（規約の適用）</h2>
        <ol className="space-y-1 text-washi-text list-decimal list-inside">
          <li>この利用規約（以下「本規約」といいます）は、本サービスの利用に関して、ユーザーと管理者との間に適用されます。</li>
          <li>ユーザーが本サービスを利用した時点で、本規約に同意したものとみなします。</li>
        </ol>
      </section>

      <section className="space-y-2">
        <h2 className="font-bold text-washi-text">第2条（規約の変更）</h2>
        <ol start={3} className="space-y-1 text-washi-text list-decimal list-inside">
          <li>管理者は、必要に応じて本規約を変更できるものとします。</li>
          <li>規約を変更する場合、管理者は変更内容と適用開始日を本サービス上に掲示します。</li>
          <li>掲示後、適用開始日以降に本サービスを利用した場合、ユーザーは変更後の規約に同意したものとみなします。</li>
        </ol>
      </section>

      <section className="space-y-2">
        <h2 className="font-bold text-washi-text">第3条（禁止事項）</h2>
        <p>ユーザーは、以下の行為をしてはなりません。</p>
        <ol start={6} className="space-y-1 text-washi-muted list-decimal list-inside">
          <li>法令、公序良俗に反する行為</li>
          <li>知的財産権、肖像権、プライバシーその他第三者の権利を侵害する行為</li>
          <li>サーバーやネットワークに過度な負荷をかける行為（不正なクローリング、APIの乱用、DoS攻撃など）</li>
          <li>ストリーム情報（チャンネル名など）へのスクリプト・過度な広告・不適切な表現の掲載</li>
          <li>その他、管理者が運営上不適当と判断する行為</li>
        </ol>
      </section>

      <section className="space-y-2">
        <h2 className="font-bold text-washi-text">第4条（掲載情報の削除・利用制限）</h2>
        <ol start={11} className="space-y-2 text-washi-text list-decimal list-inside">
          <li>
            管理者は、ユーザーが前条に違反した場合、または違反のおそれがあると判断した場合、事前の通知なく以下の措置をとることができます。
            <ol className="mt-1 ml-4 space-y-1 text-washi-muted list-[lower-roman] list-inside">
              <li>掲載情報の削除・修正</li>
              <li>本サービスの利用制限</li>
              <li>その他、管理者が必要と判断する措置</li>
            </ol>
          </li>
          <li>管理者がこれらの措置をとったことにより、ユーザーに生じた損害について、管理者は責任を負いません。</li>
        </ol>
      </section>

      <section className="space-y-2">
        <h2 className="font-bold text-washi-text">第5条（サービスの内容と保証の否認）</h2>
        <ol start={13} className="space-y-1 text-washi-text list-decimal list-inside">
          <li>本サービスは、ストリーム情報のインデックス（索引）を提供するものであり、配信データ（映像・音声など）を保存・管理するものではありません。</li>
          <li>管理者は、本サービスの内容（情報の正確さ、合法性、特定の目的への適合性など）について、いかなる保証もしません。</li>
          <li>管理者は、本サービスの内容を予告なく変更・追加・削除できるものとします。</li>
        </ol>
      </section>

      <section className="space-y-2">
        <h2 className="font-bold text-washi-text">第6条（免責事項）</h2>
        <p>管理者に故意または重大な過失がある場合を除き、管理者は以下について責任を負いません。</p>
        <ol start={16} className="space-y-1 text-washi-muted list-decimal list-inside">
          <li>本サービスの利用により、ユーザーに生じた損害</li>
          <li>P2P通信に起因するトラブル（通信帯域の消費、機器の不具合、ユーザー間の紛争など）</li>
          <li>本サービスの停止・変更・終了により生じた損害</li>
          <li>第三者が提供するコンテンツに関する損害</li>
        </ol>
      </section>

      <section className="space-y-2">
        <h2 className="font-bold text-washi-text">第7条（損害賠償・求償）</h2>
        <ol start={20} className="space-y-1 text-washi-text list-decimal list-inside">
          <li>ユーザーが本規約に違反し、管理者に損害を与えた場合、ユーザーはその損害（弁護士費用など合理的な費用を含みます）を賠償するものとします。</li>
          <li>ユーザーの行為が原因で、管理者が第三者から請求を受けた場合、ユーザーは自分の責任と費用でこれを解決し、管理者の損失を補償するものとします。</li>
        </ol>
      </section>

      <section className="space-y-2">
        <h2 className="font-bold text-washi-text">第8条（サービスの停止・中断）</h2>
        <ol start={22} className="space-y-2 text-washi-text list-decimal list-inside">
          <li>
            管理者は、以下の場合に本サービスの全部または一部を、事前の通知なく停止・中断できるものとします。
            <ol className="mt-1 ml-4 space-y-1 text-washi-muted list-[lower-roman] list-inside">
              <li>システムの保守・点検を行う場合</li>
              <li>天災・通信障害など不可抗力が生じた場合</li>
              <li>その他、管理者がやむを得ないと判断した場合</li>
            </ol>
          </li>
          <li>サービスの停止・中断によりユーザーに生じた損害について、管理者は責任を負いません。</li>
        </ol>
      </section>

      <section className="space-y-2">
        <h2 className="font-bold text-washi-text">第9条（個人情報の取り扱い）</h2>
        <ol start={24} className="space-y-1 text-washi-text list-decimal list-inside">
          <li>管理者は、本サービスの運営にあたり取得した個人情報（IPアドレス、アクセスログなどを含みます）を、個人情報保護法その他の法令に従い適切に取り扱います。</li>
          <li>個人情報の取り扱いの詳細は、別途定めるプライバシーポリシーによるものとします。</li>
        </ol>
      </section>

      <section className="space-y-2">
        <h2 className="font-bold text-washi-text">第10条（分離可能性）</h2>
        <p>本規約のいずれかの条項が法令により無効または執行不能と判断された場合でも、残りの条項は引き続き有効に存続するものとします。</p>
      </section>

      <section className="space-y-2">
        <h2 className="font-bold text-washi-text">第11条（準拠法・裁判管轄）</h2>
        <ol start={26} className="space-y-1 text-washi-text list-decimal list-inside">
          <li>本規約は日本法に基づき解釈されます。</li>
          <li>本サービスに関する紛争は、管理者の所在地を管轄する裁判所を専属的合意管轄裁判所とします。</li>
        </ol>
      </section>

      <p className="text-washi-muted text-sm pt-2">以上</p>
    </div>
  )
}
