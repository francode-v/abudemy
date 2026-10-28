import L from 'leaflet'

/**
 * `leaflet.heat` es un plugin heredado que espera encontrar Leaflet en la
 * variable global `L` (hace `L.HeatLayer = ...` al evaluarse). Con el
 * empaquetado ESM de Vite esa global no existe, así que la exponemos aquí.
 *
 * Este módulo debe evaluarse ANTES de importar `leaflet.heat`: el orden de las
 * sentencias `import` en ESM lo garantiza, por eso se importa primero en
 * `TerritorialMap.tsx`.
 */
const scope = globalThis as typeof globalThis & { L?: typeof L }
if (!scope.L) {
  scope.L = L
}
