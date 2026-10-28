import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { humanizeError, type Session } from '@/api'
import { useRevokeSession, useSessions } from '@/hooks/queries'
import { useAuth } from '@/auth/AuthProvider'
import { useToast } from '@/hooks/useToast'
import {
  Badge,
  Button,
  Card,
  CardBody,
  EmptyState,
  ErrorState,
  PageHeader,
} from '@/components/ui/primitives'
import { DataTable, type Column } from '@/components/ui/data'
import { ConfirmDialog } from '@/components/ui/overlays'
import { formatDateTime, relativeTime } from '@/lib/format'

type SessionState = 'current' | 'active' | 'revoked' | 'expired'

function sessionState(session: Session): SessionState {
  if (session.current) return 'current'
  if (session.revoked_at) return 'revoked'
  if (new Date(session.expires_at).getTime() < Date.now()) return 'expired'
  return 'active'
}

const STATE_LABEL: Record<SessionState, string> = {
  current: 'Sesión actual',
  active: 'Activa',
  revoked: 'Revocada',
  expired: 'Expirada',
}

const STATE_TONE: Record<SessionState, 'primary' | 'success' | 'neutral' | 'warning'> = {
  current: 'primary',
  active: 'success',
  revoked: 'neutral',
  expired: 'warning',
}

export function SessionsPage() {
  const { logout, logoutAll } = useAuth()
  const toast = useToast()
  const navigate = useNavigate()

  const sessions = useSessions()
  const revoke = useRevokeSession()

  const [revoking, setRevoking] = useState<Session | null>(null)
  const [confirmLogoutAll, setConfirmLogoutAll] = useState(false)

  const columns: Column<Session>[] = [
    {
      key: 'device',
      header: 'Dispositivo',
      render: (row) => (
        <div>
          <strong>{row.user_agent || 'Dispositivo no identificado'}</strong>
          <div className="muted text-sm">{row.ip_address || 'IP no registrada'}</div>
        </div>
      ),
    },
    {
      key: 'created',
      header: 'Inicio',
      render: (row) => (
        <span className="text-sm">
          {formatDateTime(row.created_at)}
          <div className="muted">{relativeTime(row.created_at)}</div>
        </span>
      ),
      hideOnMobile: true,
    },
    {
      key: 'expires',
      header: 'Expira',
      render: (row) => <span className="text-sm">{formatDateTime(row.expires_at)}</span>,
      hideOnMobile: true,
    },
    {
      key: 'state',
      header: 'Estado',
      render: (row) => {
        const state = sessionState(row)
        return <Badge tone={STATE_TONE[state]}>{STATE_LABEL[state]}</Badge>
      },
    },
    {
      key: 'actions',
      header: '',
      align: 'right',
      render: (row) => {
        const state = sessionState(row)
        if (state === 'current') {
          return (
            <Button
              variant="ghost"
              size="sm"
              icon="logout"
              onClick={() => void logout().then(() => navigate('/'))}
            >
              Cerrar esta sesión
            </Button>
          )
        }
        if (state === 'active') {
          return (
            <Button variant="ghost" size="sm" icon="trash" onClick={() => setRevoking(row)}>
              Revocar
            </Button>
          )
        }
        return <span className="muted text-sm">Sin acciones</span>
      },
    },
  ]

  return (
    <>
      <PageHeader
        title="Sesiones y seguridad"
        description="Revisa desde qué dispositivos se ha accedido a tu cuenta y cierra las sesiones que no reconozcas."
        actions={
          <Button variant="danger" icon="logout" onClick={() => setConfirmLogoutAll(true)}>
            Cerrar todas las sesiones
          </Button>
        }
      />

      <Card>
        <CardBody>
          {sessions.isError ? (
            <ErrorState description={humanizeError(sessions.error)} onRetry={() => void sessions.refetch()} />
          ) : (
            <>
              <DataTable
                columns={columns}
                rows={sessions.data ?? []}
                rowKey={(row) => row.id}
                loading={sessions.isLoading}
                empty={
                  <EmptyState
                    title="Sin sesiones registradas"
                    description="No hay sesiones asociadas a tu cuenta."
                    icon="shield"
                  />
                }
              />
              <p className="split-note" style={{ marginTop: 8 }}>
                Se listan hasta 100 sesiones, incluidas las revocadas y expiradas. Actualizado{' '}
                {relativeTime(sessions.dataUpdatedAt)}.
              </p>
            </>
          )}
        </CardBody>
      </Card>

      <ConfirmDialog
        open={revoking !== null}
        title="Revocar sesión"
        message={
          <>
            ¿Confirmas revocar la sesión de{' '}
            <strong>{revoking?.user_agent || 'dispositivo no identificado'}</strong>? El dispositivo
            deberá iniciar sesión nuevamente.
          </>
        }
        confirmLabel="Revocar"
        loading={revoke.isPending}
        onCancel={() => setRevoking(null)}
        onConfirm={() => {
          if (!revoking) return
          revoke.mutate(revoking.id, {
            onSuccess: () => {
              toast.success('Sesión revocada')
              setRevoking(null)
            },
            onError: (error) => {
              toast.error('No se pudo revocar', humanizeError(error))
              setRevoking(null)
            },
          })
        }}
      />

      <ConfirmDialog
        open={confirmLogoutAll}
        title="Cerrar todas las sesiones"
        message="Se revocarán todas tus sesiones, incluida la actual. Deberás iniciar sesión nuevamente en todos tus dispositivos."
        confirmLabel="Cerrar todas"
        onCancel={() => setConfirmLogoutAll(false)}
        onConfirm={() => {
          setConfirmLogoutAll(false)
          void logoutAll().then(() => navigate('/'))
        }}
      />
    </>
  )
}
