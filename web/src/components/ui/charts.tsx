import {
  Bar,
  BarChart,
  CartesianGrid,
  Cell,
  Legend,
  Line,
  LineChart,
  Pie,
  PieChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts'
import type { CountByLabel, TerritorialDistribution } from '@/api'

export const chartPalette = [
  '#145C6B',
  '#E9A23B',
  '#2E7D5B',
  '#3F6FB5',
  '#B4553F',
  '#6E5AA8',
  '#4E93A8',
  '#9C7A3C',
  '#7A8B94',
  '#C0619A',
]

const axisStyle = { fontSize: 12, fill: '#5A6B75' }

export function CategoryBarChart({ data, color = chartPalette[0] }: { data: CountByLabel[]; color?: string }) {
  return (
    <ResponsiveContainer width="100%" height={280}>
      <BarChart data={data} margin={{ top: 8, right: 16, bottom: 40, left: 0 }}>
        <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="#E3E9ED" />
        <XAxis
          dataKey="label"
          tick={axisStyle}
          interval={0}
          angle={-25}
          textAnchor="end"
          height={60}
          tickLine={false}
        />
        <YAxis allowDecimals={false} tick={axisStyle} width={38} tickLine={false} axisLine={false} />
        <Tooltip
          formatter={(value: number) => [value, 'Registros']}
          contentStyle={{ borderRadius: 10, border: '1px solid #E3E9ED', fontSize: 13 }}
        />
        <Bar dataKey="count" fill={color} radius={[6, 6, 0, 0]} maxBarSize={46} />
      </BarChart>
    </ResponsiveContainer>
  )
}

export function DistributionPieChart({ data }: { data: CountByLabel[] }) {
  return (
    <ResponsiveContainer width="100%" height={280}>
      <PieChart>
        <Pie data={data} dataKey="count" nameKey="label" innerRadius={54} outerRadius={92} paddingAngle={2}>
          {data.map((entry, index) => (
            <Cell key={entry.label} fill={chartPalette[index % chartPalette.length]} stroke="#fff" strokeWidth={2} />
          ))}
        </Pie>
        <Tooltip
          formatter={(value: number, name: string) => [value, name]}
          contentStyle={{ borderRadius: 10, border: '1px solid #E3E9ED', fontSize: 13 }}
        />
        <Legend
          verticalAlign="bottom"
          height={54}
          iconType="circle"
          wrapperStyle={{ fontSize: 12, color: '#42535C' }}
        />
      </PieChart>
    </ResponsiveContainer>
  )
}

export function TerritorialBarChart({ data }: { data: TerritorialDistribution[] }) {
  const sorted = [...data].sort((a, b) => b.count - a.count).slice(0, 10)
  return (
    <ResponsiveContainer width="100%" height={Math.max(240, sorted.length * 34 + 40)}>
      <BarChart data={sorted} layout="vertical" margin={{ top: 8, right: 24, bottom: 8, left: 8 }}>
        <CartesianGrid strokeDasharray="3 3" horizontal={false} stroke="#E3E9ED" />
        <XAxis type="number" allowDecimals={false} tick={axisStyle} tickLine={false} axisLine={false} />
        <YAxis
          type="category"
          dataKey="comuna"
          width={140}
          tick={axisStyle}
          tickLine={false}
          axisLine={false}
        />
        <Tooltip
          formatter={(value: number) => [value, 'Servicios']}
          contentStyle={{ borderRadius: 10, border: '1px solid #E3E9ED', fontSize: 13 }}
        />
        <Bar dataKey="count" fill={chartPalette[1]} radius={[0, 6, 6, 0]} maxBarSize={22} />
      </BarChart>
    </ResponsiveContainer>
  )
}

export interface TrendPoint {
  label: string
  value: number
}

/**
 * Serie temporal genérica. Solo se usa cuando la API entrega datos con fecha
 * (por ejemplo, inscripciones por mes calculadas desde `created_at`).
 */
export function TrendLineChart({
  data,
  label = 'Registros',
  color = chartPalette[2],
}: {
  data: TrendPoint[]
  label?: string
  color?: string
}) {
  return (
    <ResponsiveContainer width="100%" height={280}>
      <LineChart data={data} margin={{ top: 12, right: 16, bottom: 8, left: 0 }}>
        <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="#E3E9ED" />
        <XAxis dataKey="label" tick={axisStyle} tickLine={false} />
        <YAxis allowDecimals={false} tick={axisStyle} width={38} tickLine={false} axisLine={false} />
        <Tooltip
          formatter={(value: number) => [value, label]}
          contentStyle={{ borderRadius: 10, border: '1px solid #E3E9ED', fontSize: 13 }}
        />
        <Line
          type="monotone"
          dataKey="value"
          stroke={color}
          strokeWidth={2.5}
          dot={{ r: 3, fill: color }}
          activeDot={{ r: 5 }}
        />
      </LineChart>
    </ResponsiveContainer>
  )
}

export function HorizontalCountList({ data, unit = 'registros' }: { data: CountByLabel[]; unit?: string }) {
  const max = Math.max(1, ...data.map((d) => d.count))
  return (
    <ul className="count-list">
      {data.map((item) => (
        <li key={item.label} className="count-list__item">
          <span className="count-list__label">{item.label}</span>
          <span className="count-list__bar" aria-hidden="true">
            <span style={{ width: `${(item.count / max) * 100}%` }} />
          </span>
          <span className="count-list__value">
            {item.count} <small>{unit}</small>
          </span>
        </li>
      ))}
    </ul>
  )
}
