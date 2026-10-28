import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { api, apiClient, sessionEvents, tokenStore } from '@/api'
import type { OrganizationMembership, TokenPair, User } from '@/api'
import { buildPermissions, type Permissions } from './permissions'

interface AuthState {
  status: 'loading' | 'authenticated' | 'anonymous'
  user: User | null
  memberships: OrganizationMembership[]
  permissions: Permissions
}

interface AuthContextValue extends AuthState {
  login(email: string, password: string): Promise<User>
  logout(): Promise<void>
  logoutAll(): Promise<void>
  reload(): Promise<void>
  refreshMemberships(): Promise<void>
  applyTokenPair(pair: TokenPair): void
}

const anonymousState: AuthState = {
  status: 'anonymous',
  user: null,
  memberships: [],
  permissions: buildPermissions(null, []),
}

const AuthContext = createContext<AuthContextValue | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [state, setState] = useState<AuthState>({
    ...anonymousState,
    status: 'loading',
  })
  const queryClient = useQueryClient()

  const loadSession = useCallback(async () => {
    if (!tokenStore.hasSession()) {
      setState(anonymousState)
      return
    }
    try {
      const user = await api.users.me()
      const memberships = await loadMemberships()
      setState({
        status: 'authenticated',
        user,
        memberships,
        permissions: buildPermissions(user.role, memberships, user.institutional_approved),
      })
    } catch {
      tokenStore.clear()
      setState(anonymousState)
    }
  }, [])

  useEffect(() => {
    void loadSession()
  }, [loadSession])

  // La expiración del refresh token cierra la sesión de forma transparente.
  useEffect(() => {
    return sessionEvents.onExpired(() => {
      queryClient.clear()
      setState(anonymousState)
    })
  }, [queryClient])

  const applyTokenPair = useCallback(
    (pair: TokenPair) => {
      tokenStore.setTokens(pair)
      setState((prev) => {
        const user = pair.user ?? prev.user
        return {
          status: 'authenticated',
          user,
          memberships: prev.memberships,
          permissions: buildPermissions(
            user?.role,
            prev.memberships,
            user?.institutional_approved,
          ),
        }
      })
    },
    [],
  )

  const login = useCallback(
    async (email: string, password: string) => {
      const pair = await api.auth.login(email, password)
      tokenStore.setTokens(pair)
      const memberships = await loadMemberships()
      setState({
        status: 'authenticated',
        user: pair.user,
        memberships,
        permissions: buildPermissions(
          pair.user?.role,
          memberships,
          pair.user?.institutional_approved,
        ),
      })
      return pair.user
    },
    [],
  )

  const logout = useCallback(async () => {
    try {
      await api.auth.logout()
    } catch {
      /* la sesión local se cierra igual */
    }
    tokenStore.clear()
    queryClient.clear()
    setState(anonymousState)
  }, [queryClient])

  const logoutAll = useCallback(async () => {
    try {
      await api.auth.logoutAll()
    } catch {
      /* ignorar */
    }
    tokenStore.clear()
    queryClient.clear()
    setState(anonymousState)
  }, [queryClient])

  const refreshMemberships = useCallback(async () => {
    const memberships = await loadMemberships()
    setState((prev) => ({
      ...prev,
      memberships,
      permissions: buildPermissions(prev.user?.role, memberships, prev.user?.institutional_approved),
    }))
  }, [])

  const reload = useCallback(async () => {
    await loadSession()
  }, [loadSession])

  const value = useMemo<AuthContextValue>(
    () => ({
      ...state,
      login,
      logout,
      logoutAll,
      reload,
      refreshMemberships,
      applyTokenPair,
    }),
    [state, login, logout, logoutAll, reload, refreshMemberships, applyTokenPair],
  )

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

async function loadMemberships(): Promise<OrganizationMembership[]> {
  try {
    const result = await api.organizations.mine()
    return result.items ?? []
  } catch {
    return []
  }
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth debe usarse dentro de <AuthProvider>')
  return ctx
}

export { apiClient }
