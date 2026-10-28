import { describe, expect, it } from 'vitest'
import {
  buildPermissions,
  canManageMembersRole,
  canViewRegisteredPeople,
  canWriteOrgRole,
  isAdmin,
  isInstitutional,
} from './permissions'

describe('helpers de roles', () => {
  it('identifica al administrador de plataforma', () => {
    expect(isAdmin('admin')).toBe(true)
    expect(isAdmin('institutional_client')).toBe(false)
    expect(isAdmin(undefined)).toBe(false)
  })

  it('identifica los roles con acceso institucional', () => {
    expect(isInstitutional('admin')).toBe(true)
    expect(isInstitutional('institutional_client')).toBe(true)
    expect(isInstitutional('org_collaborator')).toBe(true)
    expect(isInstitutional('mobile_user')).toBe(false)
    expect(isInstitutional(null)).toBe(false)
  })

  it('un cliente institucional pendiente no se considera institucional', () => {
    expect(isInstitutional('institutional_client', false)).toBe(false)
    expect(isInstitutional('institutional_client', true)).toBe(true)
    // Los roles administrativos no dependen de la aprobación.
    expect(isInstitutional('admin', false)).toBe(true)
  })

  it('determina qué roles pueden ver personas registradas', () => {
    expect(canViewRegisteredPeople('admin')).toBe(true)
    expect(canViewRegisteredPeople('institutional_client')).toBe(true)
    expect(canViewRegisteredPeople('institutional_client', true)).toBe(true)
    expect(canViewRegisteredPeople('institutional_client', false)).toBe(false)
    expect(canViewRegisteredPeople('org_collaborator')).toBe(false)
    expect(canViewRegisteredPeople('mobile_user')).toBe(false)
  })

  it('determina qué roles de organización pueden escribir', () => {
    expect(canWriteOrgRole('owner')).toBe(true)
    expect(canWriteOrgRole('admin')).toBe(true)
    expect(canWriteOrgRole('editor')).toBe(true)
    expect(canWriteOrgRole('viewer')).toBe(false)
    expect(canWriteOrgRole(null)).toBe(false)
  })

  it('determina qué roles de organización gestionan integrantes', () => {
    expect(canManageMembersRole('owner')).toBe(true)
    expect(canManageMembersRole('admin')).toBe(true)
    expect(canManageMembersRole('editor')).toBe(false)
    expect(canManageMembersRole('viewer')).toBe(false)
  })
})

describe('buildPermissions', () => {
  it('un administrador puede escribir cualquier organización, incluso de plataforma', () => {
    const permissions = buildPermissions('admin', [])
    expect(permissions.isAdmin).toBe(true)
    expect(permissions.canWriteOrganization('cualquiera')).toBe(true)
    expect(permissions.canWriteOrganization(null)).toBe(true)
    expect(permissions.canManageAnyContent).toBe(true)
  })

  it('un cliente institucional solo escribe donde tiene un rol de escritura', () => {
    const permissions = buildPermissions('institutional_client', [
      { role: 'editor', organization: { id: 'org-1' } },
      { role: 'viewer', organization: { id: 'org-2' } },
    ])
    expect(permissions.canWriteOrganization('org-1')).toBe(true)
    expect(permissions.canWriteOrganization('org-2')).toBe(false)
    // Los recursos sin organización son de plataforma: solo administradores.
    expect(permissions.canWriteOrganization(null)).toBe(false)
    expect(permissions.writableOrganizationIds).toEqual(['org-1'])
    expect(permissions.canManageAnyContent).toBe(true)
  })

  it('un colaborador sin rol de escritura no puede gestionar contenido', () => {
    const permissions = buildPermissions('org_collaborator', [
      { role: 'viewer', organization: { id: 'org-1' } },
    ])
    expect(permissions.canWriteOrganization('org-1')).toBe(false)
    expect(permissions.canManageAnyContent).toBe(false)
    expect(permissions.canManageOrganization('org-1')).toBe(false)
  })

  it('gestiona integrantes solo con rol propietario o administrador', () => {
    const permissions = buildPermissions('institutional_client', [
      { role: 'owner', organization: { id: 'org-1' } },
      { role: 'editor', organization: { id: 'org-2' } },
    ])
    expect(permissions.canManageOrganization('org-1')).toBe(true)
    expect(permissions.canManageOrganization('org-2')).toBe(false)
  })

  it('un usuario móvil no tiene permisos institucionales', () => {
    const permissions = buildPermissions('mobile_user', [])
    expect(permissions.isInstitutional).toBe(false)
    expect(permissions.canWriteOrganization('org-1')).toBe(false)
    expect(permissions.canManageAnyContent).toBe(false)
    expect(permissions.canCreateOrganization).toBe(false)
  })

  it('solo los clientes institucionales aprobados crean organizaciones', () => {
    expect(buildPermissions('institutional_client', [], false).canCreateOrganization).toBe(false)
    expect(buildPermissions('institutional_client', [], true).canCreateOrganization).toBe(true)
    expect(buildPermissions('org_collaborator', []).canCreateOrganization).toBe(false)
    expect(buildPermissions('admin', []).canCreateOrganization).toBe(true)
  })

  it('los clientes institucionales aprobados ven personas registradas', () => {
    expect(buildPermissions('institutional_client', [], false).canViewRegisteredPeople).toBe(false)
    expect(buildPermissions('institutional_client', [], true).canViewRegisteredPeople).toBe(true)
    expect(buildPermissions('admin', []).canViewRegisteredPeople).toBe(true)
    expect(buildPermissions('org_collaborator', []).canViewRegisteredPeople).toBe(false)
    expect(buildPermissions('mobile_user', []).canViewRegisteredPeople).toBe(false)
  })

  it('un cliente pendiente pierde la escritura aunque conserve membresías', () => {
    const memberships = [{ role: 'owner' as const, organization: { id: 'org-1' } }]
    const pending = buildPermissions('institutional_client', memberships, false)
    expect(pending.isInstitutional).toBe(false)
    expect(pending.canWriteOrganization('org-1')).toBe(false)
    expect(pending.canManageOrganization('org-1')).toBe(false)
    expect(pending.writableOrganizationIds).toEqual([])
    expect(pending.canManageAnyContent).toBe(false)

    const approved = buildPermissions('institutional_client', memberships, true)
    expect(approved.canWriteOrganization('org-1')).toBe(true)
    expect(approved.canManageOrganization('org-1')).toBe(true)
  })
})
