import React, { useState, useEffect, useCallback } from 'react';
import { fetchSavingsTrend, type SavingsTrendResponse, type SavingsTrendPoint } from '../api/reports';
import { formatRupiah } from '../utils/currency';
import { usePrivacy } from '../context/PrivacyContext';
import { TrendingUp, TrendingDown, Minus, RefreshCw, BarChart3, AlertCircle } from 'lucide-react';

interface SavingsCycleChartProps {
  initialCycles?: number;
  showSelector?: boolean;
}

export const SavingsCycleChart: React.FC<SavingsCycleChartProps> = ({
  initialCycles = 6,
  showSelector = true,
}) => {
  const { maskValue } = usePrivacy();
  const mask = (val: string): string => (maskValue ? maskValue(val) : val);

  const [numCycles, setNumCycles] = useState<number>(initialCycles);
  const [data, setData] = useState<SavingsTrendResponse | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);
  const [hoveredPoint, setHoveredPoint] = useState<{ point: SavingsTrendPoint; index: number; x: number; y: number } | null>(null);

  const loadData = useCallback(async (cycles: number) => {
    try {
      setLoading(true);
      setError(null);
      const res = await fetchSavingsTrend(cycles);
      setData(res);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Gagal memuat data tren tabungan');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    loadData(numCycles);
  }, [numCycles, loadData]);

  const cycles = data?.cycles || [];
  const hasData = cycles.length > 0;

  // Chart dimensions
  const svgWidth = 720;
  const svgHeight = 260;
  const padding = { top: 35, right: 30, bottom: 45, left: 65 };
  const chartWidth = svgWidth - padding.left - padding.right;
  const chartHeight = svgHeight - padding.top - padding.bottom;

  // Calculate scales
  const maxSavings = Math.max(...cycles.map((c) => Math.max(c.net_savings, 0)), 1000000);
  const minSavings = Math.min(...cycles.map((c) => Math.min(c.net_savings, 0)), 0);
  const range = maxSavings - minSavings || 1;

  const getY = (val: number) => {
    const ratio = (val - minSavings) / range;
    return padding.top + chartHeight - ratio * chartHeight;
  };

  const zeroY = getY(0);

  const stepX = cycles.length > 1 ? chartWidth / (cycles.length - 1) : chartWidth / 2;
  const getX = (index: number) => {
    if (cycles.length === 1) return padding.left + chartWidth / 2;
    return padding.left + index * stepX;
  };

  // Trend line path
  const linePoints = cycles.map((c, i) => `${getX(i)},${getY(c.net_savings)}`).join(' ');

  return (
    <div className="card savings-cycle-chart-card" id="savings-cycle-chart-section">
      {/* Header */}
      <div className="chart-header">
        <div className="chart-title-group">
          <div className="chart-badge-title">
            <BarChart3 size={18} className="text-primary" />
            <h3 className="section-title chart-main-heading">Tren Tabungan Antar Siklus</h3>
          </div>
          <p className="section-subtitle chart-sub-heading">
            Pantau pertumbuhan atau penurunan net savings per siklus yang telah tercatat
          </p>
        </div>

        <div className="chart-controls">
          {showSelector && (
            <div className="segmented-control chart-period-selector" role="group" aria-label="Jumlah Siklus">
              {[3, 6, 12].map((n) => (
                <button
                  key={n}
                  type="button"
                  className={`segmented-btn ${numCycles === n ? 'segmented-btn--active' : ''}`}
                  onClick={() => setNumCycles(n)}
                >
                  {n} Siklus
                </button>
              ))}
            </div>
          )}
          <button
            type="button"
            className="btn btn--outline btn--sm btn-icon"
            onClick={() => loadData(numCycles)}
            title="Muat Ulang Data"
            disabled={loading}
          >
            <RefreshCw size={14} className={loading ? 'spin' : ''} />
          </button>
        </div>
      </div>

      {/* KPI Highlight Summary Cards */}
      {data && (
        <div className="savings-kpi-grid">
          {/* Latest Status */}
          <div className="kpi-card">
            <span className="kpi-label">Performa Siklus Terakhir</span>
            <div className="kpi-value-row">
              <span className="kpi-value">
                {mask(formatRupiah(cycles[cycles.length - 1]?.net_savings || 0))}
              </span>
              {data.current_status === 'INCREASED' && (
                <span className="badge badge--success kpi-trend-badge">
                  <TrendingUp size={14} /> Naik {data.current_vs_previous_diff > 0 ? `+${mask(formatRupiah(data.current_vs_previous_diff))}` : ''}
                </span>
              )}
              {data.current_status === 'DECREASED' && (
                <span className="badge badge--danger kpi-trend-badge">
                  <TrendingDown size={14} /> Turun {mask(formatRupiah(Math.abs(data.current_vs_previous_diff)))}
                </span>
              )}
              {data.current_status === 'STABLE' && (
                <span className="badge badge--neutral kpi-trend-badge">
                  <Minus size={14} /> Stabil
                </span>
              )}
            </div>
            <span className="kpi-hint">
              {cycles[cycles.length - 1]?.label || 'Siklus saat ini'}
            </span>
          </div>

          {/* Average Savings */}
          <div className="kpi-card">
            <span className="kpi-label">Rata-rata Tabungan</span>
            <div className="kpi-value">{mask(data.formatted_average_savings || formatRupiah(data.average_savings))}</div>
            <span className="kpi-hint">Dari {cycles.length} siklus tercatat</span>
          </div>

          {/* Peak Savings */}
          <div className="kpi-card">
            <span className="kpi-label">Tabungan Tertinggi</span>
            <div className="kpi-value text-success">{mask(data.formatted_highest_savings || formatRupiah(data.highest_savings))}</div>
            <span className="kpi-hint">Rekor siklus tabungan terbaik</span>
          </div>

          {/* Lowest Savings */}
          <div className="kpi-card">
            <span className="kpi-label">Tabungan Terendah</span>
            <div className="kpi-value">{mask(data.formatted_lowest_savings || formatRupiah(data.lowest_savings))}</div>
            <span className="kpi-hint">Batas minimum siklus</span>
          </div>
        </div>
      )}

      {/* Main Chart Canvas */}
      {loading && !data && (
        <div className="chart-loading-state">
          <div className="spinner" />
          <p>Memuat tren tabungan siklus…</p>
        </div>
      )}

      {error && (
        <div className="auth-alert auth-alert--error" style={{ margin: '1rem 0' }}>
          <AlertCircle size={16} />
          <span>{error}</span>
          <button className="btn btn--outline btn--sm" onClick={() => loadData(numCycles)} style={{ marginLeft: 'auto' }}>
            Coba Lagi
          </button>
        </div>
      )}

      {!loading && !error && !hasData && (
        <div className="empty-state" style={{ padding: '2rem 1rem' }}>
          <p>Belum ada data transaksi yang tercatat untuk menghitung siklus tabungan.</p>
        </div>
      )}

      {hasData && (
        <div className="chart-svg-container" style={{ position: 'relative' }}>
          <svg
            viewBox={`0 0 ${svgWidth} ${svgHeight}`}
            className="savings-trend-svg"
            style={{ width: '100%', height: 'auto', overflow: 'visible' }}
          >
            <defs>
              <linearGradient id="savingsBarGradientUp" x1="0" y1="0" x2="0" y2="1">
                <stop offset="0%" stopColor="#10b981" stopOpacity="0.85" />
                <stop offset="100%" stopColor="#059669" stopOpacity="0.4" />
              </linearGradient>
              <linearGradient id="savingsBarGradientDown" x1="0" y1="0" x2="0" y2="1">
                <stop offset="0%" stopColor="#ef4444" stopOpacity="0.85" />
                <stop offset="100%" stopColor="#b91c1c" stopOpacity="0.4" />
              </linearGradient>
              <linearGradient id="savingsLineGrad" x1="0" y1="0" x2="1" y2="0">
                <stop offset="0%" stopColor="#3b82f6" />
                <stop offset="100%" stopColor="#8b5cf6" />
              </linearGradient>
              <filter id="shadowFilter" x="-10%" y="-10%" width="120%" height="120%">
                <feDropShadow dx="0" dy="2" stdDeviation="3" floodOpacity="0.25" />
              </filter>
            </defs>

            {/* Horizontal Grid lines */}
            {[0, 0.25, 0.5, 0.75, 1].map((ratio) => {
              const val = minSavings + ratio * range;
              const y = getY(val);
              return (
                <g key={ratio} className="chart-grid-line">
                  <line
                    x1={padding.left}
                    y1={y}
                    x2={svgWidth - padding.right}
                    y2={y}
                    stroke="var(--border-color, #334155)"
                    strokeDasharray={ratio === 0 || val === 0 ? 'none' : '3,3'}
                    strokeWidth={val === 0 ? '1.5' : '1'}
                    opacity={val === 0 ? 0.7 : 0.3}
                  />
                  <text
                    x={padding.left - 8}
                    y={y + 4}
                    textAnchor="end"
                    fontSize="10"
                    fill="var(--text-muted, #94a3b8)"
                    fontFamily="monospace"
                  >
                    {mask(formatRupiah(Math.round(val / 1000) * 1000))}
                  </text>
                </g>
              );
            })}

            {/* Zero Base Line */}
            {minSavings < 0 && maxSavings > 0 && (
              <line
                x1={padding.left}
                y1={zeroY}
                x2={svgWidth - padding.right}
                y2={zeroY}
                stroke="#64748b"
                strokeWidth="1.5"
              />
            )}

            {/* Columns & Delta Badges */}
            {cycles.map((pt, idx) => {
              const cx = getX(idx);
              const barWidth = Math.min(42, Math.max(20, (chartWidth / cycles.length) * 0.45));
              const barY = pt.net_savings >= 0 ? getY(pt.net_savings) : zeroY;
              const barH = Math.max(3, Math.abs(getY(pt.net_savings) - zeroY));
              const isUp = pt.status === 'INCREASED';
              const isDown = pt.status === 'DECREASED';

              return (
                <g key={pt.period} className="chart-bar-group">
                  {/* Vertical bar */}
                  <rect
                    x={cx - barWidth / 2}
                    y={barY}
                    width={barWidth}
                    height={barH}
                    rx="4"
                    fill={pt.net_savings >= 0 ? 'url(#savingsBarGradientUp)' : 'url(#savingsBarGradientDown)'}
                    className="chart-bar"
                    style={{ cursor: 'pointer', transition: 'all 0.2s ease' }}
                    onMouseEnter={() => setHoveredPoint({ point: pt, index: idx, x: cx, y: barY })}
                    onMouseLeave={() => setHoveredPoint(null)}
                  />

                  {/* Change Delta Badge above bar */}
                  {idx > 0 && (
                    <g transform={`translate(${cx}, ${barY - 14})`}>
                      <rect
                        x="-24"
                        y="-10"
                        width="48"
                        height="15"
                        rx="7.5"
                        fill={isUp ? 'rgba(16, 185, 129, 0.15)' : isDown ? 'rgba(239, 68, 68, 0.15)' : 'rgba(148, 163, 184, 0.15)'}
                        stroke={isUp ? '#10b981' : isDown ? '#ef4444' : '#94a3b8'}
                        strokeWidth="1"
                      />
                      <text
                        x="0"
                        y="1"
                        textAnchor="middle"
                        fontSize="9"
                        fontWeight="600"
                        fill={isUp ? '#10b981' : isDown ? '#ef4444' : '#94a3b8'}
                      >
                        {isUp ? `+${pt.change_percentage.toFixed(0)}%` : isDown ? `${pt.change_percentage.toFixed(0)}%` : '0%'}
                      </text>
                    </g>
                  )}

                  {/* X-axis Label */}
                  <text
                    x={cx}
                    y={svgHeight - padding.bottom + 18}
                    textAnchor="middle"
                    fontSize="11"
                    fontWeight="500"
                    fill="var(--text-color, #cbd5e1)"
                  >
                    {pt.label}
                  </text>
                  <text
                    x={cx}
                    y={svgHeight - padding.bottom + 32}
                    textAnchor="middle"
                    fontSize="9"
                    fill="var(--text-muted, #94a3b8)"
                  >
                    {pt.start_date.slice(5)}
                  </text>
                </g>
              );
            })}

            {/* Connecting Trend Polyline */}
            {cycles.length > 1 && (
              <polyline
                fill="none"
                stroke="url(#savingsLineGrad)"
                strokeWidth="2.5"
                strokeLinecap="round"
                strokeLinejoin="round"
                points={linePoints}
                filter="url(#shadowFilter)"
              />
            )}

            {/* Data Dots on Trend Line */}
            {cycles.map((pt, idx) => {
              const cx = getX(idx);
              const cy = getY(pt.net_savings);
              const isHovered = hoveredPoint?.index === idx;

              return (
                <g key={`dot-${pt.period}`}>
                  <circle
                    cx={cx}
                    cy={cy}
                    r={isHovered ? 6 : 4}
                    fill={pt.net_savings >= 0 ? '#10b981' : '#ef4444'}
                    stroke="#ffffff"
                    strokeWidth={isHovered ? '2.5' : '1.5'}
                    style={{ cursor: 'pointer', transition: 'r 0.15s ease' }}
                    onMouseEnter={() => setHoveredPoint({ point: pt, index: idx, x: cx, y: cy })}
                    onMouseLeave={() => setHoveredPoint(null)}
                  />
                </g>
              );
            })}
          </svg>

          {/* Interactive Hover Tooltip */}
          {hoveredPoint && (
            <div
              className="chart-tooltip"
              style={{
                position: 'absolute',
                left: `${(hoveredPoint.x / svgWidth) * 100}%`,
                top: `${(hoveredPoint.y / svgHeight) * 100}%`,
                transform: 'translate(-50%, -115%)',
                pointerEvents: 'none',
                zIndex: 10,
              }}
            >
              <div className="tooltip-title">{hoveredPoint.point.label} ({hoveredPoint.point.period})</div>
              <div className="tooltip-row">
                <span>Pemasukan:</span>
                <strong className="text-success">{mask(hoveredPoint.point.formatted_income)}</strong>
              </div>
              <div className="tooltip-row">
                <span>Pengeluaran:</span>
                <strong className="text-danger">{mask(hoveredPoint.point.formatted_expense)}</strong>
              </div>
              <div className="tooltip-divider" />
              <div className="tooltip-row">
                <span>Net Savings:</span>
                <strong className={hoveredPoint.point.net_savings >= 0 ? 'text-success' : 'text-danger'}>
                  {mask(hoveredPoint.point.formatted_net_savings)}
                </strong>
              </div>
              <div className="tooltip-row">
                <span>Saving Rate:</span>
                <strong>{hoveredPoint.point.saving_rate.toFixed(1)}%</strong>
              </div>
              {hoveredPoint.index > 0 && (
                <div className="tooltip-row">
                  <span>Delta vs Siklus Lalu:</span>
                  <strong className={hoveredPoint.point.status === 'INCREASED' ? 'text-success' : hoveredPoint.point.status === 'DECREASED' ? 'text-danger' : ''}>
                    {hoveredPoint.point.status === 'INCREASED' ? '▲ Naik ' : hoveredPoint.point.status === 'DECREASED' ? '▼ Turun ' : '— '}
                    {mask(hoveredPoint.point.formatted_change_amount)} ({hoveredPoint.point.change_percentage.toFixed(1)}%)
                  </strong>
                </div>
              )}
            </div>
          )}
        </div>
      )}

      {/* Cycle Breakdown Table */}
      {hasData && (
        <div className="chart-breakdown-section">
          <div className="table-responsive">
            <table className="data-table">
              <thead>
                <tr>
                  <th>Siklus</th>
                  <th>Periode</th>
                  <th style={{ textAlign: 'right' }}>Pemasukan</th>
                  <th style={{ textAlign: 'right' }}>Pengeluaran</th>
                  <th style={{ textAlign: 'right' }}>Tabungan (Net)</th>
                  <th style={{ textAlign: 'center' }}>Saving Rate</th>
                  <th style={{ textAlign: 'right' }}>Perubahan Siklus</th>
                </tr>
              </thead>
              <tbody>
                {cycles.map((c, i) => (
                  <tr key={c.period}>
                    <td>
                      <strong>{c.label}</strong>
                    </td>
                    <td className="text-muted" style={{ fontSize: '0.85rem' }}>
                      {c.start_date} s/d {c.end_date}
                    </td>
                    <td style={{ textAlign: 'right' }}>{mask(c.formatted_income)}</td>
                    <td style={{ textAlign: 'right' }}>{mask(c.formatted_expense)}</td>
                    <td style={{ textAlign: 'right' }} className={c.net_savings >= 0 ? 'text-success' : 'text-danger'}>
                      <strong>{mask(c.formatted_net_savings)}</strong>
                    </td>
                    <td style={{ textAlign: 'center' }}>
                      <span className="badge badge--neutral">{c.saving_rate.toFixed(1)}%</span>
                    </td>
                    <td style={{ textAlign: 'right' }}>
                      {i === 0 ? (
                        <span className="text-muted">— (Awal)</span>
                      ) : c.status === 'INCREASED' ? (
                        <span className="badge badge--success">
                          ▲ +{mask(c.formatted_change_amount)} (+{c.change_percentage.toFixed(1)}%)
                        </span>
                      ) : c.status === 'DECREASED' ? (
                        <span className="badge badge--danger">
                          ▼ -{mask(c.formatted_change_amount)} ({c.change_percentage.toFixed(1)}%)
                        </span>
                      ) : (
                        <span className="badge badge--neutral">Stabil (0%)</span>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </div>
  );
};
