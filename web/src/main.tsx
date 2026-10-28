import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { AuthProvider } from '@/auth/AuthProvider'
import { OrgScopeProvider } from '@/app/orgScope'
import { ToastProvider } from '@/hooks/useToast'
import { App } from './App'
import './styles.css'

/**
 * Cliente de React Query compartido por toda la aplicación.
 *
 * Se desactiva el reconsultado automático al recuperar el foco (los datos
 * institucionales no requieren esa agresividad) y se limita a un reintento
 * para no ocultar errores reales del backend.
 */
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 1,
      refetchOnWindowFocus: false,
    },
  },
})

const container = document.getElementById('root')
if (!container) {
  throw new Error('No se encontró el elemento #root en index.html')
}

createRoot(container).render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <AuthProvider>
          <OrgScopeProvider>
            <ToastProvider>
              <App />
            </ToastProvider>
          </OrgScopeProvider>
        </AuthProvider>
      </BrowserRouter>
    </QueryClientProvider>
  </StrictMode>,
)
