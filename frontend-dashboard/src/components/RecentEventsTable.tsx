import { useRecentEvents } from '../api/client';

const TYPE_COLORS: Record<string, string> = {
  on_search: 'bg-emerald-400/10 text-emerald-400',
  on_select: 'bg-teal-400/10 text-teal-400',
  on_init: 'bg-cyan-400/10 text-cyan-400',
  on_confirm: 'bg-indigo-400/10 text-indigo-400',
  on_cancel: 'bg-red-400/10 text-red-400',
};

export default function RecentEventsTable() {
  const { data, isLoading, isError } = useRecentEvents(15);

  if (isLoading) return <div className="h-48 bg-white/5 rounded-xl animate-pulse" />;
  if (isError) return <div className="text-red-400/60">Failed to load</div>;

  const events = data?.events || [];

  if (events.length === 0) {
    return <div className="h-48 flex items-center justify-center text-white/30">No events yet</div>;
  }

  return (
    <div className="overflow-x-auto max-h-[360px] overflow-y-auto">
      <table className="w-full text-sm">
        <thead className="sticky top-0 bg-slate-900/90 backdrop-blur">
          <tr className="border-b border-white/10">
            <th className="text-left py-2 text-white/40 font-medium">Type</th>
            <th className="text-left py-2 text-white/40 font-medium">City</th>
            <th className="text-right py-2 text-white/40 font-medium">Amount</th>
            <th className="text-right py-2 text-white/40 font-medium">Buyer</th>
          </tr>
        </thead>
        <tbody>
          {events.map((e, i) => (
            <tr key={i} className="border-b border-white/5 hover:bg-white/5 transition-colors">
              <td className="py-2">
                <span className={`px-2 py-0.5 rounded-md text-xs font-medium ${TYPE_COLORS[e.event_type] || 'bg-white/10 text-white/60'}`}>
                  {e.event_type.replace('on_', '')}
                </span>
              </td>
              <td className="py-2 text-white/70">{e.city}</td>
              <td className="py-2 text-right text-white/80 font-mono">₹{e.amount.toFixed(0)}</td>
              <td className="py-2 text-right text-white/30 font-mono text-xs">{e.buyer_hash.slice(0, 8)}…</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
