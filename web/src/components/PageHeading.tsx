export function PageHeading({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex items-baseline gap-3 mb-4 border-b-2 border-washi-header pb-3">
      {children}
    </div>
  )
}
