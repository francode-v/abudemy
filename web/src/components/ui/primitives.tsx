import {
  forwardRef,
  useId,
  type ButtonHTMLAttributes,
  type InputHTMLAttributes,
  type ReactNode,
  type SelectHTMLAttributes,
  type TextareaHTMLAttributes,
} from 'react'
import { Icon, type IconName } from './Icon'

type ButtonVariant = 'primary' | 'secondary' | 'ghost' | 'danger' | 'subtle'
type ButtonSize = 'sm' | 'md' | 'lg'

export interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: ButtonVariant
  size?: ButtonSize
  icon?: IconName
  iconRight?: IconName
  loading?: boolean
  block?: boolean
}

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(function Button(
  { variant = 'primary', size = 'md', icon, iconRight, loading, block, children, className, disabled, ...rest },
  ref,
) {
  return (
    <button
      ref={ref}
      className={['btn', `btn--${variant}`, `btn--${size}`, block ? 'btn--block' : '', className]
        .filter(Boolean)
        .join(' ')}
      disabled={disabled || loading}
      {...rest}
    >
      {loading ? <span className="spinner spinner--sm" aria-hidden="true" /> : icon ? <Icon name={icon} size={16} /> : null}
      {children != null && <span>{children}</span>}
      {iconRight && <Icon name={iconRight} size={16} />}
    </button>
  )
})

export interface IconButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  icon: IconName
  label: string
  variant?: ButtonVariant
  size?: ButtonSize
}

export const IconButton = forwardRef<HTMLButtonElement, IconButtonProps>(function IconButton(
  { icon, label, variant = 'ghost', size = 'md', className, ...rest },
  ref,
) {
  return (
    <button
      ref={ref}
      className={['btn', 'btn--icon', `btn--${variant}`, `btn--${size}`, className]
        .filter(Boolean)
        .join(' ')}
      title={label}
      aria-label={label}
      {...rest}
    >
      <Icon name={icon} size={size === 'sm' ? 14 : 16} />
    </button>
  )
})

export interface FieldProps {
  label: string
  htmlFor?: string
  hint?: string
  error?: string
  required?: boolean
  children: ReactNode
}

export function Field({ label, htmlFor, hint, error, required, children }: FieldProps) {
  return (
    <div className={['field', error ? 'field--invalid' : ''].filter(Boolean).join(' ')}>
      <label className="field__label" htmlFor={htmlFor}>
        {label}
        {required && <span className="field__required" aria-hidden="true"> *</span>}
      </label>
      {children}
      {hint && !error && <p className="field__hint">{hint}</p>}
      {error && (
        <p className="field__error" role="alert">
          <Icon name="alert" size={13} /> {error}
        </p>
      )}
    </div>
  )
}

export interface TextInputProps extends InputHTMLAttributes<HTMLInputElement> {
  label?: string
  hint?: string
  error?: string
  leadingIcon?: IconName
}

export const TextInput = forwardRef<HTMLInputElement, TextInputProps>(function TextInput(
  { label, hint, error, leadingIcon, id, className, ...rest },
  ref,
) {
  const autoId = useId()
  const inputId = id ?? autoId
  const input = (
    <div className={['input-wrap', leadingIcon ? 'input-wrap--icon' : ''].filter(Boolean).join(' ')}>
      {leadingIcon && <Icon name={leadingIcon} size={16} className="input-wrap__icon" />}
      <input
        ref={ref}
        id={inputId}
        className={['input', error ? 'input--invalid' : '', className].filter(Boolean).join(' ')}
        aria-invalid={Boolean(error)}
        {...rest}
      />
    </div>
  )
  if (!label) return input
  return (
    <Field label={label} htmlFor={inputId} hint={hint} error={error} required={rest.required}>
      {input}
    </Field>
  )
})

export interface SelectInputProps extends SelectHTMLAttributes<HTMLSelectElement> {
  label?: string
  hint?: string
  error?: string
}

export const SelectInput = forwardRef<HTMLSelectElement, SelectInputProps>(function SelectInput(
  { label, hint, error, id, className, children, ...rest },
  ref,
) {
  const autoId = useId()
  const selectId = id ?? autoId
  const select = (
    <select
      ref={ref}
      id={selectId}
      className={['input', 'input--select', error ? 'input--invalid' : '', className]
        .filter(Boolean)
        .join(' ')}
      aria-invalid={Boolean(error)}
      {...rest}
    >
      {children}
    </select>
  )
  if (!label) return select
  return (
    <Field label={label} htmlFor={selectId} hint={hint} error={error} required={rest.required}>
      {select}
    </Field>
  )
})

export interface TextAreaInputProps extends TextareaHTMLAttributes<HTMLTextAreaElement> {
  label?: string
  hint?: string
  error?: string
}

export const TextAreaInput = forwardRef<HTMLTextAreaElement, TextAreaInputProps>(function TextAreaInput(
  { label, hint, error, id, className, ...rest },
  ref,
) {
  const autoId = useId()
  const areaId = id ?? autoId
  const area = (
    <textarea
      ref={ref}
      id={areaId}
      className={['input', 'input--area', error ? 'input--invalid' : '', className]
        .filter(Boolean)
        .join(' ')}
      aria-invalid={Boolean(error)}
      {...rest}
    />
  )
  if (!label) return area
  return (
    <Field label={label} htmlFor={areaId} hint={hint} error={error} required={rest.required}>
      {area}
    </Field>
  )
})

export interface CheckboxInputProps extends InputHTMLAttributes<HTMLInputElement> {
  label: ReactNode
}

export function CheckboxInput({ label, id, ...rest }: CheckboxInputProps) {
  const autoId = useId()
  const inputId = id ?? autoId
  return (
    <label className="checkbox" htmlFor={inputId}>
      <input id={inputId} type="checkbox" {...rest} />
      <span>{label}</span>
    </label>
  )
}

type BadgeTone = 'neutral' | 'success' | 'warning' | 'danger' | 'info' | 'primary'

export function Badge({ children, tone = 'neutral' }: { children: ReactNode; tone?: BadgeTone }) {
  return <span className={`badge badge--${tone}`}>{children}</span>
}

export function Card({
  children,
  className,
  as: Tag = 'section',
}: {
  children: ReactNode
  className?: string
  as?: 'section' | 'div' | 'article'
}) {
  return <Tag className={['card', className].filter(Boolean).join(' ')}>{children}</Tag>
}

export function CardHeader({
  title,
  description,
  actions,
  icon,
}: {
  title: ReactNode
  description?: ReactNode
  actions?: ReactNode
  icon?: IconName
}) {
  return (
    <header className="card__header">
      {icon && (
        <span className="card__icon" aria-hidden="true">
          <Icon name={icon} size={18} />
        </span>
      )}
      <div className="card__heading">
        <h2 className="card__title">{title}</h2>
        {description && <p className="card__description">{description}</p>}
      </div>
      {actions && <div className="card__actions">{actions}</div>}
    </header>
  )
}

export function CardBody({ children, className }: { children: ReactNode; className?: string }) {
  return <div className={['card__body', className].filter(Boolean).join(' ')}>{children}</div>
}

export function Spinner({ size = 'md' }: { size?: 'sm' | 'md' | 'lg' }) {
  return <span className={`spinner spinner--${size}`} role="status" aria-label="Cargando" />
}

export function Skeleton({ width = '100%', height = 14 }: { width?: string | number; height?: number }) {
  return <span className="skeleton" style={{ width, height }} aria-hidden="true" />
}

export function EmptyState({
  title,
  description,
  icon = 'info',
  action,
}: {
  title: string
  description?: string
  icon?: IconName
  action?: ReactNode
}) {
  return (
    <div className="state state--empty">
      <span className="state__icon" aria-hidden="true">
        <Icon name={icon} size={22} />
      </span>
      <h3 className="state__title">{title}</h3>
      {description && <p className="state__description">{description}</p>}
      {action && <div className="state__action">{action}</div>}
    </div>
  )
}

export function ErrorState({
  title = 'No se pudieron cargar los datos',
  description,
  onRetry,
}: {
  title?: string
  description?: string
  onRetry?: () => void
}) {
  return (
    <div className="state state--error" role="alert">
      <span className="state__icon" aria-hidden="true">
        <Icon name="alert" size={22} />
      </span>
      <h3 className="state__title">{title}</h3>
      {description && <p className="state__description">{description}</p>}
      {onRetry && (
        <div className="state__action">
          <Button variant="secondary" size="sm" icon="refresh" onClick={onRetry}>
            Reintentar
          </Button>
        </div>
      )}
    </div>
  )
}

export function PageHeader({
  title,
  description,
  actions,
  breadcrumb,
}: {
  title: string
  description?: string
  actions?: ReactNode
  breadcrumb?: ReactNode
}) {
  return (
    <header className="page-header">
      <div className="page-header__text">
        {breadcrumb && <div className="page-header__breadcrumb">{breadcrumb}</div>}
        <h1 className="page-header__title">{title}</h1>
        {description && <p className="page-header__description">{description}</p>}
      </div>
      {actions && <div className="page-header__actions">{actions}</div>}
    </header>
  )
}

export function Callout({
  tone = 'info',
  title,
  children,
}: {
  tone?: 'info' | 'warning' | 'danger' | 'success'
  title?: string
  children: ReactNode
}) {
  const icon: IconName = tone === 'danger' ? 'alert' : tone === 'warning' ? 'alert' : tone === 'success' ? 'check' : 'info'
  return (
    <div className={`callout callout--${tone}`}>
      <Icon name={icon} size={16} />
      <div>
        {title && <strong className="callout__title">{title}</strong>}
        <div className="callout__body">{children}</div>
      </div>
    </div>
  )
}
