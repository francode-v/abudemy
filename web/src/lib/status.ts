import type { ActivityStatus, AssistanceStatus, EnrollmentStatus, OpportunityStatus, ServiceStatus, UserStatus } from '@/api'

export type Tone = 'neutral' | 'success' | 'warning' | 'danger' | 'info' | 'primary'

export function serviceStatusTone(status: ServiceStatus): Tone {
  switch (status) {
    case 'active':
      return 'success'
    case 'draft':
      return 'warning'
    case 'inactive':
      return 'neutral'
    default:
      return 'neutral'
  }
}

export function activityStatusTone(status: ActivityStatus): Tone {
  switch (status) {
    case 'published':
      return 'success'
    case 'draft':
      return 'warning'
    case 'cancelled':
      return 'danger'
    case 'finished':
      return 'info'
    default:
      return 'neutral'
  }
}

export function enrollmentStatusTone(status: EnrollmentStatus): Tone {
  switch (status) {
    case 'enrolled':
      return 'success'
    case 'waitlist':
      return 'warning'
    case 'attended':
      return 'info'
    case 'cancelled':
      return 'danger'
    default:
      return 'neutral'
  }
}

export function assistanceStatusTone(status: AssistanceStatus): Tone {
  switch (status) {
    case 'pending':
      return 'warning'
    case 'notified':
      return 'info'
    case 'resolved':
      return 'success'
    case 'cancelled':
      return 'neutral'
    default:
      return 'neutral'
  }
}

export function userStatusTone(status: UserStatus): Tone {
  switch (status) {
    case 'active':
      return 'success'
    case 'suspended':
      return 'danger'
    case 'inactive':
      return 'neutral'
    default:
      return 'neutral'
  }
}

export function opportunityStatusTone(status: OpportunityStatus): Tone {
  switch (status) {
    case 'open':
      return 'success'
    case 'draft':
      return 'warning'
    case 'closed':
      return 'neutral'
    default:
      return 'neutral'
  }
}
