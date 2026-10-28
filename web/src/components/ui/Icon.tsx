import type { SVGProps } from 'react'

export type IconName =
  | 'dashboard'
  | 'map'
  | 'services'
  | 'activities'
  | 'reports'
  | 'organizations'
  | 'users'
  | 'user'
  | 'shield'
  | 'logout'
  | 'plus'
  | 'edit'
  | 'trash'
  | 'close'
  | 'filter'
  | 'download'
  | 'search'
  | 'calendar'
  | 'clock'
  | 'pin'
  | 'phone'
  | 'mail'
  | 'globe'
  | 'chevron-left'
  | 'chevron-right'
  | 'chevron-down'
  | 'check'
  | 'alert'
  | 'info'
  | 'menu'
  | 'refresh'
  | 'external'
  | 'list'
  | 'heart'
  | 'sparkles'
  | 'layers'
  | 'lock'
  | 'eye'
  | 'eye-off'
  | 'building'
  | 'trend'

const paths: Record<IconName, string> = {
  dashboard: 'M3 3h7v7H3zM14 3h7v4h-7zM14 10h7v11h-7zM3 13h7v8H3z',
  map: 'M9 20 3 17V4l6 3 6-3 6 3v13l-6-3-6 3zM9 7v13M15 4v13',
  services: 'M3 21h18M5 21V7l7-4 7 4v14M9 21v-6h6v6',
  activities: 'M8 3v3M16 3v3M4 8h16M5 5h14a1 1 0 0 1 1 1v14a1 1 0 0 1-1 1H5a1 1 0 0 1-1-1V6a1 1 0 0 1 1-1z',
  reports: 'M3 3v18h18M8 17V9M13 17V5M18 17v-6',
  organizations: 'M4 21V4a1 1 0 0 1 1-1h9a1 1 0 0 1 1 1v17M15 21V9h4a1 1 0 0 1 1 1v11M8 7h3M8 11h3M8 15h3',
  users: 'M17 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2M9.5 11a4 4 0 1 0 0-8 4 4 0 0 0 0 8zM22 21v-2a4 4 0 0 0-3-3.87M16 3.13A4 4 0 0 1 16 11',
  user: 'M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2M12 11a4 4 0 1 0 0-8 4 4 0 0 0 0 8z',
  shield: 'M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z',
  logout: 'M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4M16 17l5-5-5-5M21 12H9',
  plus: 'M12 5v14M5 12h14',
  edit: 'M12 20h9M16.5 3.5a2.12 2.12 0 0 1 3 3L7 19l-4 1 1-4z',
  trash: 'M3 6h18M8 6V4a1 1 0 0 1 1-1h6a1 1 0 0 1 1 1v2M19 6l-1 14a2 2 0 0 1-2 2H8a2 2 0 0 1-2-2L5 6M10 11v6M14 11v6',
  close: 'M18 6 6 18M6 6l12 12',
  filter: 'M22 3H2l8 9.46V19l4 2v-8.54z',
  download: 'M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4M7 10l5 5 5-5M12 15V3',
  search: 'M11 19a8 8 0 1 0 0-16 8 8 0 0 0 0 16zM21 21l-4.35-4.35',
  calendar: 'M8 2v4M16 2v4M3 10h18M5 6h14a2 2 0 0 1 2 2v12a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2z',
  clock: 'M12 22a10 10 0 1 0 0-20 10 10 0 0 0 0 20zM12 7v5l3 2',
  pin: 'M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0zM12 13a3 3 0 1 0 0-6 3 3 0 0 0 0 6z',
  phone: 'M22 16.92v3a2 2 0 0 1-2.18 2 19.79 19.79 0 0 1-8.63-3.07 19.5 19.5 0 0 1-6-6 19.79 19.79 0 0 1-3.07-8.67A2 2 0 0 1 4.11 2h3a2 2 0 0 1 2 1.72c.13.96.36 1.9.7 2.81a2 2 0 0 1-.45 2.11L8.09 9.91a16 16 0 0 0 6 6l1.27-1.27a2 2 0 0 1 2.11-.45c.9.34 1.85.57 2.81.7A2 2 0 0 1 22 16.92z',
  mail: 'M4 4h16a2 2 0 0 1 2 2v12a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2zM22 6l-10 7L2 6',
  globe: 'M12 22a10 10 0 1 0 0-20 10 10 0 0 0 0 20zM2 12h20M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z',
  'chevron-left': 'M15 18l-6-6 6-6',
  'chevron-right': 'M9 18l6-6-6-6',
  'chevron-down': 'M6 9l6 6 6-6',
  check: 'M20 6 9 17l-5-5',
  alert: 'M10.29 3.86 1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0zM12 9v4M12 17h.01',
  info: 'M12 22a10 10 0 1 0 0-20 10 10 0 0 0 0 20zM12 16v-4M12 8h.01',
  menu: 'M3 12h18M3 6h18M3 18h18',
  refresh: 'M23 4v6h-6M1 20v-6h6M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15',
  external: 'M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6M15 3h6v6M10 14 21 3',
  list: 'M8 6h13M8 12h13M8 18h13M3 6h.01M3 12h.01M3 18h.01',
  heart: 'M20.84 4.61a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 0 0-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 0 0 0-7.78z',
  sparkles: 'M12 3l1.9 5.1L19 10l-5.1 1.9L12 17l-1.9-5.1L5 10l5.1-1.9zM19 17l.7 1.8L21.5 19.5l-1.8.7L19 22l-.7-1.8L16.5 19.5l1.8-.7z',
  layers: 'M12 2 2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5',
  lock: 'M5 11h14a1 1 0 0 1 1 1v9a1 1 0 0 1-1 1H5a1 1 0 0 1-1-1v-9a1 1 0 0 1 1-1zM8 11V7a4 4 0 0 1 8 0v4',
  eye: 'M1 12s4-7 11-7 11 7 11 7-4 7-11 7-11-7-11-7zM12 15a3 3 0 1 0 0-6 3 3 0 0 0 0 6z',
  'eye-off': 'M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19M1 1l22 22M14.12 14.12a3 3 0 1 1-4.24-4.24',
  building: 'M6 22V4a2 2 0 0 1 2-2h8a2 2 0 0 1 2 2v18M2 22h20M10 6h.01M14 6h.01M10 10h.01M14 10h.01M10 14h.01M14 14h.01',
  trend: 'M23 6l-9.5 9.5-5-5L1 18M17 6h6v6',
}

export interface IconProps extends SVGProps<SVGSVGElement> {
  name: IconName
  size?: number
}

export function Icon({ name, size = 18, ...rest }: IconProps) {
  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth={1.8}
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
      focusable="false"
      {...rest}
    >
      <path d={paths[name]} />
    </svg>
  )
}

/** Marca de Abudemy (isotipo + nombre). */
export function AbudemyLogo({ compact = false }: { compact?: boolean }) {
  return (
    <span className="brand">
      <span className="brand__mark" aria-hidden="true">
        <svg viewBox="0 0 32 32" width="26" height="26" fill="none">
          <circle cx="16" cy="16" r="15" fill="currentColor" opacity="0.14" />
          <path
            d="M16 6c3.6 0 6.5 2.9 6.5 6.5 0 4.6-4.3 8.6-6.5 12.5-2.2-3.9-6.5-7.9-6.5-12.5C9.5 8.9 12.4 6 16 6z"
            fill="currentColor"
            opacity="0.35"
          />
          <path
            d="M16 9.5a3.2 3.2 0 1 1 0 6.4 3.2 3.2 0 0 1 0-6.4z"
            fill="currentColor"
          />
          <path d="M16 15.9V25" stroke="currentColor" strokeWidth="2" strokeLinecap="round" />
        </svg>
      </span>
      {!compact && (
        <span className="brand__text">
          <span className="brand__name">Abudemy</span>
          <span className="brand__tagline">Plataforma institucional</span>
        </span>
      )}
    </span>
  )
}
