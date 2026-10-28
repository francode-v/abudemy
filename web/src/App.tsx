import { lazy } from 'react'
import { Route, Routes } from 'react-router-dom'
import { PublicLayout } from '@/components/layout/PublicLayout'
import { AppLayout } from '@/components/layout/AppLayout'
import { RequireAdmin, RequireAuth, RequireInstitutional, RequirePeopleViewer } from '@/app/guards'
import { HomePage } from '@/pages/public/HomePage'
import { FeaturesPage } from '@/pages/public/FeaturesPage'
import { ContactPage } from '@/pages/public/ContactPage'
import { LoginPage } from '@/pages/auth/LoginPage'
import { RegisterPage } from '@/pages/auth/RegisterPage'
import { ForgotPasswordPage, ResetPasswordPage } from '@/pages/auth/PasswordPages'
import { DashboardPage } from '@/pages/app/DashboardPage'
import { ServicesPage } from '@/pages/app/ServicesPage'
import { ActivitiesPage } from '@/pages/app/ActivitiesPage'
import { ActivityDetailPage } from '@/pages/app/ActivityDetailPage'
import { AssistancePage } from '@/pages/app/AssistancePage'
import { OpportunitiesPage } from '@/pages/app/OpportunitiesPage'
import { OrganizationsPage } from '@/pages/app/OrganizationsPage'
import { OrganizationDetailPage } from '@/pages/app/OrganizationDetailPage'
import { ProfilePage } from '@/pages/app/ProfilePage'
import { SessionsPage } from '@/pages/app/SessionsPage'
import { UsersAdminPage } from '@/pages/app/UsersAdminPage'
import { RegisteredPeoplePage } from '@/pages/app/RegisteredPeoplePage'
import { InstitutionalAccountsPage } from '@/pages/app/InstitutionalAccountsPage'
import { NotFoundPage } from '@/pages/app/NotFoundPage'

// Las vistas más pesadas (Leaflet y jsPDF) se cargan bajo demanda para no
// penalizar la carga inicial del portal.
const MapPage = lazy(() => import('@/pages/app/MapPage').then((module) => ({ default: module.MapPage })))
const ReportsPage = lazy(() =>
  import('@/pages/app/ReportsPage').then((module) => ({ default: module.ReportsPage })),
)

/**
 * Tabla de rutas de la aplicación.
 *
 * - Sitio público: páginas informativas sin sesión.
 * - Autenticación: inicio de sesión, registro y recuperación.
 * - Portal institucional (`/app`): requiere sesión; las secciones de gestión
 *   requieren además un rol institucional y la administración de usuarios,
 *   rol de administrador. Los permisos finos por organización los aplica el
 *   backend y la interfaz solo oculta acciones según el rol del usuario.
 */
export function App() {
  return (
    <Routes>
      <Route element={<PublicLayout />}>
        <Route path="/" element={<HomePage />} />
        <Route path="/funcionalidades" element={<FeaturesPage />} />
        <Route path="/contacto" element={<ContactPage />} />
        <Route path="*" element={<NotFoundPage />} />
      </Route>

      <Route path="/login" element={<LoginPage />} />
      <Route path="/registro" element={<RegisterPage />} />
      <Route path="/recuperar" element={<ForgotPasswordPage />} />
      <Route path="/restablecer" element={<ResetPasswordPage />} />

      <Route element={<RequireAuth />}>
        <Route path="/app" element={<AppLayout />}>
          <Route element={<RequireInstitutional />}>
            <Route index element={<DashboardPage />} />
            <Route path="mapa" element={<MapPage />} />
            <Route path="servicios" element={<ServicesPage />} />
            <Route path="actividades" element={<ActivitiesPage />} />
            <Route path="actividades/:id" element={<ActivityDetailPage />} />
            <Route path="asistencia" element={<AssistancePage />} />
            <Route path="oportunidades" element={<OpportunitiesPage />} />
            <Route path="informes" element={<ReportsPage />} />
            <Route path="organizaciones" element={<OrganizationsPage />} />
            <Route path="organizaciones/:id" element={<OrganizationDetailPage />} />
            <Route element={<RequirePeopleViewer />}>
              <Route path="personas" element={<RegisteredPeoplePage />} />
            </Route>
          </Route>

          <Route path="cuenta" element={<ProfilePage />} />
          <Route path="cuenta/sesiones" element={<SessionsPage />} />

          <Route element={<RequireAdmin />}>
            <Route path="admin/usuarios" element={<UsersAdminPage />} />
            <Route path="admin/institucionales" element={<InstitutionalAccountsPage />} />
          </Route>

          <Route path="*" element={<NotFoundPage />} />
        </Route>
      </Route>

      <Route path="*" element={<NotFoundPage />} />
    </Routes>
  )
}
