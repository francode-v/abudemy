/** Utilidades geográficas para el mapa territorial. */

export interface LatLng {
  latitude: number
  longitude: number
}

export interface Point extends LatLng {
  id: string
  label: string
  comuna?: string
  region?: string
  category?: string
  status?: string
  detail?: Record<string, unknown>
}

/** Distancia en kilómetros entre dos coordenadas (fórmula del haversine). */
export function distanceKm(a: LatLng, b: LatLng): number {
  const R = 6371
  const dLat = toRad(b.latitude - a.latitude)
  const dLng = toRad(b.longitude - a.longitude)
  const lat1 = toRad(a.latitude)
  const lat2 = toRad(b.latitude)
  const h =
    Math.sin(dLat / 2) ** 2 + Math.cos(lat1) * Math.cos(lat2) * Math.sin(dLng / 2) ** 2
  return 2 * R * Math.asin(Math.sqrt(h))
}

function toRad(deg: number): number {
  return (deg * Math.PI) / 180
}

export function hasCoordinates<T extends { latitude?: number | null; longitude?: number | null }>(
  item: T,
): item is T & { latitude: number; longitude: number } {
  return (
    typeof item.latitude === 'number' &&
    typeof item.longitude === 'number' &&
    Number.isFinite(item.latitude) &&
    Number.isFinite(item.longitude) &&
    item.latitude >= -90 &&
    item.latitude <= 90 &&
    item.longitude >= -180 &&
    item.longitude <= 180
  )
}

export interface Bounds {
  minLat: number
  maxLat: number
  minLng: number
  maxLng: number
}

export function computeBounds(points: LatLng[]): Bounds | null {
  if (points.length === 0) return null
  return points.reduce<Bounds>(
    (acc, p) => ({
      minLat: Math.min(acc.minLat, p.latitude),
      maxLat: Math.max(acc.maxLat, p.latitude),
      minLng: Math.min(acc.minLng, p.longitude),
      maxLng: Math.max(acc.maxLng, p.longitude),
    }),
    {
      minLat: points[0].latitude,
      maxLat: points[0].latitude,
      minLng: points[0].longitude,
      maxLng: points[0].longitude,
    },
  )
}

/** Centro de Chile continental: punto de partida razonable del mapa. */
export const CHILE_CENTER: [number, number] = [-33.45, -70.66]

export function formatLatLng(lat: number, lng: number): string {
  return `${lat.toFixed(5)}, ${lng.toFixed(5)}`
}
