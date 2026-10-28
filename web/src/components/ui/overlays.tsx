import { useEffect, useRef, type ReactNode } from 'react'
import { Icon } from './Icon'
import { Button, IconButton, Spinner } from './primitives'
import { useToast } from '@/hooks/useToast'

export function Modal({
  open,
  onClose,
  title,
  description,
  children,
  footer,
  size = 'md',
}: {
  open: boolean
  onClose: () => void
  title: string
  description?: string
  children: ReactNode
  footer?: ReactNode
  size?: 'sm' | 'md' | 'lg' | 'xl'
}) {
  const ref = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!open) return
    const onKey = (event: KeyboardEvent) => {
      if (event.key === 'Escape') onClose()
    }
    document.addEventListener('keydown', onKey)
    const previous = document.body.style.overflow
    document.body.style.overflow = 'hidden'
    ref.current?.focus()
    return () => {
      document.removeEventListener('keydown', onKey)
      document.body.style.overflow = previous
    }
  }, [open, onClose])

  if (!open) return null

  return (
    <div className="overlay" role="presentation" onMouseDown={(e) => e.target === e.currentTarget && onClose()}>
      <div
        ref={ref}
        className={`modal modal--${size}`}
        role="dialog"
        aria-modal="true"
        aria-label={title}
        tabIndex={-1}
      >
        <header className="modal__header">
          <div>
            <h2 className="modal__title">{title}</h2>
            {description && <p className="modal__description">{description}</p>}
          </div>
          <IconButton icon="close" label="Cerrar" onClick={onClose} />
        </header>
        <div className="modal__body">{children}</div>
        {footer && <footer className="modal__footer">{footer}</footer>}
      </div>
    </div>
  )
}

export function ConfirmDialog({
  open,
  title,
  message,
  confirmLabel = 'Confirmar',
  cancelLabel = 'Cancelar',
  tone = 'danger',
  loading,
  onConfirm,
  onCancel,
}: {
  open: boolean
  title: string
  message: ReactNode
  confirmLabel?: string
  cancelLabel?: string
  tone?: 'danger' | 'primary'
  loading?: boolean
  onConfirm: () => void
  onCancel: () => void
}) {
  return (
    <Modal
      open={open}
      onClose={onCancel}
      title={title}
      size="sm"
      footer={
        <>
          <Button variant="secondary" onClick={onCancel} disabled={loading}>
            {cancelLabel}
          </Button>
          <Button variant={tone === 'danger' ? 'danger' : 'primary'} onClick={onConfirm} loading={loading}>
            {confirmLabel}
          </Button>
        </>
      }
    >
      <div className="confirm__message">{message}</div>
    </Modal>
  )
}

export function Toaster() {
  const { toasts, dismiss } = useToast()
  if (toasts.length === 0) return null
  return (
    <div className="toaster" role="region" aria-label="Notificaciones">
      {toasts.map((toast) => (
        <div key={toast.id} className={`toast toast--${toast.tone}`} role="status">
          <span className="toast__icon" aria-hidden="true">
            {toast.tone === 'error' ? (
              <Icon name="alert" size={16} />
            ) : toast.tone === 'success' ? (
              <Icon name="check" size={16} />
            ) : toast.tone === 'warning' ? (
              <Icon name="alert" size={16} />
            ) : (
              <Icon name="info" size={16} />
            )}
          </span>
          <div className="toast__content">
            <strong className="toast__title">{toast.title}</strong>
            {toast.description && <p className="toast__description">{toast.description}</p>}
          </div>
          <button className="toast__close" onClick={() => dismiss(toast.id)} aria-label="Cerrar notificación">
            <Icon name="close" size={14} />
          </button>
        </div>
      ))}
    </div>
  )
}

export function FullPageLoader({ message = 'Cargando…' }: { message?: string }) {
  return (
    <div className="full-loader">
      <Spinner size="lg" />
      <p>{message}</p>
    </div>
  )
}
