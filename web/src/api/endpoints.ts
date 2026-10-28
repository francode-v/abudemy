import type { ApiClient, RequestOptions } from './client'
import type {
  Activity,
  ActivityInput,
  AnalyticsOverview,
  Application,
  AssistanceRequest,
  Enrollment,
  EnrollmentStatus,
  Organization,
  OrganizationInput,
  OrganizationMember,
  OrganizationMembership,
  OrgMemberRole,
  Opportunity,
  OpportunityInput,
  Page,
  ReportPreview,
  ReportType,
  ReportsCatalog,
  Role,
  Service,
  ServiceCategory,
  ServiceInput,
  Session,
  TokenPair,
  User,
  UserStatus,
} from './types'

/** Filtros aceptados por `GET /services`. */
export type ServiceQuery = {
  category_id?: string
  type?: string
  status?: string
  comuna?: string
  region?: string
  search?: string
  organization_id?: string
  lat?: number
  lng?: number
  radius_km?: number
  page?: number
  per_page?: number
}

export type ActivityQuery = {
  type?: string
  status?: string
  modality?: string
  comuna?: string
  region?: string
  search?: string
  organization_id?: string
  from?: string
  to?: string
  page?: number
  per_page?: number
}

export type OpportunityQuery = {
  type?: string
  status?: string
  comuna?: string
  region?: string
  search?: string
  page?: number
  per_page?: number
}

export type OrganizationQuery = {
  type?: string
  search?: string
  active?: boolean
  page?: number
  per_page?: number
}

export type UserQuery = {
  role?: string
  status?: string
  search?: string
  comuna?: string
  health_system?: string
  sex?: string
  living_situation?: string
  age_min?: number
  age_max?: number
  page?: number
  per_page?: number
}

export type AssistanceQuery = {
  status?: string
  from?: string
  to?: string
  page?: number
  per_page?: number
}

/**
 * Capa de servicios: centraliza todas las llamadas HTTP al backend de Abudemy.
 * Los componentes nunca construyen URLs ni conocen los detalles del transporte.
 */
export class Api {
  constructor(private readonly client: ApiClient) {}

  readonly auth = {
    login: (email: string, password: string) =>
      this.client.post<TokenPair>('/api/v1/auth/login', { email, password }),
    register: (payload: {
      email: string
      password: string
      full_name: string
      phone?: string
      role?: Role
    }) => this.client.post<User>('/api/v1/auth/register', payload),
    refresh: (refreshToken: string) =>
      this.client.post<TokenPair>('/api/v1/auth/refresh', { refresh_token: refreshToken }),
    logout: () => this.client.post<void>('/api/v1/auth/logout', undefined, { responseType: 'none' }),
    logoutAll: () =>
      this.client.post<void>('/api/v1/auth/logout-all', undefined, { responseType: 'none' }),
    sessions: () => this.client.get<{ items: Session[] }>('/api/v1/auth/sessions'),
    revokeSession: (id: string) =>
      this.client.del<void>(`/api/v1/auth/sessions/${id}`, { responseType: 'none' }),
    forgotPassword: (email: string) =>
      this.client.post<{ message: string }>('/api/v1/auth/password/forgot', { email }),
    resetPassword: (token: string, newPassword: string) =>
      this.client.post<void>(
        '/api/v1/auth/password/reset',
        { token, new_password: newPassword },
        { responseType: 'none' },
      ),
  }

  readonly users = {
    me: () => this.client.get<User>('/api/v1/users/me'),
    updateMe: (payload: { full_name: string; phone?: string; birth_date?: string }) =>
      this.client.patch<User>('/api/v1/users/me', payload),
    changePassword: (currentPassword: string, newPassword: string) =>
      this.client.post<void>(
        '/api/v1/users/me/password',
        { current_password: currentPassword, new_password: newPassword },
        { responseType: 'none' },
      ),
    myEnrollments: (query?: { page?: number; per_page?: number }) =>
      this.client.get<Page<Enrollment>>('/api/v1/users/me/enrollments', { query }),
    list: (query?: UserQuery) => this.client.get<Page<User>>('/api/v1/users', { query }),
    get: (id: string) => this.client.get<User>(`/api/v1/users/${id}`),
    updateStatus: (id: string, status: UserStatus) =>
      this.client.patch<User>(`/api/v1/users/${id}/status`, { status }),
    updateRole: (id: string, role: Role) =>
      this.client.patch<User>(`/api/v1/users/${id}/role`, { role }),
    setInstitutionalApproval: (id: string, approved: boolean) =>
      this.client.patch<User>(`/api/v1/users/${id}/institutional-approval`, { approved }),
  }

  readonly categories = {
    list: () => this.client.get<{ items: ServiceCategory[] }>('/api/v1/service-categories'),
  }

  readonly services = {
    list: (query?: ServiceQuery) => this.client.get<Page<Service>>('/api/v1/services', { query }),
    get: (id: string) => this.client.get<Service>(`/api/v1/services/${id}`),
    create: (input: ServiceInput) => this.client.post<Service>('/api/v1/services', input),
    update: (id: string, input: ServiceInput) =>
      this.client.put<Service>(`/api/v1/services/${id}`, input),
    setStatus: (id: string, status: Service['status']) =>
      this.client.patch<Service>(`/api/v1/services/${id}/status`, { status }),
    remove: (id: string) =>
      this.client.del<void>(`/api/v1/services/${id}`, { responseType: 'none' }),
  }

  readonly activities = {
    list: (query?: ActivityQuery) => this.client.get<Page<Activity>>('/api/v1/activities', { query }),
    get: (id: string) => this.client.get<Activity>(`/api/v1/activities/${id}`),
    create: (input: ActivityInput) => this.client.post<Activity>('/api/v1/activities', input),
    update: (id: string, input: ActivityInput) =>
      this.client.put<Activity>(`/api/v1/activities/${id}`, input),
    setStatus: (id: string, status: Activity['status']) =>
      this.client.patch<Activity>(`/api/v1/activities/${id}/status`, { status }),
    enrollments: (id: string, query?: { page?: number; per_page?: number }) =>
      this.client.get<Page<Enrollment>>(`/api/v1/activities/${id}/enrollments`, { query }),
    updateEnrollment: (activityId: string, userId: string, status: EnrollmentStatus) =>
      this.client.patch<void>(
        `/api/v1/activities/${activityId}/enrollments/${userId}`,
        { status },
        { responseType: 'none' },
      ),
  }

  readonly organizations = {
    list: (query?: OrganizationQuery) =>
      this.client.get<Page<Organization>>('/api/v1/organizations', { query }),
    mine: () =>
      this.client.get<{ items: OrganizationMembership[] }>('/api/v1/organizations/mine'),
    get: (id: string) => this.client.get<Organization>(`/api/v1/organizations/${id}`),
    create: (input: OrganizationInput) =>
      this.client.post<Organization>('/api/v1/organizations', input),
    update: (id: string, input: OrganizationInput) =>
      this.client.put<Organization>(`/api/v1/organizations/${id}`, input),
    members: (id: string) =>
      this.client.get<{ items: OrganizationMember[] }>(`/api/v1/organizations/${id}/members`),
    addMember: (id: string, payload: { user_id?: string; email?: string; role: OrgMemberRole }) =>
      this.client.post<OrganizationMember>(`/api/v1/organizations/${id}/members`, payload),
    updateMember: (id: string, userId: string, role: OrgMemberRole) =>
      this.client.patch<void>(
        `/api/v1/organizations/${id}/members/${userId}`,
        { role },
        { responseType: 'none' },
      ),
    removeMember: (id: string, userId: string) =>
      this.client.del<void>(`/api/v1/organizations/${id}/members/${userId}`, {
        responseType: 'none',
      }),
  }

  readonly opportunities = {
    list: (query?: OpportunityQuery) =>
      this.client.get<Page<Opportunity>>('/api/v1/opportunities', { query }),
    get: (id: string) => this.client.get<Opportunity>(`/api/v1/opportunities/${id}`),
    create: (input: OpportunityInput) =>
      this.client.post<Opportunity>('/api/v1/opportunities', input),
    update: (id: string, input: OpportunityInput) =>
      this.client.put<Opportunity>(`/api/v1/opportunities/${id}`, input),
    setStatus: (id: string, status: Opportunity['status']) =>
      this.client.patch<void>(`/api/v1/opportunities/${id}/status`, { status }, { responseType: 'none' }),
    applications: (id: string, query?: { page?: number; per_page?: number }) =>
      this.client.get<Page<Application>>(`/api/v1/opportunities/${id}/applications`, { query }),
    updateApplication: (id: string, appId: string, status: Application['status']) =>
      this.client.patch<void>(
        `/api/v1/opportunities/${id}/applications/${appId}`,
        { status },
        { responseType: 'none' },
      ),
  }

  readonly assistance = {
    requests: (query?: AssistanceQuery) =>
      this.client.get<Page<AssistanceRequest>>('/api/v1/assistance/requests', { query }),
    get: (id: string) => this.client.get<AssistanceRequest>(`/api/v1/assistance/requests/${id}`),
    updateStatus: (id: string, status: AssistanceRequest['status']) =>
      this.client.patch<AssistanceRequest>(`/api/v1/assistance/requests/${id}/status`, { status }),
  }

  readonly analytics = {
    overview: (organizationId?: string | null) =>
      this.client.get<AnalyticsOverview>('/api/v1/analytics/overview', {
        query: { organization_id: organizationId },
      }),
    reports: () => this.client.get<ReportsCatalog>('/api/v1/analytics/reports'),
    /** Devuelve el informe seleccionado como JSON para mostrarlo en pantalla. */
    previewReport: (type: ReportType, organizationId?: string | null) =>
      this.client.get<ReportPreview>(`/api/v1/analytics/reports/${type}`, {
        query: { organization_id: organizationId, format: 'json' },
      }),
    /** Descarga un informe como CSV. */
    exportCsv: (type: ReportType, organizationId?: string, options?: RequestOptions) =>
      this.client.get<Blob>(`/api/v1/analytics/reports/${type}`, {
        query: { organization_id: organizationId, format: 'csv' },
        responseType: 'blob',
        ...options,
      }),
  }
}
