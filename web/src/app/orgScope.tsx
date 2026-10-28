import { createContext, useCallback, useContext, useEffect, useMemo, useState, type ReactNode } from 'react'
import { useAuth } from '@/auth/AuthProvider'

const STORAGE_KEY = 'abudemy.active_organization'

interface OrgScopeValue {
  /** `null` significa "todas las organizaciones accesibles". */
  organizationId: string | null
  setOrganizationId(id: string | null): void
  /** Organizaciones que el usuario puede seleccionar. */
  options: { id: string; name: string; role: string }[]
  activeName: string
}

const OrgScopeContext = createContext<OrgScopeValue | null>(null)

export function OrgScopeProvider({ children }: { children: ReactNode }) {
  const { memberships } = useAuth()
  const [organizationId, setOrganizationIdState] = useState<string | null>(() => {
    try {
      return window.localStorage.getItem(STORAGE_KEY)
    } catch {
      return null
    }
  })

  const options = useMemo(
    () =>
      memberships.map((membership) => ({
        id: membership.organization.id,
        name: membership.organization.name,
        role: membership.role,
      })),
    [memberships],
  )

  // Si la organización guardada ya no es accesible, se limpia el filtro.
  useEffect(() => {
    if (organizationId && !options.some((option) => option.id === organizationId)) {
      setOrganizationIdState(null)
    }
  }, [options, organizationId])

  const setOrganizationId = useCallback((id: string | null) => {
    setOrganizationIdState(id)
    try {
      if (id) window.localStorage.setItem(STORAGE_KEY, id)
      else window.localStorage.removeItem(STORAGE_KEY)
    } catch {
      /* almacenamiento no disponible */
    }
  }, [])

  const value = useMemo<OrgScopeValue>(
    () => ({
      organizationId,
      setOrganizationId,
      options,
      activeName:
        options.find((option) => option.id === organizationId)?.name ?? 'Todas mis organizaciones',
    }),
    [organizationId, setOrganizationId, options],
  )

  return <OrgScopeContext.Provider value={value}>{children}</OrgScopeContext.Provider>
}

export function useOrgScope(): OrgScopeValue {
  const ctx = useContext(OrgScopeContext)
  if (!ctx) throw new Error('useOrgScope debe usarse dentro de <OrgScopeProvider>')
  return ctx
}
