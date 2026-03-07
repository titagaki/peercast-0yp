export default function TermsPage() {
  return (
    <div className="max-w-2xl space-y-6 text-sm text-gray-700">
      <h1 className="text-xl font-bold text-gray-900">利用規約</h1>

      <section className="space-y-2">
        <h2 className="font-semibold text-gray-900">第1条（適用）</h2>
        <p>本規約はユーザーと管理者間の本サービス利用に関する一切の関係に適用されます。管理者は個別規定を定めることがあり、矛盾が生じた場合は個別規定が優先されます。</p>
      </section>

      <section className="space-y-2">
        <h2 className="font-semibold text-gray-900">第2条（禁止事項）</h2>
        <p>ユーザーは以下の行為を行ってはなりません。</p>
        <ul className="list-disc list-inside space-y-1 text-gray-600">
          <li>法令または公序良俗に違反する行為</li>
          <li>犯罪行為に関連する行為</li>
          <li>著作権・商標権等の知的財産権を侵害する行為</li>
          <li>本サービスのサーバーまたはネットワークの機能を破壊・妨害する行為</li>
          <li>不正アクセスまたはこれを試みる行為</li>
          <li>他のユーザーに関する個人情報等を収集または蓄積する行為</li>
          <li>反社会的勢力に対して直接または間接に利益を供与する行為</li>
          <li>その他、管理者が不適切と判断する行為</li>
        </ul>
      </section>

      <section className="space-y-2">
        <h2 className="font-semibold text-gray-900">第3条（本サービスの提供の停止等）</h2>
        <p>管理者は、以下のいずれかに該当する場合、ユーザーへの事前通知なく本サービスの提供を停止または中断できます。停止・中断によりユーザーまたは第三者が被った損害について、管理者は責任を負いません。</p>
        <ul className="list-disc list-inside space-y-1 text-gray-600">
          <li>システムの保守・点検を行う場合</li>
          <li>地震・落雷・火災・停電等の不可抗力により運営が困難となった場合</li>
          <li>その他、管理者が必要と判断した場合</li>
        </ul>
      </section>

      <section className="space-y-2">
        <h2 className="font-semibold text-gray-900">第4条（利用制限）</h2>
        <p>管理者は、ユーザーが本規約に違反した場合、またはその他管理者が不適当と判断した場合、予告なく当該ユーザーの利用を制限できます。これによりユーザーが被った損害について管理者は責任を負いません。</p>
      </section>

      <section className="space-y-2">
        <h2 className="font-semibold text-gray-900">第5条（保証の否認および免責事項）</h2>
        <p>管理者は本サービスに事実上または法律上の瑕疵がないことを保証しません。本サービスに起因してユーザーに生じた損害、およびユーザー間または第三者との間で生じた取引・紛争について管理者は責任を負いません。ユーザーは自身が公開した情報およびコンテンツについて一切の責任を負います。</p>
      </section>

      <section className="space-y-2">
        <h2 className="font-semibold text-gray-900">第6条（サービス内容の変更等）</h2>
        <p>管理者は、ユーザーへの通知なく本サービスの内容を変更または中止できます。これによりユーザーが被った損害について管理者は責任を負いません。</p>
      </section>

      <section className="space-y-2">
        <h2 className="font-semibold text-gray-900">第7条（利用規約の変更）</h2>
        <p>管理者は、ユーザーへの通知なく本規約を変更できます。変更後に本サービスを利用したユーザーは、変更後の規約に同意したものとみなします。</p>
      </section>

      <section className="space-y-2">
        <h2 className="font-semibold text-gray-900">第8条（権利義務の譲渡の禁止）</h2>
        <p>ユーザーは、管理者の書面による事前承諾なく、本規約上の地位または権利義務を第三者に譲渡できません。</p>
      </section>

      <section className="space-y-2">
        <h2 className="font-semibold text-gray-900">第9条（準拠法・裁判管轄）</h2>
        <p>本規約の解釈にあたっては日本法を準拠法とします。本サービスに関して紛争が生じた場合、管理者の所在地を管轄する裁判所を専属的合意管轄とします。</p>
      </section>
    </div>
  )
}
