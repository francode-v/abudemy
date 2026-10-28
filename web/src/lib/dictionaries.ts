import type {
  ActivityStatus,
  ActivityType,
  AssistanceStatus,
  EnrollmentStatus,
  HealthSystem,
  LivingSituation,
  Modality,
  OpportunityStatus,
  OpportunityType,
  OrgMemberRole,
  OrganizationType,
  ReportType,
  Role,
  ServiceStatus,
  ServiceType,
  Sex,
  UserStatus,
} from '@/api'

/** Diccionarios de traducción de los valores de la API al español. */

export const roleLabels: Record<Role, string> = {
  mobile_user: 'Persona mayor (app móvil)',
  institutional_client: 'Cliente institucional',
  org_collaborator: 'Colaborador de organización',
  admin: 'Administrador de plataforma',
}

export const userStatusLabels: Record<UserStatus, string> = {
  active: 'Activo',
  inactive: 'Inactivo',
  suspended: 'Suspendido',
}

export const healthSystemLabels: Record<HealthSystem, string> = {
  fonasa: 'FONASA',
  isapre: 'Isapre',
  capredena: 'CAPREDENA',
  dipreca: 'DIPRECA',
  private: 'Particular (sin previsión)',
  other: 'Otro / no estoy seguro',
}

export const sexLabels: Record<Sex, string> = {
  female: 'Femenino',
  male: 'Masculino',
  other: 'Otro',
  prefer_not_to_say: 'Prefiero no decirlo',
}

export const livingSituationLabels: Record<LivingSituation, string> = {
  alone: 'Vivo solo/a',
  with_partner: 'Con mi pareja',
  with_children: 'Con mis hijos/as',
  with_partner_and_children: 'Con mi pareja e hijos/as',
  with_family: 'Con otros familiares',
  with_caregiver: 'Con una persona cuidadora',
  other: 'Otra situación',
}

export const orgMemberRoleLabels: Record<OrgMemberRole, string> = {
  owner: 'Propietario',
  admin: 'Administrador',
  editor: 'Editor',
  viewer: 'Solo lectura',
}

export const organizationTypeLabels: Record<OrganizationType, string> = {
  health: 'Salud',
  community: 'Comunidad',
  ngo: 'Fundación / ONG',
  employer: 'Empleador',
  government: 'Municipalidad / Gobierno',
  educational: 'Educacional',
  other: 'Otra',
}

export const serviceTypeLabels: Record<ServiceType, string> = {
  establishment: 'Establecimiento',
  resource: 'Recurso de apoyo',
}

export const serviceStatusLabels: Record<ServiceStatus, string> = {
  active: 'Activo',
  inactive: 'Inactivo',
  draft: 'Borrador / en revisión',
}

export const activityTypeLabels: Record<ActivityType, string> = {
  workshop: 'Taller',
  event: 'Evento',
  course: 'Curso',
  support_group: 'Grupo de apoyo',
}

export const modalityLabels: Record<Modality, string> = {
  in_person: 'Presencial',
  online: 'En línea',
  hybrid: 'Híbrida',
}

export const activityStatusLabels: Record<ActivityStatus, string> = {
  draft: 'Borrador / en revisión',
  published: 'Publicada',
  cancelled: 'Cancelada',
  finished: 'Finalizada',
}

export const enrollmentStatusLabels: Record<EnrollmentStatus, string> = {
  enrolled: 'Inscrito',
  waitlist: 'Lista de espera',
  cancelled: 'Cancelada',
  attended: 'Asistió',
}

export const opportunityTypeLabels: Record<OpportunityType, string> = {
  job: 'Empleo',
  gig: 'Trabajo puntual',
  volunteer: 'Voluntariado',
  economic: 'Oportunidad económica',
}

export const opportunityStatusLabels: Record<OpportunityStatus, string> = {
  draft: 'Borrador',
  open: 'Abierta',
  closed: 'Cerrada',
}

export const assistanceStatusLabels: Record<AssistanceStatus, string> = {
  pending: 'Pendiente',
  notified: 'Notificada',
  resolved: 'Resuelta',
  cancelled: 'Cancelada',
}

export const reportTypeLabels: Record<ReportType, string> = {
  services: 'Servicios',
  activities: 'Actividades',
  enrollments: 'Inscripciones',
  assistance: 'Solicitudes de asistencia',
  opportunities: 'Oportunidades',
  territorial: 'Distribución territorial',
}

export const dayOfWeekLabels = [
  'Domingo',
  'Lunes',
  'Martes',
  'Miércoles',
  'Jueves',
  'Viernes',
  'Sábado',
] as const

/** Etiquetas legibles para los tipos de actividad que entrega la API. */
export const activityTypeFromValue: Record<string, string> = activityTypeLabels
