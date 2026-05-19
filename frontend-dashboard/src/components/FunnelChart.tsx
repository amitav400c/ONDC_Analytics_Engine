import { useFunnel } from '../api/client';
import { BarChart, Bar, XAxis, YAxis, Tooltip, ResponsiveContainer, Cell } from 'recharts';

const STAGE_LABELS: Record<string, string> = {
  on_search: 'Search',
  on_select: 'Select',
  on_init: 'Init',
  on_confirm: 'Confirm',
  on_cancel: 'Cancel',
};

const STAGE_ORDER = ['on_search', 'on_select', 'on_init', 'on_confirm', 'on_cancel'];
const COLORS = ['#34d399', '#2dd4bf', '#22d3ee', '#818cf8', '#f87171'];

export default function FunnelChart() {
  const { data, isLoading, isError } = useFunnel();

  if (isLoading) return <Skeleton />;
  if (isError) return <ErrorState />;

  const stages = data?.stages || [];
  const sorted = STAGE_ORDER
    .map(key => stages.find(s => s.event_type === key))
    .filter(Boolean)
    .map((s, i) => ({
      name: STAGE_LABELS[s!.event_type] || s!.event_type,
      count: s!.count,
      rate: s!.rate,
      color: COLORS[i % COLORS.length],
    }));

  if (sorted.length === 0) return <EmptyState />;

  return (
    <div>
      <ResponsiveContainer width="100%" height={280}>
        <BarChart data={sorted} layout="vertical" margin={{ left: 20 }}>
          <XAxis type="number" stroke="#ffffff20" tick={{ fill: '#ffffff60', fontSize: 12 }} />
          <YAxis type="category" dataKey="name" stroke="#ffffff20" tick={{ fill: '#ffffff80', fontSize: 13 }} width={70} />
          <Tooltip
            contentStyle={{ background: '#1e293b', border: '1px solid #ffffff15', borderRadius: '12px' }}
            labelStyle={{ color: '#ffffff90' }}
            itemStyle={{ color: '#34d399' }}
            formatter={(value: number, _: string, entry: unknown) => {
              const rate = (entry as { payload?: { rate?: number } })?.payload?.rate;
              return [`${value.toLocaleString()}${rate ? ` (${rate.toFixed(1)}%)` : ''}`, 'Events'];
            }}
          />
          <Bar dataKey="count" radius={[0, 8, 8, 0]} barSize={32}>
            {sorted.map((entry, i) => (
              <Cell key={i} fill={entry.color} fillOpacity={0.8} />
            ))}
          </Bar>
        </BarChart>
      </ResponsiveContainer>
    </div>
  );
}

function Skeleton() {
  return (
    <div className="space-y-4 animate-pulse">
      {[...Array(5)].map((_, i) => (
        <div key={i} className="flex items-center gap-3">
          <div className="w-16 h-4 bg-white/5 rounded" />
          <div className="h-8 bg-white/5 rounded-lg" style={{ width: `${100 - i * 15}%` }} />
        </div>
      ))}
    </div>
  );
}

function EmptyState() {
  return (
    <div className="flex flex-col items-center justify-center h-64 text-white/30">
      <svg className="w-12 h-12 mb-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1}>
        <path strokeLinecap="round" strokeLinejoin="round" d="M3 4h18M3 8h18M3 12h12M3 16h6" />
      </svg>
      <p>No funnel data yet</p>
      <p className="text-xs mt-1">Run the load tester to generate events</p>
    </div>
  );
}

function ErrorState() {
  return (
    <div className="flex items-center justify-center h-64 text-red-400/60">
      <p>Failed to load funnel data</p>
    </div>
  );
}
