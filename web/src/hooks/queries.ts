import {
  useMutation,
  useQuery,
  useQueryClient,
  type UseMutationOptions,
  type UseQueryOptions,
} from '@tanstack/react-query'
import { api } from '@/api'
import type {
  Activity,
  ActivityInput,
  AnalyticsOverview,
  AssistanceRequest,
  Enrollment,
  EnrollmentStatus,
  OrgMemberRole,
  Organization,
  OrganizationInput,
  OrganizationMember,
  OrganizationMembership,
  Opportunity,
  OpportunityInput,
  Page,
  ReportPreview,
  ReportType,
  Role,
  Service,
  ServiceCategory,
  ServiceInput,
  Session,
  User,
  UserStatus,
} from '@/api'
import type {
  ActivityQuery,
  AssistanceQuery,
  OpportunityQuery,
  OrganizationQuery,
  ServiceQuery,
  UserQuery,
} from '@/api/endpoints'

/** Claves de caché centralizadas para invalidaciones coherentes. */
export const queryKeys = {
  me: ['me'] as const,
  memberships: ['memberships'] as const,
  sessions: ['sessions'] as const,
  categories: ['service-categories'] as const,
  services: (query?: ServiceQuery) => ['services', query ?? {}] as const,
  service: (id: string) => ['service', id] as const,
  activities: (query?: ActivityQuery) => ['activities', query ?? {}] as const,
  activity: (id: string) => ['activity', id] as const,
  activityEnrollments: (id: string, page: number) => ['activity-enrollments', id, page] as const,
  organizations: (query?: OrganizationQuery) => ['organizations', query ?? {}] as const,
  organization: (id: string) => ['organization', id] as const,
  organizationMembers: (id: string) => ['organization-members', id] as const,
  users: (query?: UserQuery) => ['users', query ?? {}] as const,
  user: (id: string) => ['user', id] as const,
  overview: (organizationId?: string | null) => ['analytics-overview', organizationId ?? 'global'] as const,
  reports: ['reports-catalog'] as const,
  reportPreview: (type: ReportType, organizationId?: string | null) =>
    ['report-preview', type, organizationId ?? 'global'] as const,
  assistance: (query?: AssistanceQuery) => ['assistance', query ?? {}] as const,
  opportunities: (query?: OpportunityQuery) => ['opportunities', query ?? {}] as const,
}

const STALE = {
  short: 30_000,
  medium: 5 * 60_000,
  long: 30 * 60_000,
}

// ---------------------------------------------------------------- lectura

export function useMe(options?: Partial<UseQueryOptions<User>>) {
  return useQuery({ queryKey: queryKeys.me, queryFn: () => api.users.me(), staleTime: STALE.medium, ...options })
}

export function useMemberships() {
  return useQuery({
    queryKey: queryKeys.memberships,
    queryFn: () => api.organizations.mine(),
    staleTime: STALE.medium,
  })
}

export function useSessions() {
  return useQuery({
    queryKey: queryKeys.sessions,
    queryFn: () => api.auth.sessions(),
    select: (data: { items: Session[] }) => data.items,
  })
}

export function useCategories() {
  return useQuery({
    queryKey: queryKeys.categories,
    queryFn: () => api.categories.list(),
    select: (data: { items: ServiceCategory[] }) => data.items,
    staleTime: STALE.long,
  })
}

export function useServices(query: ServiceQuery, enabled = true) {
  return useQuery({
    queryKey: queryKeys.services(query),
    queryFn: () => api.services.list(query),
    enabled,
    staleTime: STALE.short,
    placeholderData: (previous: Page<Service> | undefined) => previous,
  })
}

export function useService(id?: string) {
  return useQuery({
    queryKey: queryKeys.service(id ?? ''),
    queryFn: () => api.services.get(id as string),
    enabled: Boolean(id),
    staleTime: STALE.short,
  })
}

export function useActivities(query: ActivityQuery, enabled = true) {
  return useQuery({
    queryKey: queryKeys.activities(query),
    queryFn: () => api.activities.list(query),
    enabled,
    staleTime: STALE.short,
    placeholderData: (previous: Page<Activity> | undefined) => previous,
  })
}

export function useActivity(id?: string) {
  return useQuery({
    queryKey: queryKeys.activity(id ?? ''),
    queryFn: () => api.activities.get(id as string),
    enabled: Boolean(id),
    staleTime: STALE.short,
  })
}

export function useActivityEnrollments(id?: string, page = 1, perPage = 20) {
  return useQuery({
    queryKey: queryKeys.activityEnrollments(id ?? '', page),
    queryFn: () => api.activities.enrollments(id as string, { page, per_page: perPage }),
    enabled: Boolean(id),
  })
}

export function useOrganizations(query: OrganizationQuery = {}, enabled = true) {
  return useQuery({
    queryKey: queryKeys.organizations(query),
    queryFn: () => api.organizations.list(query),
    enabled,
    staleTime: STALE.medium,
    placeholderData: (previous: Page<Organization> | undefined) => previous,
  })
}

export function useOrganization(id?: string) {
  return useQuery({
    queryKey: queryKeys.organization(id ?? ''),
    queryFn: () => api.organizations.get(id as string),
    enabled: Boolean(id),
  })
}

export function useOrganizationMembers(id?: string, enabled = true) {
  return useQuery({
    queryKey: queryKeys.organizationMembers(id ?? ''),
    queryFn: () => api.organizations.members(id as string),
    select: (data: { items: OrganizationMember[] }) => data.items,
    enabled: Boolean(id) && enabled,
  })
}

export function useUsers(query: UserQuery, enabled = true) {
  return useQuery({
    queryKey: queryKeys.users(query),
    queryFn: () => api.users.list(query),
    enabled,
    placeholderData: (previous: Page<User> | undefined) => previous,
  })
}

export function useOverview(organizationId?: string | null, enabled = true) {
  return useQuery<AnalyticsOverview>({
    queryKey: queryKeys.overview(organizationId),
    queryFn: () => api.analytics.overview(organizationId),
    enabled,
    staleTime: STALE.short,
  })
}

export function useReportsCatalog() {
  return useQuery({
    queryKey: queryKeys.reports,
    queryFn: () => api.analytics.reports(),
    staleTime: STALE.long,
  })
}

export function useReportPreview(
  type: ReportType,
  organizationId?: string | null,
  enabled = true,
) {
  return useQuery<ReportPreview>({
    queryKey: queryKeys.reportPreview(type, organizationId),
    queryFn: () => api.analytics.previewReport(type, organizationId),
    enabled,
    staleTime: STALE.short,
    placeholderData: (previous: ReportPreview | undefined) => previous,
  })
}

export function useAssistanceRequests(query: AssistanceQuery = {}, enabled = true) {
  return useQuery({
    queryKey: queryKeys.assistance(query),
    queryFn: () => api.assistance.requests(query),
    enabled,
    placeholderData: (previous: Page<AssistanceRequest> | undefined) => previous,
  })
}

export function useOpportunities(query: OpportunityQuery = {}, enabled = true) {
  return useQuery({
    queryKey: queryKeys.opportunities(query),
    queryFn: () => api.opportunities.list(query),
    enabled,
  })
}

export function useMyEnrollments() {
  return useQuery({
    queryKey: ['my-enrollments'],
    queryFn: () => api.users.myEnrollments({ per_page: 50 }),
    select: (data: Page<Enrollment>) => data.items,
  })
}

// ---------------------------------------------------------------- mutaciones

function useInvalidate() {
  const client = useQueryClient()
  return (...keys: (readonly unknown[])[]) => {
    for (const key of keys) void client.invalidateQueries({ queryKey: key })
  }
}

export function useSaveService(options?: UseMutationOptions<Service, unknown, { id?: string; input: ServiceInput }>) {
  const invalidate = useInvalidate()
  return useMutation({
    mutationFn: ({ id, input }: { id?: string; input: ServiceInput }) =>
      id ? api.services.update(id, input) : api.services.create(input),
    ...options,
    onSuccess: (data, variables, onMutateResult, context) => {
      invalidate(['services'], ['analytics-overview'], ['service'])
      options?.onSuccess?.(data, variables, onMutateResult, context)
    },
  })
}

export function useServiceStatus() {
  const invalidate = useInvalidate()
  return useMutation({
    mutationFn: ({ id, status }: { id: string; status: Service['status'] }) =>
      api.services.setStatus(id, status),
    onSuccess: () => {
      invalidate(['services'], ['analytics-overview'], ['service'])
    },
  })
}

export function useDeleteService() {
  const invalidate = useInvalidate()
  return useMutation({
    mutationFn: (id: string) => api.services.remove(id),
    onSuccess: () => {
      invalidate(['services'], ['analytics-overview'])
    },
  })
}

export function useSaveActivity(
  options?: UseMutationOptions<Activity, unknown, { id?: string; input: ActivityInput }>,
) {
  const invalidate = useInvalidate()
  return useMutation({
    mutationFn: ({ id, input }: { id?: string; input: ActivityInput }) =>
      id ? api.activities.update(id, input) : api.activities.create(input),
    ...options,
    onSuccess: (data, variables, onMutateResult, context) => {
      invalidate(['activities'], ['analytics-overview'], ['activity'])
      options?.onSuccess?.(data, variables, onMutateResult, context)
    },
  })
}

export function useActivityStatus() {
  const invalidate = useInvalidate()
  return useMutation({
    mutationFn: ({ id, status }: { id: string; status: Activity['status'] }) =>
      api.activities.setStatus(id, status),
    onSuccess: () => {
      invalidate(['activities'], ['analytics-overview'], ['activity'])
    },
  })
}

export function useEnrollmentStatus(activityId: string) {
  const invalidate = useInvalidate()
  return useMutation({
    mutationFn: ({ userId, status }: { userId: string; status: EnrollmentStatus }) =>
      api.activities.updateEnrollment(activityId, userId, status),
    onSuccess: () => {
      invalidate(['activity-enrollments'], ['activities'], ['analytics-overview'])
    },
  })
}

export function useSaveOrganization(
  options?: UseMutationOptions<Organization, unknown, { id?: string; input: OrganizationInput }>,
) {
  const invalidate = useInvalidate()
  return useMutation({
    mutationFn: ({ id, input }: { id?: string; input: OrganizationInput }) =>
      id ? api.organizations.update(id, input) : api.organizations.create(input),
    ...options,
    onSuccess: (data, variables, onMutateResult, context) => {
      invalidate(['organizations'], ['memberships'])
      options?.onSuccess?.(data, variables, onMutateResult, context)
    },
  })
}

export function useOrganizationMemberMutations(organizationId: string) {
  const invalidate = useInvalidate()
  const add = useMutation({
    mutationFn: (payload: { email?: string; user_id?: string; role: OrgMemberRole }) =>
      api.organizations.addMember(organizationId, payload),
    onSuccess: () => invalidate(['organization-members']),
  })
  const updateRole = useMutation({
    mutationFn: ({ userId, role }: { userId: string; role: OrgMemberRole }) =>
      api.organizations.updateMember(organizationId, userId, role),
    onSuccess: () => invalidate(['organization-members']),
  })
  const remove = useMutation({
    mutationFn: (userId: string) => api.organizations.removeMember(organizationId, userId),
    onSuccess: () => invalidate(['organization-members']),
  })
  return { add, updateRole, remove }
}

export function useUserAdminMutations() {
  const invalidate = useInvalidate()
  const updateStatus = useMutation({
    mutationFn: ({ id, status }: { id: string; status: UserStatus }) =>
      api.users.updateStatus(id, status),
    onSuccess: () => invalidate(['users']),
  })
  const updateRole = useMutation({
    mutationFn: ({ id, role }: { id: string; role: User['role'] }) => api.users.updateRole(id, role),
    onSuccess: () => invalidate(['users']),
  })
  const setInstitutionalApproval = useMutation({
    mutationFn: ({ id, approved }: { id: string; approved: boolean }) =>
      api.users.setInstitutionalApproval(id, approved),
    onSuccess: () => invalidate(['users'], ['me']),
  })
  return { updateStatus, updateRole, setInstitutionalApproval }
}

export function useUpdateProfile() {
  const invalidate = useInvalidate()
  return useMutation({
    mutationFn: (payload: { full_name: string; phone?: string; birth_date?: string }) =>
      api.users.updateMe(payload),
    onSuccess: () => invalidate(['me']),
  })
}

export function useChangePassword() {
  return useMutation({
    mutationFn: (payload: { currentPassword: string; newPassword: string }) =>
      api.users.changePassword(payload.currentPassword, payload.newPassword),
  })
}

export function useRevokeSession() {
  const invalidate = useInvalidate()
  return useMutation({
    mutationFn: (id: string) => api.auth.revokeSession(id),
    onSuccess: () => invalidate(['sessions']),
  })
}

export function useRegister() {
  return useMutation({
    mutationFn: (payload: {
      email: string
      password: string
      full_name: string
      phone?: string
      role?: Role
    }) => api.auth.register(payload),
  })
}

export function useUpdateAssistanceStatus() {
  const invalidate = useInvalidate()
  return useMutation({
    mutationFn: ({ id, status }: { id: string; status: AssistanceRequest['status'] }) =>
      api.assistance.updateStatus(id, status),
    onSuccess: () => invalidate(['assistance'], ['analytics-overview']),
  })
}

export function useUpdateOpportunityStatus() {
  const invalidate = useInvalidate()
  return useMutation({
    mutationFn: ({ id, status }: { id: string; status: Opportunity['status'] }) =>
      api.opportunities.setStatus(id, status),
    onSuccess: () => invalidate(['opportunities']),
  })
}

export function useSaveOpportunity() {
  const invalidate = useInvalidate()
  return useMutation({
    mutationFn: ({ id, input }: { id?: string; input: OpportunityInput }) =>
      id ? api.opportunities.update(id, input) : api.opportunities.create(input),
    onSuccess: () => invalidate(['opportunities']),
  })
}

export type { OrganizationMembership, ReportType }
export { api }
