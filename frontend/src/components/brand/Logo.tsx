export function Logo({ light = false }: { light?: boolean }) {
  return (
    <svg
      aria-label="0xSocial"
      className={`brand-logo ${light ? 'brand-logo-light' : ''}`}
      role="img"
      viewBox="0 0 32 32"
      xmlns="http://www.w3.org/2000/svg"
    >
      <rect className="brand-logo-background" width="32" height="32" rx="8" />
      {/* Montserrat SemiBold, outlined so the mark never falls back to another font. */}
      <g className="brand-logo-glyph" transform="translate(5 22.36) scale(.01766 -.01766)">
        <path d="M336-10Q253-10 186.5 32T82 154.5Q44 235 44 350t38 195.5Q120 626 186.5 668T336 710q84 0 150-42t104.5-122.5Q629 465 629 350t-38.5-195.5Q552 74 486 32T336-10Zm0 113q49 0 85 26.5t56.5 81.5Q498 266 498 350t-20.5 139.5Q457 544 421 570.5T336 597q-47 0-83.5-26.5T195.5 489.5Q175 435 175 350t20.5-138.5Q216 156 252.5 129.5T336 103Z" />
        <path d="M9 0 244 306l-2-67L18 534h139l157-210h-53l158 210h135L328 239l1 67L563 0H422L257 223l52-7L147 0Z" transform="translate(673)" />
      </g>
    </svg>
  )
}
