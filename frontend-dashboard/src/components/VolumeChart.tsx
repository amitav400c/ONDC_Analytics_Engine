import { useState } from 'react';
import { useVolume } from '../api/client';
import { AreaChart, Area, XAxis, YAxis, Tooltip, ResponsiveContainer } from 'recharts';

export default function VolumeChart() {
  const [days, setDays] = useState(7);
  const { data, isLoading, isError } = useVolume(days);

  if (isLoading) return <div className="h-64 bg-white/5 rounded-xl animate-pulse" />;
  if (isError) return <div className="h-64 flex items-center justify-center text-red-400/60">Failed to load</div>;

  const points = (data?.volume || []).map(p => ({
    day: p.day.split('T')[0].slice(5),
    count: p.count,
  }));

  return (
    <div>
      <div className="flex gap-2 mb-4">
        {[7, 14, 30].map(d => (
          <button key={d} onClick={() => setDays(d)}
            className={`px-3 py-1 rounded-lg text-xs font-medium transition-all ${
              days === d ? 'bg-emerald-500/20 text-emerald-400 border border-emerald-500/30' : 'text-white/30 hover:text-white/60 border border-transparent'
            }`}>{d}D</button>
        ))}
      </div>
      {points.length === 0 ? (
        <div className="h-56 flex flex-col items-center justify-center text-white/30">
          <p>No volume data — run the load tester</p>
        </div>
      ) : (
        <ResponsiveContainer width="100%" height={240}>
          <AreaChart data={points}>
            <defs>
              <linearGradient id="volGrad" x1="0" y1="0" x2="0" y2="1">
                <stop offset="5%" stopColor="#34d399" stopOpacity={0.3} />
                <stop offset="95%" stopColor="#34d399" stopOpacity={0} />
              </linearGradient>
            </defs>
            <XAxis dataKey="day" stroke="#ffffff20" tick={{ fill: '#ffffff40', fontSize: 11 }} />
            <YAxis stroke="#ffffff20" tick={{ fill: '#ffffff40', fontSize: 11 }} />
            <Tooltip contentStyle={{ background: '#1e293b', border: '1px solid #ffffff15', borderRadius: '12px' }} />
            <Area type="monotone" dataKey="count" stroke="#34d399" strokeWidth={2} fill="url(#volGrad)" />
          </AreaChart>
        </ResponsiveContainer>
      )}
    </div>
  );
}
