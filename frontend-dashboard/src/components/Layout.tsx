import type { User } from '../types';
import HealthBadge from './HealthBadge';
import FunnelChart from './FunnelChart';
import VolumeChart from './VolumeChart';
import CancellationTable from './CancellationTable';
import RecentEventsTable from './RecentEventsTable';
import { LatencyChart } from './LatencyChart';

interface Props {
  user: User;
  onLogout: () => void;
}

export default function Layout({ user, onLogout }: Props) {
  return (
    <div className="min-h-screen bg-slate-950">
      {/* Header */}
      <header className="border-b border-white/5 bg-slate-950/80 backdrop-blur-xl sticky top-0 z-50">
        <div className="max-w-7xl mx-auto px-6 py-4 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-emerald-500 to-cyan-500 flex items-center justify-center">
              <svg className="w-5 h-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
              </svg>
            </div>
            <div>
              <h1 className="text-lg font-semibold text-white">ONDC Analytics</h1>
              <p className="text-xs text-white/30">Seller Intelligence Gateway</p>
            </div>
          </div>

          <div className="flex items-center gap-4">
            <HealthBadge />
            <div className="flex items-center gap-3 pl-4 border-l border-white/10">
              <div className="text-right">
                <p className="text-sm font-medium text-white/80">{user.name}</p>
                <p className="text-xs text-white/30">{user.role}</p>
              </div>
              <button
                onClick={onLogout}
                className="text-white/40 hover:text-white/80 transition-colors text-sm"
                title="Sign out"
              >
                <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
                  <path strokeLinecap="round" strokeLinejoin="round" d="M15.75 9V5.25A2.25 2.25 0 0013.5 3h-6a2.25 2.25 0 00-2.25 2.25v13.5A2.25 2.25 0 007.5 21h6a2.25 2.25 0 002.25-2.25V15m3 0l3-3m0 0l-3-3m3 3H9" />
                </svg>
              </button>
            </div>
          </div>
        </div>
      </header>

      {/* Dashboard Grid */}
      <main className="max-w-7xl mx-auto px-6 py-8 space-y-6">
        {/* Top row: Funnel + Volume */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <div className="glass-card p-6 animate-slide-up">
            <h2 className="text-lg font-semibold text-white/90 mb-4">Order Funnel</h2>
            <FunnelChart />
          </div>
          <div className="glass-card p-6 animate-slide-up" style={{ animationDelay: '100ms' }}>
            <h2 className="text-lg font-semibold text-white/90 mb-4">Daily Volume</h2>
            <VolumeChart />
          </div>
        </div>

        {/* Middle row: Latency */}
        <div className="grid grid-cols-1 gap-6">
          <div className="glass-card p-6 animate-slide-up" style={{ animationDelay: '150ms' }}>
            <LatencyChart />
          </div>
        </div>

        {/* Bottom row: Cancellations + Recent Events */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <div className="glass-card p-6 animate-slide-up" style={{ animationDelay: '200ms' }}>
            <h2 className="text-lg font-semibold text-white/90 mb-4">Cancellations by City</h2>
            <CancellationTable />
          </div>
          <div className="glass-card p-6 animate-slide-up" style={{ animationDelay: '300ms' }}>
            <h2 className="text-lg font-semibold text-white/90 mb-4">Recent Transactions</h2>
            <RecentEventsTable />
          </div>
        </div>
      </main>
    </div>
  );
}
