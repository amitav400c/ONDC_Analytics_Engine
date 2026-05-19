import { useCancellations } from '../api/client';

export default function CancellationTable() {
  const { data, isLoading, isError } = useCancellations();

  if (isLoading) return <div className="h-48 bg-white/5 rounded-xl animate-pulse" />;
  if (isError) return <div className="text-red-400/60">Failed to load</div>;

  const rows = data?.cancellations || [];

  if (rows.length === 0) {
    return <div className="h-48 flex items-center justify-center text-white/30">No cancellation data</div>;
  }

  return (
    <div className="overflow-x-auto">
      <table className="w-full text-sm">
        <thead>
          <tr className="border-b border-white/10">
            <th className="text-left py-2 text-white/40 font-medium">City</th>
            <th className="text-right py-2 text-white/40 font-medium">Count</th>
            <th className="text-right py-2 text-white/40 font-medium">Date</th>
          </tr>
        </thead>
        <tbody>
          {rows.slice(0, 10).map((r, i) => (
            <tr key={i} className="border-b border-white/5 hover:bg-white/5 transition-colors">
              <td className="py-2.5 text-white/80 font-medium">{r.city}</td>
              <td className="py-2.5 text-right text-red-400">{r.count}</td>
              <td className="py-2.5 text-right text-white/30">{r.day.split('T')[0]}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
