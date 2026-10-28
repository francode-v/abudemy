// Tipos que reflejan el contrato de la API REST de Abudemy.
// Fuente de verdad: backend/docs/openapi.yaml. No inventar campos.

export type Role = 'mobile_user' | 'institutional_client' | 'org_collaborator' | 'admin'
export type UserStatus = 'active' | 'inactive' | 'suspended'
export type OrgMemberRole = 'owner' | 'admin' | 'editor' | 'viewer'
export type HealthSystem = 'fonasa' | 'isapre' | 'capredena' | 'dipreca' | 'private' | 'other'
export type Sex = 'female' | 'male' | 'other' | 'prefer_not_to_say'
export type LivingSituation =
  | 'alone'
  | 'with_partner'
  | 'with_children'
  | 'with_partner_and_children'
  | 'with_family'
  | 'with_caregiver'
  | 'other'

export interface User {
  id: string
  email: string
  full_name: string
  phone?: string
  birth_date?: string
  comuna?: string
  health_system?: HealthSystem
  sex?: Sex
  living_situation?: LivingSituation
  role: Role
  status: UserStatus
  /** Solo relevante para `institutional_client`: aprobación de un administrador. */
  institutional_approved: boolean
  created_at: string
  updated_at: string
  last_login_at?: string
}

export interface Session {
  id: string
  user_id: string
  user_agent?: string
  ip_address?: string
  created_at: string
  expires_at: string
  revoked_at?: string
  current?: boolean
}

export interface TokenPair {
  access_token: string
  refresh_token: string
  token_type: string
  expires_in: number
  session_id: string
  user: User
}

export interface Pagination {
  page: number
  per_page: number
  total: number
  total_pages: number
}

export interface Page<T> {
  items: T[]
  pagination: Pagination
}

export interface CountByLabel {
  label: string
  count: number
}

export interface TerritorialDistribution {
  comuna: string
  region: string
  count: number
}

export interface AnalyticsOverview {
  services_total: number
  services_active: number
  activities_total: number
  activities_published: number
  activities_upcoming: number
  enrollments_total: number
  enrollments_active: number
  users_total: number
  users_mobile: number
  organizations_total: number
  opportunities_open: number
  assistance_total: number
  assistance_pending: number
  average_occupancy: number
  services_by_category: CountByLabel[]
  activities_by_type: CountByLabel[]
  territorial_distribution: TerritorialDistribution[]
}

export interface ServiceCategory {
  id: string
  name: string
  slug: string
  description?: string
}

export interface ScheduleEntry {
  id?: string
  day_of_week: number
  opens_at: string
  closes_at: string
}

export type ServiceType = 'establishment' | 'resource'
export type ServiceStatus = 'active' | 'inactive' | 'draft'

export interface Service {
  id: string
  organization_id?: string | null
  /** Nombre de la organización que presta el servicio. */
  organization_name?: string
  category_id?: string | null
  category?: ServiceCategory | null
  name: string
  description?: string
  type: ServiceType
  status: ServiceStatus
  address?: string
  comuna?: string
  region?: string
  latitude?: number | null
  longitude?: number | null
  phone?: string
  email?: string
  website?: string
  schedule: ScheduleEntry[]
  created_by?: string | null
  created_at: string
  updated_at: string
}

export interface ServiceInput {
  organization_id?: string | null
  category_id?: string | null
  name: string
  description?: string
  type?: ServiceType
  status?: ServiceStatus
  address?: string
  comuna?: string
  region?: string
  latitude?: number | null
  longitude?: number | null
  phone?: string
  email?: string
  website?: string
  schedule?: ScheduleEntry[]
}

export type ActivityType = 'workshop' | 'event' | 'course' | 'support_group'
export type Modality = 'in_person' | 'online' | 'hybrid'
export type ActivityStatus = 'draft' | 'published' | 'cancelled' | 'finished'

export interface Activity {
  id: string
  organization_id?: string | null
  service_id?: string | null
  category_id?: string | null
  title: string
  description?: string
  type: ActivityType
  modality: Modality
  status: ActivityStatus
  location?: string
  comuna?: string
  region?: string
  latitude?: number | null
  longitude?: number | null
  start_at: string
  end_at?: string | null
  capacity: number
  enrolled_count: number
  price: number
  currency: string
  created_at: string
  updated_at: string
}

export interface ActivityInput {
  organization_id?: string | null
  service_id?: string | null
  category_id?: string | null
  title: string
  description?: string
  type?: ActivityType
  modality?: Modality
  status?: ActivityStatus
  location?: string
  comuna?: string
  region?: string
  latitude?: number | null
  longitude?: number | null
  start_at: string
  end_at?: string | null
  capacity?: number
  price?: number
  currency?: string
}

export type EnrollmentStatus = 'enrolled' | 'waitlist' | 'cancelled' | 'attended'

export interface Enrollment {
  id: string
  activity_id: string
  user_id: string
  status: EnrollmentStatus
  created_at: string
  updated_at: string
  activity_title?: string
  user_email?: string
  user_full_name?: string
}

export type OrganizationType =
  | 'health'
  | 'community'
  | 'ngo'
  | 'employer'
  | 'government'
  | 'educational'
  | 'other'

export interface Organization {
  id: string
  name: string
  type: OrganizationType
  legal_id?: string
  description?: string
  email?: string
  phone?: string
  address?: string
  comuna?: string
  region?: string
  latitude?: number | null
  longitude?: number | null
  active: boolean
  created_at: string
  updated_at: string
}

export interface OrganizationInput {
  name: string
  type?: OrganizationType
  legal_id?: string
  description?: string
  email?: string
  phone?: string
  address?: string
  comuna?: string
  region?: string
  latitude?: number | null
  longitude?: number | null
  active?: boolean
}

export interface OrganizationMembership {
  organization: Organization
  role: OrgMemberRole
}

export interface OrganizationMember {
  id: string
  organization_id: string
  user_id: string
  role: OrgMemberRole
  user_email?: string
  user_full_name?: string
  created_at: string
}

export type OpportunityType = 'job' | 'gig' | 'volunteer' | 'economic'
export type OpportunityStatus = 'draft' | 'open' | 'closed'

export interface Opportunity {
  id: string
  organization_id?: string | null
  title: string
  description?: string
  type: OpportunityType
  status: OpportunityStatus
  comuna?: string
  region?: string
  salary_min?: number | null
  salary_max?: number | null
  currency?: string
  contact_email?: string
  expires_at?: string | null
  created_at: string
  updated_at: string
}

export interface OpportunityInput {
  organization_id?: string | null
  title: string
  description?: string
  type?: OpportunityType
  status?: OpportunityStatus
  comuna?: string
  region?: string
  salary_min?: number | null
  salary_max?: number | null
  currency?: string
  contact_email?: string
  expires_at?: string | null
}

export type ApplicationStatus = 'applied' | 'accepted' | 'rejected'

export interface Application {
  id: string
  opportunity_id: string
  user_id: string
  message?: string
  status: ApplicationStatus
  created_at: string
}

export type AssistanceStatus = 'pending' | 'notified' | 'resolved' | 'cancelled'

export interface AssistanceNotification {
  id: string
  assistance_request_id: string
  contact_id?: string | null
  channel: 'sms' | 'email'
  recipient: string
  provider: string
  status: 'pending' | 'sent' | 'failed' | 'skipped'
  error?: string
  sent_at?: string | null
}

export interface AssistanceRequest {
  id: string
  user_id: string
  type: string
  description?: string
  latitude?: number | null
  longitude?: number | null
  status: AssistanceStatus
  created_at: string
  updated_at: string
  resolved_at?: string | null
  notifications?: AssistanceNotification[]
}

export type ReportType =
  | 'services'
  | 'activities'
  | 'enrollments'
  | 'assistance'
  | 'opportunities'
  | 'territorial'

export interface ReportsCatalog {
  items: string[]
  formats: string[]
}

/** Vista previa de un informe: encabezados y filas tal como se exportan. */
export interface ReportPreview {
  type: string
  headers: string[]
  rows: string[][]
  /** Total de filas del informe (la vista previa puede estar recortada). */
  total?: number
  truncated?: boolean
}

// Error envelope shared by every endpoint.
export interface ApiErrorBody {
  error: {
    code: string
    message: string
    fields?: { field: string; message: string }[]
  }
}
