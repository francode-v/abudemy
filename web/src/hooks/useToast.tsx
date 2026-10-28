import {
  createContext,
  useCallback,
  useContext,
  useMemo,
  useRef,
  useState,
  type ReactNode,
} from 'react'

export type ToastTone = 'success' | 'error' | 'info' | 'warning'

export interface Toast {
  id: number
  tone: ToastTone
  title: string
  description?: string
}

interface ToastContextValue {
  toasts: Toast[]
  notify(toast: Omit<Toast, 'id'>): void
  success(title: string, description?: string): void
  error(title: string, description?: string): void
  info(title: string, description?: string): void
  warning(title: string, description?: string): void
  dismiss(id: number): void
}

const ToastContext = createContext<ToastContextValue | null>(null)

let counter = 0

export function ToastProvider({ children }: { children: ReactNode }) {
  const [toasts, setToasts] = useState<Toast[]>([])
  const timers = useRef(new Map<number, number>())

  const dismiss = useCallback((id: number) => {
    const timer = timers.current.get(id)
    if (timer) {
      window.clearTimeout(timer)
      timers.current.delete(id)
    }
    setToasts((prev) => prev.filter((toast) => toast.id !== id))
  }, [])

  const notify = useCallback(
    (toast: Omit<Toast, 'id'>) => {
      counter += 1
      const id = counter
      setToasts((prev) => [...prev.slice(-3), { ...toast, id }])
      const timer = window.setTimeout(() => dismiss(id), toast.tone === 'error' ? 8000 : 5000)
      timers.current.set(id, timer)
    },
    [dismiss],
  )

  const value = useMemo<ToastContextValue>(
    () => ({
      toasts,
      notify,
      dismiss,
      success: (title, description) => notify({ tone: 'success', title, description }),
      error: (title, description) => notify({ tone: 'error', title, description }),
      info: (title, description) => notify({ tone: 'info', title, description }),
      warning: (title, description) => notify({ tone: 'warning', title, description }),
    }),
    [toasts, notify, dismiss],
  )

  return <ToastContext.Provider value={value}>{children}</ToastContext.Provider>
}

export function useToast(): ToastContextValue {
  const ctx = useContext(ToastContext)
  if (!ctx) throw new Error('useToast debe usarse dentro de <ToastProvider>')
  return ctx
}
