import type { ReactNode } from 'react'
import { useState } from 'react'
import { Icon, type IconName } from './Icon'
import { CardBody, Skeleton } from './primitives'
import { relativeTime } from '@/lib/format'

export interface Column<T> {
  key: string
  header: ReactNode
  render: (row: T) => ReactNode
  align?: 'left' | 'right' | 'center'
  width?: string
  hideOnMobile?: boolean
}

export interface DataTableProps<T> {
  columns: Column<T>[]
  rows: T[]
  rowKey: (row: T) => string
  loading?: boolean
  empty?: ReactNode
  onRowClick?: (row: T) => void
  caption?: string
}

export function DataTable<T>({
  columns,
  rows,
  rowKey,
  loading,
  empty,
  onRowClick,
  caption,
}: DataTableProps<T>) {
  if (loading) {
    return (
      <div className="table-wrap">
        <table className="table">
          <thead>
            <tr>
              {columns.map((column) => (
                <th key={column.key} style={{ width: column.width, textAlign: column.align ?? 'left' }}>
                  {column.header}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {Array.from({ length: 5 }).map((_, index) => (
              <tr key={index}>
                {columns.map((column) => (
                  <td key={column.key}>
                    <Skeleton height={12} width="80%" />
                  </td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    )
  }

  if (rows.length === 0) {
    return <div className="table-empty">{empty ?? <p className="muted">Sin resultados.</p>}</div>
  }

  return (
    <div className="table-wrap">
      <table className="table">
        {caption && <caption className="sr-only">{caption}</caption>}
        <thead>
          <tr>
            {columns.map((column) => (
              <th
                key={column.key}
                scope="col"
                style={{ width: column.width, textAlign: column.align ?? 'left' }}
                className={column.hideOnMobile ? 'hide-mobile' : undefined}
              >
                {column.header}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {rows.map((row) => (
            <tr
              key={rowKey(row)}
              onClick={onRowClick ? () => onRowClick(row) : undefined}
              className={onRowClick ? 'table__row--clickable' : undefined}
              tabIndex={onRowClick ? 0 : undefined}
              onKeyDown={
                onRowClick
                  ? (event) => {
                      if (event.key === 'Enter') onRowClick(row)
                    }
                  : undefined
              }
            >
              {columns.map((column) => (
                <td
                  key={column.key}
                  style={{ textAlign: column.align ?? 'left' }}
                  className={column.hideOnMobile ? 'hide-mobile' : undefined}
                >
                  {column.render(row)}
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

export function Pagination({
  page,
  totalPages,
  total,
  perPage,
  onPageChange,
}: {
  page: number
  totalPages: number
  total: number
  perPage: number
  onPageChange: (page: number) => void
}) {
  const from = total === 0 ? 0 : (page - 1) * perPage + 1
  const to = Math.min(page * perPage, total)
  return (
    <nav className="pagination" aria-label="Paginación">
      <span className="pagination__summary">
        {from}–{to} de {total}
      </span>
      <div className="pagination__controls">
        <button
          className="btn btn--ghost btn--sm"
          onClick={() => onPageChange(page - 1)}
          disabled={page <= 1}
          aria-label="Página anterior"
        >
          <Icon name="chevron-left" size={16} />
        </button>
        <span className="pagination__page">
          Página {page} de {Math.max(1, totalPages)}
        </span>
        <button
          className="btn btn--ghost btn--sm"
          onClick={() => onPageChange(page + 1)}
          disabled={page >= totalPages}
          aria-label="Página siguiente"
        >
          <Icon name="chevron-right" size={16} />
        </button>
      </div>
    </nav>
  )
}

export interface StatCardProps {
  label: string
  value: ReactNode
  hint?: string
  description?: string
  icon?: IconName
  tone?: 'primary' | 'accent' | 'success' | 'warning' | 'danger' | 'info' | 'neutral'
  updatedAt?: string | number | Date | null
  loading?: boolean
}

export function StatCard({
  label,
  value,
  hint,
  description,
  icon,
  tone = 'primary',
  updatedAt,
  loading,
}: StatCardProps) {
  return (
    <article className={`stat stat--${tone}`}>
      <header className="stat__header">
        <span className="stat__label">
          {label}
          {description && <MetricInfo description={description} />}
        </span>
        {icon && (
          <span className="stat__icon" aria-hidden="true">
            <Icon name={icon} size={18} />
          </span>
        )}
      </header>
      <p className="stat__value">{loading ? <Skeleton width={70} height={26} /> : value}</p>
      {hint && <p className="stat__hint">{hint}</p>}
      {updatedAt !== undefined && (
        <p className="stat__updated">
          <Icon name="clock" size={12} /> Datos consultados {relativeTime(updatedAt)}
        </p>
      )}
    </article>
  )
}

/** Explica qué representa un indicador (requisito de transparencia de datos). */
export function MetricInfo({ description }: { description: string }) {
  const [open, setOpen] = useState(false)
  return (
    <span className="metric-info">
      <button
        type="button"
        className="metric-info__button"
        aria-label="Qué representa este indicador"
        aria-expanded={open}
        onClick={() => setOpen((prev) => !prev)}
        onBlur={() => setOpen(false)}
      >
        <Icon name="info" size={13} />
      </button>
      {open && (
        <span role="tooltip" className="metric-info__tip">
          {description}
        </span>
      )}
    </span>
  )
}

export function KeyValue({ items }: { items: { label: string; value: ReactNode }[] }) {
  return (
    <dl className="kv">
      {items.map((item) => (
        <div className="kv__row" key={item.label}>
          <dt>{item.label}</dt>
          <dd>{item.value}</dd>
        </div>
      ))}
    </dl>
  )
}

export function SegmentedControl<T extends string>({
  options,
  value,
  onChange,
  label,
}: {
  options: { value: T; label: string }[]
  value: T
  onChange: (value: T) => void
  label: string
}) {
  return (
    <div className="segmented" role="tablist" aria-label={label}>
      {options.map((option) => (
        <button
          key={option.value}
          role="tab"
          aria-selected={option.value === value}
          className={`segmented__item ${option.value === value ? 'is-active' : ''}`}
          onClick={() => onChange(option.value)}
          type="button"
        >
          {option.label}
        </button>
      ))}
    </div>
  )
}

export function ChartCard({
  title,
  description,
  actions,
  children,
  updatedAt,
}: {
  title: string
  description?: string
  actions?: ReactNode
  children: ReactNode
  updatedAt?: string | number | Date | null
}) {
  return (
    <section className="card chart-card">
      <header className="card__header">
        <div className="card__heading">
          <h2 className="card__title">{title}</h2>
          {description && <p className="card__description">{description}</p>}
        </div>
        {actions && <div className="card__actions">{actions}</div>}
      </header>
      <CardBody>
        {children}
        {updatedAt !== undefined && (
          <p className="chart-card__updated">
            <Icon name="clock" size={12} /> Consulta realizada {relativeTime(updatedAt)}
          </p>
        )}
      </CardBody>
    </section>
  )
}
