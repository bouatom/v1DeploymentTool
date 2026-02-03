interface BarChartDataPoint {
  label: string
  value: number
}

interface BarChartProps {
  data: BarChartDataPoint[]
  color: string
}

export function BarChart({ data, color }: BarChartProps) {
  const maxValue = Math.max(...data.map((point) => point.value), 1)

  return (
    <div className="space-y-3 rounded-xl border border-slate-800 bg-slate-950/60 p-4">
      {data.map((point) => {
        const width = `${(point.value / maxValue) * 100}%`
        return (
          <div key={point.label} className="space-y-1">
            <div className="flex items-center justify-between text-xs text-slate-400">
              <span>{point.label}</span>
              <span className="text-slate-200">{point.value}</span>
            </div>
            <div className="h-2 w-full rounded-full bg-slate-800">
              <div className="h-2 rounded-full" style={{ width, background: color }} />
            </div>
          </div>
        )
      })}
    </div>
  )
}
