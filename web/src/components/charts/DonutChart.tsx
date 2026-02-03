interface DonutChartProps {
  value: number
  total: number
  label: string
  color: string
}

export function DonutChart({ value, total, label, color }: DonutChartProps) {
  const radius = 40
  const stroke = 10
  const normalizedValue = total > 0 ? Math.min(value / total, 1) : 0
  const circumference = 2 * Math.PI * radius
  const dash = normalizedValue * circumference

  return (
    <div className="flex items-center gap-4 rounded-xl border border-slate-800 bg-slate-950/60 p-4">
      <svg width="100" height="100" className="shrink-0">
        <circle
          cx="50"
          cy="50"
          r={radius}
          stroke="rgba(148, 163, 184, 0.2)"
          strokeWidth={stroke}
          fill="none"
        />
        <circle
          cx="50"
          cy="50"
          r={radius}
          stroke={color}
          strokeWidth={stroke}
          fill="none"
          strokeDasharray={`${dash} ${circumference - dash}`}
          strokeLinecap="round"
          transform="rotate(-90 50 50)"
        />
      </svg>
      <div>
        <p className="text-sm text-slate-400">{label}</p>
        <p className="text-2xl font-semibold text-white">
          {value} / {total}
        </p>
      </div>
    </div>
  )
}
