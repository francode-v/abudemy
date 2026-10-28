import { useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { useCategories, useOverview } from '@/hooks/queries'
import { useAllActivities, useAllServices } from '@/hooks/paged'
import { useOrgScope } from '@/app/orgScope'
import { MapLegend, TerritorialMap, type MapPoint } from '@/components/map/TerritorialMap'
import { Card, CardBody, CardHeader, Callout, ErrorState, PageHeader, SelectInput, Spinner, TextInput, Button } from '@/components/ui/primitives'
import { SegmentedControl, KeyValue } from '@/components/ui/data'
import { Icon } from '@/components/ui/Icon'
import { humanizeError } from '@/api'
import { hasCoordinates, distanceKm } from '@/lib/geo'
import { formatNumber } from '@/lib/format'
import { activityStatusLabels, serviceStatusLabels, serviceTypeLabels } from '@/lib/dictionaries'

type Layer = 'services' | 'activities' | 'both'

export function MapPage() {
  const { organizationId, activeName } = useOrgScope()
  const categories = useCategories()
  const overview = useOverview(organizationId)

  const [layer, setLayer] = useState<Layer>('both')
  const [heat, setHeat] = useState(false)
  const [comuna, setComuna] = useState('')
  const [categoryId, setCategoryId] = useState('')
  const [from, setFrom] = useState('')
  const [to, setTo] = useState('')
  const [selected, setSelected] = useState<MapPoint | null>(null)

  const services = useAllServices({
    organization_id: organizationId ?? undefined,
    comuna: comuna || undefined,
    category_id: categoryId || undefined,
  })
  const activities = useAllActivities({
    organization_id: organizationId ?? undefined,
    comuna: comuna || undefined,
    from: from ? new Date(`${from}T00:00:00`).toISOString() : undefined,
    to: to ? new Date(`${to}T23:59:59`).toISOString() : undefined,
  })

  const comunaOptions = useMemo(() => {
    const set = new Set<string>()
    for (const item of overview.data?.territorial_distribution ?? []) if (item.comuna) set.add(item.comuna)
    for (const item of services.items) if (item.comuna) set.add(item.comuna)
    for (const item of activities.items) if (item.comuna) set.add(item.comuna)
    return [...set].sort((a, b) => a.localeCompare(b, 'es'))
  }, [overview.data, services.items, activities.items])

  const servicePoints: MapPoint[] = useMemo(
    () =>
      services.items.filter(hasCoordinates).map((service) => ({
        id: `service-${service.id}`,
        kind: 'service' as const,
        label: service.name,
        latitude: service.latitude,
        longitude: service.longitude,
        comuna: service.comuna,
        region: service.region,
        status: service.status,
        detail: [
          { label: 'Organización', value: service.organization_name || 'Sin organización' },
          { label: 'Estado', value: serviceStatusLabels[service.status] },
          { label: 'Tipo', value: serviceTypeLabels[service.type] },
          { label: 'Categoría', value: service.category?.name ?? 'Sin categoría' },
        ],
      })),
    [services.items],
  )

  const activityPoints: MapPoint[] = useMemo(
    () =>
      activities.items.filter(hasCoordinates).map((activity) => ({
        id: `activity-${activity.id}`,
        kind: 'activity' as const,
        label: activity.title,
        latitude: activity.latitude,
        longitude: activity.longitude,
        comuna: activity.comuna,
        region: activity.region,
        status: activity.status,
        detail: [
          { label: 'Estado', value: activityStatusLabels[activity.status] },
          { label: 'Inscritos', value: `${activity.enrolled_count}${
            activity.capacity > 0 ? ` / ${activity.capacity}` : ''
          }` },
        ],
      })),
    [activities.items],
  )

  const points = useMemo(() => {
    if (layer === 'services') return servicePoints
    if (layer === 'activities') return activityPoints
    return [...servicePoints, ...activityPoints]
  }, [layer, servicePoints, activityPoints])

  const zoneIndex = useMemo(() => {
    const map = new Map<string, { services: number; activities: number; upcoming: number; categories: Map<string, number> }>()
    for (const item of overview.data?.territorial_distribution ?? []) {
      const entry = map.get(item.comuna) ?? { services: 0, activities: 0, upcoming: 0, categories: new Map() }
      entry.services = Math.max(entry.services, item.count)
      map.set(item.comuna, entry)
    }
    for (const service of services.items) {
      if (!service.comuna) continue
      const entry = map.get(service.comuna) ?? { services: 0, activities: 0, upcoming: 0, categories: new Map() }
      if (!overview.data) entry.services += 1
      const label = service.category?.name ?? 'Sin categoría'
      entry.categories.set(label, (entry.categories.get(label) ?? 0) + 1)
      map.set(service.comuna, entry)
    }
    for (const activity of activities.items) {
      if (!activity.comuna) continue
      const entry = map.get(activity.comuna) ?? { services: 0, activities: 0, upcoming: 0, categories: new Map() }
      entry.activities += 1
      if (activity.status === 'published' && new Date(activity.start_at) >= new Date()) entry.upcoming += 1
      map.set(activity.comuna, entry)
    }
    return map
  }, [overview.data, services.items, activities.items])

  const selectedComuna = selected?.comuna ?? comuna
  const zone = selectedComuna ? zoneIndex.get(selectedComuna) : undefined

  const nearby = useMemo(() => {
    if (!selected) return []
    return services.items
      .filter(hasCoordinates)
      .filter((service) => distanceKm(selected, service) <= 2)
      .slice(0, 6)
  }, [selected, services.items])

  const failed = services.isError || activities.isError
  const loading = services.isLoading || activities.isLoading

  return (
    <>
      <PageHeader
        title="Mapa territorial"
        description={`Ubicación de servicios y actividades de ${activeName}. Solo se muestran coordenadas de establecimientos y sedes, nunca domicilios particulares.`}
        actions={
          <Button variant="secondary" icon="refresh" onClick={() => { services.refetch(); activities.refetch() }} loading={loading}>
            Actualizar
          </Button>
        }
      />

      {(services.truncated || activities.truncated) && (
        <Callout tone="warning" title="Mostrando una muestra">
          Hay más de {formatNumber(Math.max(services.loaded, activities.loaded))} registros. Para
          mantener el mapa ágil se cargan hasta 5 páginas por capa. Aplica filtros de comuna o
          categoría para acotar el resultado.
        </Callout>
      )}

      <Card>
        <CardBody>
          <div className="toolbar">
            <SelectInput
              label="Comuna"
              value={comuna}
              onChange={(event) => setComuna(event.target.value)}
            >
              <option value="">Todas las comunas</option>
              {comunaOptions.map((option) => (
                <option key={option} value={option}>
                  {option}
                </option>
              ))}
            </SelectInput>

            <SelectInput
              label="Categoría de servicio"
              value={categoryId}
              onChange={(event) => setCategoryId(event.target.value)}
            >
              <option value="">Todas las categorías</option>
              {(categories.data ?? []).map((category) => (
                <option key={category.id} value={category.id}>
                  {category.name}
                </option>
              ))}
            </SelectInput>

            <TextInput
              label="Actividades desde"
              type="date"
              value={from}
              onChange={(event) => setFrom(event.target.value)}
            />
            <TextInput
              label="Actividades hasta"
              type="date"
              value={to}
              onChange={(event) => setTo(event.target.value)}
            />

            <div className="toolbar__actions">
              <Button
                variant="ghost"
                size="sm"
                icon="close"
                onClick={() => {
                  setComuna('')
                  setCategoryId('')
                  setFrom('')
                  setTo('')
                }}
              >
                Limpiar
              </Button>
            </div>
          </div>

          <div className="row row--between" style={{ marginTop: 16 }}>
            <SegmentedControl
              label="Capa del mapa"
              value={layer}
              onChange={setLayer}
              options={[
                { value: 'both', label: 'Todo' },
                { value: 'services', label: 'Servicios' },
                { value: 'activities', label: 'Actividades' },
              ]}
            />
            <div className="row">
              <label className="checkbox">
                <input type="checkbox" checked={heat} onChange={(event) => setHeat(event.target.checked)} />
                <span>Mapa de calor</span>
              </label>
              <MapLegend />
            </div>
          </div>
        </CardBody>
      </Card>

      {failed ? (
        <ErrorState
          description={humanizeError(services.error ?? activities.error)}
          onRetry={() => {
            services.refetch()
            activities.refetch()
          }}
        />
      ) : (
        <div className="grid-2">
          <Card>
            <CardBody>
              {loading && points.length === 0 ? (
                <div className="full-loader" style={{ minHeight: 360 }}>
                  <Spinner size="lg" />
                  <p>Cargando puntos del mapa…</p>
                </div>
              ) : (
                <TerritorialMap points={points} heat={heat} onSelect={setSelected} height={520} />
              )}
              <p className="split-note" style={{ marginTop: 10 }}>
                {formatNumber(points.length)} puntos georreferenciados en pantalla
                {heat && ' · la capa de calor pondera la densidad de registros'}.
              </p>
            </CardBody>
          </Card>

          <div className="stack">
            <Card>
              <CardHeader
                title={selectedComuna ? `Zona: ${selectedComuna}` : 'Resumen de zona'}
                description={
                  selectedComuna
                    ? 'Indicadores agregados para la comuna seleccionada.'
                    : 'Selecciona una comuna o un punto del mapa para ver su resumen.'
                }
                icon="layers"
              />
              <CardBody>
                {!selectedComuna ? (
                  <p className="muted">
                    Usa el filtro de comuna o haz clic en un punto para ver cuántos servicios y
                    actividades hay en esa zona.
                  </p>
                ) : (
                  <>
                    <KeyValue
                      items={[
                        {
                          label: 'Servicios en el catálogo',
                          value: zone ? formatNumber(zone.services) : '—',
                        },
                        {
                          label: 'Actividades cargadas',
                          value: zone ? formatNumber(zone.activities) : '0',
                        },
                        {
                          label: 'Actividades próximas',
                          value: zone ? formatNumber(zone.upcoming) : '0',
                        },
                      ]}
                    />
                    {zone && zone.categories.size > 0 && (
                      <div style={{ marginTop: 14 }}>
                        <p className="form-section__title">Servicios por categoría en la zona</p>
                        <ul className="count-list">
                          {[...zone.categories.entries()].map(([label, count]) => (
                            <li key={label} className="count-list__item">
                              <span className="count-list__label">{label}</span>
                              <span className="count-list__bar">
                                <span
                                  style={{
                                    width: `${Math.min(
                                      100,
                                      (count / Math.max(...zone.categories.values())) * 100,
                                    )}%`,
                                  }}
                                />
                              </span>
                              <span className="count-list__value">{count}</span>
                            </li>
                          ))}
                        </ul>
                      </div>
                    )}
                    <Link
                      className="btn btn--secondary btn--sm"
                      style={{ marginTop: 14 }}
                      to={`/app/servicios?comuna=${encodeURIComponent(selectedComuna)}`}
                    >
                      <Icon name="list" size={14} /> Ver servicios de {selectedComuna}
                    </Link>
                  </>
                )}
              </CardBody>
            </Card>

            {selected && (
              <Card>
                <CardHeader title="Punto seleccionado" icon="pin" />
                <CardBody>
                  <p style={{ fontWeight: 600 }}>{selected.label}</p>
                  <p className="muted text-sm">
                    {[selected.comuna, selected.region].filter(Boolean).join(' · ') || 'Sin ubicación declarada'}
                  </p>
                  {selected.detail && selected.detail.length > 0 && (
                    <KeyValue items={selected.detail} />
                  )}
                  {nearby.length > 0 && (
                    <div style={{ marginTop: 12 }}>
                      <p className="form-section__title">Servicios a menos de 2 km</p>
                      <ul className="stack" style={{ gap: 6 }}>
                        {nearby.map((service) => (
                          <li key={service.id} className="text-sm">
                            <Icon name="pin" size={13} /> {service.name}
                          </li>
                        ))}
                      </ul>
                    </div>
                  )}
                </CardBody>
              </Card>
            )}

            <Card>
              <CardHeader
                title="Cobertura por comuna"
                description="Servicios registrados en el ámbito seleccionado."
              />
              <CardBody>
                {(overview.data?.territorial_distribution ?? []).length === 0 ? (
                  <p className="muted">Sin información territorial disponible.</p>
                ) : (
                  <ul className="count-list">
                    {[...(overview.data?.territorial_distribution ?? [])]
                      .sort((a, b) => b.count - a.count)
                      .slice(0, 12)
                      .map((item) => (
                        <li key={`${item.comuna}-${item.region}`} className="count-list__item">
                          <button
                            className="count-list__label"
                            style={{ border: 'none', background: 'none', cursor: 'pointer', textAlign: 'left', padding: 0 }}
                            onClick={() => setComuna(item.comuna)}
                          >
                            {item.comuna}
                          </button>
                          <span className="count-list__bar">
                            <span
                              style={{
                                width: `${
                                  (item.count /
                                    Math.max(
                                      ...(overview.data?.territorial_distribution ?? [{ count: 1 }]).map(
                                        (entry) => entry.count,
                                      ),
                                    )) *
                                  100
                                }%`,
                              }}
                            />
                          </span>
                          <span className="count-list__value">{item.count}</span>
                        </li>
                      ))}
                  </ul>
                )}
              </CardBody>
            </Card>
          </div>
        </div>
      )}

      <Callout tone="info" title="Privacidad de los datos">
        El mapa excluye deliberadamente las solicitudes de asistencia y cualquier dato personal:
        las coordenadas de una solicitud solo son visibles para su autor o para administradores y
        nunca se agregan aquí.
      </Callout>
    </>
  )
}
