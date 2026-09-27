export function Logo({ light = false }: { light?: boolean }) {
  return (
    <div className={`brand-logo ${light ? 'text-white' : 'text-base-content'}`}>
      <span>0x</span><strong>social</strong>
    </div>
  )
}
