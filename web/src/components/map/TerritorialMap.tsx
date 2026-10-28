import 'leaflet/dist/leaflet.css'
import '@/lib/leafletSetup'
import 'leaflet.heat'
import { useEffect, useMemo } from 'react'
import L from 'leaflet'
import { CircleMarker, MapContainer, Popup, TileLayer, Tooltip, useMap } from 'react-leaflet'
import type { LatLng } from '@/lib/geo'
import { CHILE_CENTER, computeBounds } from '@/lib/geo'

export interface MapPoint extends LatLng {
  id: string
  label: string
  kind: 'service' | 'activity' | 'organization'
  comuna?: string
  region?: string
  status?: string
  detail?: { label: string; value: string }[]
}

const KIND_COLORS: Record<MapPoint['kind'], string> = {
  service: '#145C6B',
  activity: '#E9A23B',
  organization: '#6E5AA8',
}

interface HeatLayerProps {
  points: LatLng[]
  enabled: boolean
}

interface HeatFactory {
  heatLayer(
    latlngs: [number, number, number][],
    options?: Record<string, unknown>,
  ): L.Layer
}

function HeatLayer({ points, enabled }: HeatLayerProps) {
  const map = useMap()

  useEffect(() => {
    if (!enabled || points.length === 0) return
    const factory = L as unknown as HeatFactory
    if (typeof factory.heatLayer !== 'function') return
    const layer = factory.heatLayer(
      points.map((p) => [p.latitude, p.longitude, 0.7] as [number, number, number]),
      { radius: 28, blur: 22, maxZoom: 14, minOpacity: 0.35 },
    )
    layer.addTo(map)
    return () => {
      map.removeLayer(layer)
    }
  }, [enabled, map, points])

  return null
}

function FitBounds({ points }: { points: LatLng[] }) {
  const map = useMap()
  const signature = useMemo(() => points.map((p) => `${p.latitude},${p.longitude}`).join('|'), [points])

  useEffect(() => {
    if (points.length === 0) return
    const bounds = computeBounds(points)
    if (!bounds) return
    if (points.length === 1) {
      map.setView([points[0].latitude, points[0].longitude], 13)
      return
    }
    map.fitBounds(
      [
        [bounds.minLat, bounds.minLng],
        [bounds.maxLat, bounds.maxLng],
      ],
      { padding: [36, 36], maxZoom: 14 },
    )
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [signature, map])

  return null
}

export function TerritorialMap({
  points,
  heat = false,
  height = 460,
  onSelect,
}: {
  points: MapPoint[]
  heat?: boolean
  height?: number
  onSelect?: (point: MapPoint) => void
}) {
  return (
    <div className="map" style={{ height }}>
      <MapContainer
        center={CHILE_CENTER}
        zoom={5}
        scrollWheelZoom
        style={{ height: '100%', width: '100%' }}
        preferCanvas
      >
        <TileLayer
          attribution='&copy; colaboradores de <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>'
          url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
        />
        <HeatLayer points={points} enabled={heat} />
        <FitBounds points={points} />
        {points.map((point) => (
          <CircleMarker
            key={point.id}
            center={[point.latitude, point.longitude]}
            radius={7}
            pathOptions={{
              color: '#ffffff',
              weight: 2,
              fillColor: KIND_COLORS[point.kind],
              fillOpacity: 0.95,
            }}
            eventHandlers={{
              click: () => onSelect?.(point),
            }}
          >
            <Tooltip direction="top" offset={[0, -6]}>
              {point.label}
            </Tooltip>
            <Popup>
              <div className="map-popup">
                <strong>{point.label}</strong>
                <p className="map-popup__meta">
                  {[point.comuna, point.region].filter(Boolean).join(' · ') || 'Sin ubicación declarada'}
                </p>
                {point.detail && point.detail.length > 0 && (
                  <dl className="map-popup__detail">
                    {point.detail.map((item) => (
                      <div key={item.label}>
                        <dt>{item.label}</dt>
                        <dd>{item.value}</dd>
                      </div>
                    ))}
                  </dl>
                )}
              </div>
            </Popup>
          </CircleMarker>
        ))}
      </MapContainer>
    </div>
  )
}

export function MapLegend() {
  return (
    <ul className="map-legend">
      <li>
        <span className="map-legend__dot" style={{ background: KIND_COLORS.service }} /> Servicios
      </li>
      <li>
        <span className="map-legend__dot" style={{ background: KIND_COLORS.activity }} /> Actividades
      </li>
      <li>
        <span className="map-legend__dot" style={{ background: KIND_COLORS.organization }} /> Organizaciones
      </li>
    </ul>
  )
}
