import { useEffect, useMemo, useState } from 'react'
import {
  api,
  humanizeError,
  type HealthSystem,
  type LivingSituation,
  type Sex,
  type User,
} from '@/api'
import type { UserQuery } from '@/api/endpoints'
import { useUsers } from '@/hooks/queries'
import { useDebouncedValue } from '@/hooks/useDebouncedValue'
import { useToast } from '@/hooks/useToast'
import {
  Button,
  Callout,
  Card,
  CardBody,
  EmptyState,
  ErrorState,
  PageHeader,
  SelectInput,
  TextInput,
} from '@/components/ui/primitives'
import { DataTable, Pagination, StatCard, type Column } from '@/components/ui/data'
import {
  healthSystemLabels,
  livingSituationLabels,
  sexLabels,
  userStatusLabels,
} from '@/lib/dictionaries'
import { ageFrom, formatDate, formatNumber } from '@/lib/format'
import { ageDistribution, distribution } from '@/lib/people'
import { buildCsv, downloadCsv, timestampSuffix, type CsvColumn } from '@/lib/csv'
import { buildPdf, downloadPdf } from '@/lib/pdf'

const PER_PAGE = 20
const MAX_AGE = 120

const HEALTH_SYSTEMS = Object.keys(healthSystemLabels) as HealthSystem[]
const SEXES = Object.keys(sexLabels) as Sex[]
const LIVING_SITUATIONS = Object.keys(livingSituationLabels) as LivingSituation[]

/** Edad mínima para el indicador de personas mayores. */
const SENIOR_AGE = 60

/** La API permite hasta 100 filas por página; la exportación recorre todas. */
const EXPORT_PER_PAGE = 100
/** Tope de páginas para no bloquear el navegador con exportaciones enormes. */
const MAX_EXPORT_PAGES = 50
/** Máximo de filas que se listan en el PDF (el CSV no tiene tope práctico). */
const MAX_PDF_ROWS = 500

function toAge(value: string): number | undefined {
  if (value.trim() === '') return undefined
  const parsed = Number(value)
  if (!Number.isFinite(parsed) || parsed < 0 || parsed > MAX_AGE) return undefined
  return Math.trunc(parsed)
}

interface CollectedPeople {
  items: User[]
  /** true si la API tenía más páginas que el tope de exportación. */
  truncated: boolean
}

/** Recorre todas las páginas de la API para exportar el listado completo filtrado. */
async function collectPeople(filters: UserQuery): Promise<CollectedPeople> {
  const items: User[] = []
  let page = 1
  let totalPages = 1
  let truncated = false
  do {
    const result = await api.users.list({ ...filters, page, per_page: EXPORT_PER_PAGE })
    items.push(...result.items)
    totalPages = result.pagination.total_pages
    if (page >= MAX_EXPORT_PAGES) {
      truncated = totalPages > page
      break
    }
    page += 1
  } while (page <= totalPages && totalPages > 0)
  return { items, truncated }
}

/**
 * Vista de personas registradas: cuentas de la comunidad (`mobile_user`) con los
 * datos sociodemográficos que declaran al crear su cuenta en la app móvil.
 * Permite segmentar por comuna, edad, sexo, previsión y convivencia, y exportar
 * el informe (PDF/CSV) del conjunto filtrado completo, no solo de la página.
 */
export function RegisteredPeoplePage() {
  const toast = useToast()

  const [search, setSearch] = useState('')
  const [comuna, setComuna] = useState('')
  const [ageMin, setAgeMin] = useState('')
  const [ageMax, setAgeMax] = useState('')
  const [sex, setSex] = useState('')
  const [healthSystem, setHealthSystem] = useState('')
  const [livingSituation, setLivingSituation] = useState('')
  const [page, setPage] = useState(1)
  const [exporting, setExporting] = useState<'csv' | 'pdf' | null>(null)

  const debouncedSearch = useDebouncedValue(search)
  const debouncedComuna = useDebouncedValue(comuna)

  const parsedAgeMin = toAge(ageMin)
  const parsedAgeMax = toAge(ageMax)
  const invalidAge =
    (ageMin.trim() !== '' && parsedAgeMin === undefined) ||
    (ageMax.trim() !== '' && parsedAgeMax === undefined) ||
    (parsedAgeMin !== undefined && parsedAgeMax !== undefined && parsedAgeMin > parsedAgeMax)

  useEffect(() => {
    setPage(1)
  }, [
    debouncedSearch,
    debouncedComuna,
    parsedAgeMin,
    parsedAgeMax,
    sex,
    healthSystem,
    livingSituation,
  ])

  /** Filtros aplicados, sin paginación: base común para la tabla y la exportación. */
  const baseQuery = useMemo<UserQuery>(
    () => ({
      role: 'mobile_user',
      search: debouncedSearch.trim() || undefined,
      comuna: debouncedComuna.trim() || undefined,
      age_min: parsedAgeMin,
      age_max: parsedAgeMax,
      sex: sex || undefined,
      health_system: healthSystem || undefined,
      living_situation: livingSituation || undefined,
    }),
    [
      debouncedSearch,
      debouncedComuna,
      parsedAgeMin,
      parsedAgeMax,
      sex,
      healthSystem,
      livingSituation,
    ],
  )

  const query = useMemo<UserQuery>(
    () => ({ ...baseQuery, page, per_page: PER_PAGE }),
    [baseQuery, page],
  )

  const users = useUsers(query, !invalidAge)

  const items = users.data?.items ?? []
  const ages = items
    .map((item) => ageFrom(item.birth_date))
    .filter((value): value is number => value !== null)
  const withComuna = items.filter((item) => Boolean(item.comuna)).length
  const seniors = ages.filter((value) => value >= SENIOR_AGE).length
  const averageAge = ages.length
    ? Math.round(ages.reduce((total, value) => total + value, 0) / ages.length)
    : null

  /** Resumen legible de los filtros, usado en el encabezado del PDF. */
  const filterSummary = useMemo(() => {
    const parts: string[] = []
    if (baseQuery.search) parts.push(`Búsqueda: «${baseQuery.search}»`)
    if (baseQuery.comuna) parts.push(`Comuna contiene «${baseQuery.comuna}»`)
    if (baseQuery.age_min !== undefined && baseQuery.age_max !== undefined) {
      parts.push(`Edad: ${baseQuery.age_min} a ${baseQuery.age_max} años`)
    } else if (baseQuery.age_min !== undefined) {
      parts.push(`Edad: desde ${baseQuery.age_min} años`)
    } else if (baseQuery.age_max !== undefined) {
      parts.push(`Edad: hasta ${baseQuery.age_max} años`)
    }
    if (baseQuery.sex) parts.push(`Sexo: ${sexLabels[baseQuery.sex as Sex]}`)
    if (baseQuery.health_system) {
      parts.push(`Previsión: ${healthSystemLabels[baseQuery.health_system as HealthSystem]}`)
    }
    if (baseQuery.living_situation) {
      parts.push(
        `Convivencia: ${livingSituationLabels[baseQuery.living_situation as LivingSituation]}`,
      )
    }
    return parts.length ? parts.join(' · ') : 'todas las personas registradas'
  }, [baseQuery])

  const columns: Column<User>[] = [
    {
      key: 'person',
      header: 'Persona',
      render: (row) => (
        <div>
          <strong>{row.full_name}</strong>
          <div className="muted text-sm">{row.email}</div>
        </div>
      ),
    },
    {
      key: 'age',
      header: 'Edad',
      width: '80px',
      render: (row) => {
        const age = ageFrom(row.birth_date)
        return age === null ? <span className="muted">—</span> : `${age} años`
      },
    },
    {
      key: 'comuna',
      header: 'Comuna',
      render: (row) => row.comuna || <span className="muted">Sin registrar</span>,
    },
    {
      key: 'health',
      header: 'Previsión de salud',
      render: (row) =>
        row.health_system ? (
          healthSystemLabels[row.health_system]
        ) : (
          <span className="muted">Sin registrar</span>
        ),
      hideOnMobile: true,
    },
    {
      key: 'sex',
      header: 'Sexo',
      render: (row) =>
        row.sex ? sexLabels[row.sex] : <span className="muted">Sin registrar</span>,
      hideOnMobile: true,
    },
    {
      key: 'living',
      header: 'Convivencia',
      render: (row) =>
        row.living_situation ? (
          livingSituationLabels[row.living_situation]
        ) : (
          <span className="muted">Sin registrar</span>
        ),
      hideOnMobile: true,
    },
    {
      key: 'created',
      header: 'Registro',
      render: (row) => <span className="text-sm muted">{formatDate(row.created_at)}</span>,
      hideOnMobile: true,
    },
  ]

  const csvColumns: CsvColumn<User>[] = [
    { header: 'Nombre completo', value: (row) => row.full_name },
    { header: 'Correo', value: (row) => row.email },
    { header: 'Fecha de nacimiento', value: (row) => row.birth_date ?? '' },
    { header: 'Edad', value: (row) => ageFrom(row.birth_date) ?? '' },
    { header: 'Comuna', value: (row) => row.comuna ?? '' },
    {
      header: 'Previsión de salud',
      value: (row) => (row.health_system ? healthSystemLabels[row.health_system] : ''),
    },
    { header: 'Sexo', value: (row) => (row.sex ? sexLabels[row.sex] : '') },
    {
      header: 'Situación de convivencia',
      value: (row) => (row.living_situation ? livingSituationLabels[row.living_situation] : ''),
    },
    { header: 'Estado de la cuenta', value: (row) => userStatusLabels[row.status] },
    { header: 'Fecha de registro', value: (row) => row.created_at.slice(0, 10) },
  ]

  const clearFilters = () => {
    setSearch('')
    setComuna('')
    setAgeMin('')
    setAgeMax('')
    setSex('')
    setHealthSystem('')
    setLivingSituation('')
  }

  const exportDisabled =
    exporting !== null || invalidAge || (users.data?.pagination.total ?? 0) === 0

  const handleCsv = async () => {
    setExporting('csv')
    try {
      const { items: rows, truncated } = await collectPeople(baseQuery)
      const csv = buildCsv(rows, csvColumns)
      downloadCsv(`abudemy-personas-registradas-${timestampSuffix()}.csv`, csv)
      toast.success(
        'CSV descargado',
        `${formatNumber(rows.length)} personas exportadas con los filtros aplicados.`,
      )
      if (truncated) {
        toast.warning(
          'Exportación parcial',
          `Se exportaron las primeras ${formatNumber(rows.length)} personas; acota los filtros para el resto.`,
        )
      }
    } catch (error) {
      toast.error('No se pudo exportar el CSV', humanizeError(error))
    } finally {
      setExporting(null)
    }
  }

  const handlePdf = async () => {
    setExporting('pdf')
    try {
      const { items: rows, truncated } = await collectPeople(baseQuery)
      const ages = rows
        .map((row) => ageFrom(row.birth_date))
        .filter((value): value is number => value !== null)
      const average = ages.length
        ? Math.round(ages.reduce((total, value) => total + value, 0) / ages.length)
        : null
      const listed = rows.slice(0, MAX_PDF_ROWS)

      const blob = buildPdf({
        title: 'Personas registradas — Informe sociodemográfico',
        subtitle: `Filtros aplicados: ${filterSummary}`,
        metadata: [
          { label: 'Generado', value: new Date().toLocaleString('es-CL') },
          { label: 'Registros', value: formatNumber(rows.length) },
          { label: 'Fuente de datos', value: 'API institucional de Abudemy' },
          {
            label: 'Listado',
            value: truncated
              ? `Primeras ${formatNumber(listed.length)} de ${formatNumber(rows.length)}`
              : `${formatNumber(listed.length)} personas`,
          },
        ],
        sections: [
          {
            title: 'Resumen del conjunto filtrado',
            columns: ['Indicador', 'Valor'],
            rows: [
              ['Personas registradas', formatNumber(rows.length)],
              [`Personas de ${SENIOR_AGE} años o más`, formatNumber(ages.filter((age) => age >= SENIOR_AGE).length)],
              ['Edad promedio', average === null ? 'Sin datos' : `${average} años`],
              ['Con comuna registrada', formatNumber(rows.filter((row) => row.comuna).length)],
              ['Con previsión de salud registrada', formatNumber(rows.filter((row) => row.health_system).length)],
              ['Con situación de convivencia registrada', formatNumber(rows.filter((row) => row.living_situation).length)],
              ['Fichas sociodemográficas completas', formatNumber(rows.filter((row) => row.birth_date && row.comuna && row.health_system && row.sex && row.living_situation).length)],
            ],
          },
          {
            title: 'Distribución por tramo de edad',
            columns: ['Tramo', 'Personas'],
            rows: ageDistribution(rows),
          },
          {
            title: 'Distribución por comuna',
            columns: ['Comuna', 'Personas'],
            rows: distribution(rows, (row) => row.comuna ?? ''),
          },
          {
            title: 'Previsión de salud',
            columns: ['Previsión', 'Personas'],
            rows: distribution(rows, (row) =>
              row.health_system ? healthSystemLabels[row.health_system] : '',
            ),
          },
          {
            title: 'Sexo',
            columns: ['Sexo', 'Personas'],
            rows: distribution(rows, (row) => (row.sex ? sexLabels[row.sex] : '')),
          },
          {
            title: 'Situación de convivencia',
            columns: ['Convivencia', 'Personas'],
            rows: distribution(rows, (row) =>
              row.living_situation ? livingSituationLabels[row.living_situation] : '',
            ),
          },
          {
            title: truncated
              ? `Listado de personas (primeras ${formatNumber(listed.length)})`
              : 'Listado de personas',
            columns: [
              'Nombre',
              'Correo',
              'Nacimiento',
              'Edad',
              'Comuna',
              'Previsión',
              'Sexo',
              'Convivencia',
            ],
            rows: listed.map((row) => {
              const age = ageFrom(row.birth_date)
              return [
                row.full_name,
                row.email,
                row.birth_date ?? '—',
                age === null ? '—' : age,
                row.comuna || 'Sin registrar',
                row.health_system ? healthSystemLabels[row.health_system] : 'Sin registrar',
                row.sex ? sexLabels[row.sex] : 'Sin registrar',
                row.living_situation
                  ? livingSituationLabels[row.living_situation]
                  : 'Sin registrar',
              ]
            }),
          },
        ],
        footer:
          'Datos declarados por las personas al registrarse en la app móvil de Abudemy; no corresponden a registros clínicos.',
      })
      downloadPdf(blob, `abudemy-personas-registradas-${timestampSuffix()}.pdf`)
      toast.success('Informe PDF generado', `${formatNumber(rows.length)} personas incluidas.`)
    } catch (error) {
      toast.error('No se pudo generar el PDF', humanizeError(error))
    } finally {
      setExporting(null)
    }
  }

  return (
    <>
      <PageHeader
        title="Personas registradas"
        description="Cuentas de la comunidad creadas desde la app móvil, con los datos que cada persona declara al registrarse."
        actions={
          <>
            <Button
              variant="secondary"
              icon="refresh"
              onClick={() => void users.refetch()}
              loading={users.isFetching}
            >
              Actualizar
            </Button>
            <Button
              variant="secondary"
              icon="download"
              onClick={() => void handleCsv()}
              loading={exporting === 'csv'}
              disabled={exportDisabled}
            >
              Exportar CSV
            </Button>
            <Button
              icon="download"
              onClick={() => void handlePdf()}
              loading={exporting === 'pdf'}
              disabled={exportDisabled}
            >
              Exportar PDF
            </Button>
          </>
        }
      />

      <Callout tone="info" title="Datos declarados, no clínicos">
        Comuna, edad, sexo, previsión y convivencia son datos que la persona declara al crear su
        cuenta. No provienen de registros clínicos ni de fuentes externas, y algunas cuentas
        antiguas pueden no tenerlos completos.
      </Callout>

      <section className="stat-grid">
        <StatCard
          label="Total de personas"
          value={formatNumber(users.data?.pagination.total)}
          icon="users"
          description="Personas registradas que coinciden con los filtros, en todas las páginas."
          updatedAt={users.dataUpdatedAt ? new Date(users.dataUpdatedAt) : undefined}
        />
        <StatCard
          label={`${SENIOR_AGE} años o más`}
          value={formatNumber(seniors)}
          tone="success"
          icon="heart"
          description={`Personas de ${SENIOR_AGE} años o más dentro de la página actual.`}
          updatedAt={users.dataUpdatedAt ? new Date(users.dataUpdatedAt) : undefined}
        />
        <StatCard
          label="Edad promedio"
          value={averageAge === null ? '—' : `${averageAge} años`}
          tone="primary"
          icon="trend"
          description="Promedio de edad de la página actual, considerando solo cuentas con fecha de nacimiento."
          updatedAt={users.dataUpdatedAt ? new Date(users.dataUpdatedAt) : undefined}
        />
        <StatCard
          label="Con comuna registrada"
          value={formatNumber(withComuna)}
          tone="warning"
          icon="pin"
          description="Personas de la página actual que indicaron su comuna."
          updatedAt={users.dataUpdatedAt ? new Date(users.dataUpdatedAt) : undefined}
        />
      </section>

      <Card>
        <div className="filters-bar">
          <div className="field toolbar__search" style={{ flex: '2 1 220px' }}>
            <label className="field__label" htmlFor="people-search">
              Buscar
            </label>
            <TextInput
              id="people-search"
              leadingIcon="search"
              placeholder="Nombre o correo…"
              value={search}
              onChange={(event) => setSearch(event.target.value)}
            />
          </div>

          <div className="field" style={{ flex: '1 1 160px' }}>
            <label className="field__label" htmlFor="people-comuna">
              Comuna
            </label>
            <TextInput
              id="people-comuna"
              leadingIcon="pin"
              placeholder="Ej: Concepción"
              value={comuna}
              onChange={(event) => setComuna(event.target.value)}
            />
          </div>

          <div className="field" style={{ flex: '0 1 110px' }}>
            <label className="field__label" htmlFor="people-age-min">
              Edad desde
            </label>
            <TextInput
              id="people-age-min"
              type="number"
              inputMode="numeric"
              min={0}
              max={MAX_AGE}
              placeholder="0"
              value={ageMin}
              onChange={(event) => setAgeMin(event.target.value)}
            />
          </div>

          <div className="field" style={{ flex: '0 1 110px' }}>
            <label className="field__label" htmlFor="people-age-max">
              Edad hasta
            </label>
            <TextInput
              id="people-age-max"
              type="number"
              inputMode="numeric"
              min={0}
              max={MAX_AGE}
              placeholder="120"
              value={ageMax}
              onChange={(event) => setAgeMax(event.target.value)}
            />
          </div>

          <SelectInput label="Sexo" value={sex} onChange={(event) => setSex(event.target.value)}>
            <option value="">Todos</option>
            {SEXES.map((value) => (
              <option key={value} value={value}>
                {sexLabels[value]}
              </option>
            ))}
          </SelectInput>

          <SelectInput
            label="Previsión"
            value={healthSystem}
            onChange={(event) => setHealthSystem(event.target.value)}
          >
            <option value="">Todas</option>
            {HEALTH_SYSTEMS.map((value) => (
              <option key={value} value={value}>
                {healthSystemLabels[value]}
              </option>
            ))}
          </SelectInput>

          <SelectInput
            label="Convivencia"
            value={livingSituation}
            onChange={(event) => setLivingSituation(event.target.value)}
          >
            <option value="">Todas</option>
            {LIVING_SITUATIONS.map((value) => (
              <option key={value} value={value}>
                {livingSituationLabels[value]}
              </option>
            ))}
          </SelectInput>

          <div className="toolbar__actions">
            <Button variant="ghost" size="sm" icon="close" onClick={clearFilters}>
              Limpiar filtros
            </Button>
          </div>
        </div>

        {invalidAge && (
          <p className="field__error" role="alert" style={{ margin: '0 1rem' }}>
            Revisa el rango de edad: debe estar entre 0 y {MAX_AGE} años, con «desde» menor o igual
            que «hasta».
          </p>
        )}

        <CardBody>
          <p className="text-sm muted" style={{ marginTop: 0 }}>
            <strong>Filtros aplicados:</strong> {filterSummary}. Las exportaciones incluyen{' '}
            <strong>todas las personas que coinciden con los filtros</strong>, no solo esta página.
          </p>

          {users.isError ? (
            <ErrorState description={humanizeError(users.error)} onRetry={() => void users.refetch()} />
          ) : (
            <>
              <DataTable
                columns={columns}
                rows={items}
                rowKey={(row) => row.id}
                loading={users.isLoading}
                empty={
                  <EmptyState
                    title="No encontramos personas"
                    description="Ajusta la búsqueda o los filtros de comuna, edad, sexo, previsión o convivencia."
                    icon="users"
                  />
                }
              />
              {users.data && users.data.pagination.total > PER_PAGE && (
                <Pagination
                  page={users.data.pagination.page}
                  perPage={users.data.pagination.per_page}
                  total={users.data.pagination.total}
                  totalPages={users.data.pagination.total_pages}
                  onPageChange={setPage}
                />
              )}
            </>
          )}
        </CardBody>
      </Card>
    </>
  )
}
