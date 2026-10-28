import type { OrgMemberRole, Role } from '@/api'

/** Roles con acceso al portal institucional (igual que el backend). */
export const INSTITUTIONAL_ROLES: readonly Role[] = [
  'admin',
  'institutional_client',
  'org_collaborator',
] as const

export function isAdmin(role?: Role | null): boolean {
  return role === 'admin'
}

export function isInstitutional(role?: Role | null, institutionalApproved?: boolean): boolean {
  if (role == null || !INSTITUTIONAL_ROLES.includes(role)) return false
  // Los clientes institucionales recién registrados deben ser aprobados por un
  // administrador. Los roles administrativos y de colaboración no dependen de
  // esta marca. `undefined` se trata como aprobado por compatibilidad.
  if (role === 'institutional_client') return institutionalApproved !== false
  return true
}

/** Solo administradores y clientes institucionales aprobados crean organizaciones. */
export function canCreateOrganization(
  role?: Role | null,
  institutionalApproved?: boolean,
): boolean {
  if (role === 'admin') return true
  return role === 'institutional_client' && institutionalApproved !== false
}

/**
 * Administradores y clientes institucionales aprobados pueden consultar el
 * listado de personas registradas (cuentas móviles).
 */
export function canViewRegisteredPeople(
  role?: Role | null,
  institutionalApproved?: boolean,
): boolean {
  if (role === 'admin') return true
  return role === 'institutional_client' && institutionalApproved !== false
}

/** Un rol de organización puede crear o modificar datos. */
export function canWriteOrgRole(role?: OrgMemberRole | null): boolean {
  return role === 'owner' || role === 'admin' || role === 'editor'
}

/** Un rol de organización puede gestionar integrantes. */
export function canManageMembersRole(role?: OrgMemberRole | null): boolean {
  return role === 'owner' || role === 'admin'
}

export interface MembershipLike {
  role: OrgMemberRole
  organization: { id: string }
}

export interface Permissions {
  isAdmin: boolean
  isInstitutional: boolean
  /** Puede crear organizaciones (admin o cliente institucional aprobado). */
  canCreateOrganization: boolean
  /** Puede ver el listado de personas registradas (admin o cliente institucional aprobado). */
  canViewRegisteredPeople: boolean
  /** Puede crear/editar/eliminar recursos de la organización indicada. */
  canWriteOrganization(organizationId?: string | null): boolean
  /** Puede gestionar integrantes de la organización indicada. */
  canManageOrganization(organizationId?: string | null): boolean
  /** Organizaciones en las que el usuario puede escribir. */
  writableOrganizationIds: string[]
  /** Hay al menos una organización donde el usuario puede escribir. */
  canManageAnyContent: boolean
}

export function buildPermissions(
  role: Role | null | undefined,
  memberships: MembershipLike[],
  institutionalApproved?: boolean,
): Permissions {
  const admin = isAdmin(role)
  // Un cliente institucional pendiente o revocado pierde también los permisos
  // de escritura, aunque conserve membresías antiguas en caché.
  const institutional = isInstitutional(role, institutionalApproved)
  const writable = new Set(
    institutional
      ? memberships.filter((m) => canWriteOrgRole(m.role)).map((m) => m.organization.id)
      : [],
  )
  const manageable = new Set(
    institutional
      ? memberships.filter((m) => canManageMembersRole(m.role)).map((m) => m.organization.id)
      : [],
  )

  return {
    isAdmin: admin,
    isInstitutional: institutional,
    canCreateOrganization: canCreateOrganization(role, institutionalApproved),
    canViewRegisteredPeople: canViewRegisteredPeople(role, institutionalApproved),
    writableOrganizationIds: [...writable],
    canManageAnyContent: admin || writable.size > 0,
    canWriteOrganization(organizationId) {
      if (admin) return true
      if (!institutional || !organizationId) return false // recursos de plataforma: solo admin
      return writable.has(organizationId)
    },
    canManageOrganization(organizationId) {
      if (admin) return true
      if (!institutional || !organizationId) return false
      return manageable.has(organizationId)
    },
  }
}
