import React, { useState, useEffect, useCallback } from 'react';
import {
  type Investment,
  type ValuationHistoryPoint,
  fetchValuationHistory,
} from '../api/investments';
import { formatRupiah } from '../utils/currency';
import { usePrivacy } from '../context/PrivacyContext';
import { TrendingUp, TrendingDown, History, BarChart2, Calendar, RefreshCw } from 'lucide-react';

interface InvestmentValuationChartProps {
  investments: Investment[];
  onRefresh?: () => void;
}

export const InvestmentValuationChart: React.FC<InvestmentValuationChartProps> = ({
  investments,
}) => {
  const { maskValue } = usePrivacy();
  const mask = (val: string): string => (maskValue ? maskValue(val) : val);

  const [history, setHistory] = useState<ValuationHistoryPoint[]>([]);
  const [selectedInvestmentId, setSelectedInvestmentId] = useState<string>('ALL');
  const [viewMode, setViewMode] = useState<'comparison' | 'timeline'>('comparison');
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);
  const [hoveredHistory, setHoveredHistory] = useState<ValuationHistoryPoint | null>(null);

  const loadHistory = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      const filterId = selectedInvestmentId === 'ALL' ? undefined : selectedInvestmentId;
      const data = await fetchValuationHistory(filterId);
      setHistory(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Gagal memuat histori valuasi');
    } finally {
      setLoading(false);
    }
  }, [selectedInvestmentId]);

  useEffect(() => {
    loadHistory();
  }, [loadHistory]);

  const hasInvestments = investments.length > 0;

  // Filtered investments for comparison view
  const displayedInvestments =
    selectedInvestmentId === 'ALL'
      ? investments
      : investments.filter((i) => i.id === selectedInvestmentId);

  // Total summary calculations
  const totalCapital = displayedInvestments.reduce((sum, i) => sum + i.capital, 0);
  const totalCurrent = displayedInvestments.reduce((sum, i) => sum + i.current_value, 0);
  const totalGainLoss = totalCurrent - totalCapital;
  const totalGainPct = totalCapital > 0 ? (totalGainLoss / totalCapital) * 100 : 0;

  // Dimensions for comparison bar chart
  const compSvgWidth = 720;
  const compSvgHeight = Math.max(220, displayedInvestments.length * 48 + 60);
  const compPad = { top: 25, right: 140, bottom: 25, left: 160 };
  const compChartW = compSvgWidth - compPad.left - compPad.right;

  const maxVal = Math.max(
    ...displayedInvestments.map((i) => Math.max(i.capital, i.current_value)),
    1000000
  );

  // Timeline SVG calculations
  const timelineSvgW = 720;
  const timelineSvgH = 240;
  const tPad = { top: 30, right: 35, bottom: 45, left: 75 };
  const tChartW = timelineSvgW - tPad.left - tPad.right;
  const tChartH = timelineSvgH - tPad.top - tPad.bottom;

  const sortedHistory = [...history].sort(
    (a, b) => new Date(a.recorded_at).getTime() - new Date(b.recorded_at).getTime()
  );

  const minTimelineVal = Math.min(...sortedHistory.map((h) => h.valuation), 0);
  const maxTimelineVal = Math.max(...sortedHistory.map((h) => h.valuation), 1000000);
  const timelineRange = maxTimelineVal - minTimelineVal || 1;

  const getTimelineY = (val: number) => {
    const ratio = (val - minTimelineVal) / timelineRange;
    return tPad.top + tChartH - ratio * tChartH;
  };

  const getTimelineX = (idx: number) => {
    if (sortedHistory.length <= 1) return tPad.left + tChartW / 2;
    return tPad.left + (idx / (sortedHistory.length - 1)) * tChartW;
  };

  const timelinePoints = sortedHistory
    .map((h, i) => `${getTimelineX(i)},${getTimelineY(h.valuation)}`)
    .join(' ');

  return (
    <div className="card investment-valuation-chart-card" id="investment-valuation-chart-section">
      {/* Chart Header */}
      <div className="chart-header">
        <div className="chart-title-group">
          <div className="chart-badge-title">
            <BarChart2 size={18} className="text-primary" />
            <h3 className="section-title chart-main-heading">Valuasi & Performa Investasi</h3>
          </div>
          <p className="section-subtitle chart-sub-heading">
            Cek perbandingan modal vs nilai pasar terkini dan pantau histori setelah update valuasi
          </p>
        </div>

        <div className="chart-controls">
          {/* Asset Dropdown Filter */}
          <select
            className="form-select chart-asset-select"
            value={selectedInvestmentId}
            onChange={(e) => setSelectedInvestmentId(e.target.value)}
            aria-label="Pilih Aset Investasi"
          >
            <option value="ALL">Semua Portofolio ({investments.length} Aset)</option>
            {investments.map((inv) => (
              <option key={inv.id} value={inv.id}>
                {inv.name} ({inv.type})
              </option>
            ))}
          </select>

          {/* View Mode Toggle */}
          <div className="segmented-control" role="group">
            <button
              type="button"
              className={`segmented-btn ${viewMode === 'comparison' ? 'segmented-btn--active' : ''}`}
              onClick={() => setViewMode('comparison')}
            >
              📊 Modal vs Nilai
            </button>
            <button
              type="button"
              className={`segmented-btn ${viewMode === 'timeline' ? 'segmented-btn--active' : ''}`}
              onClick={() => setViewMode('timeline')}
            >
              ⏱️ Histori Update ({history.length})
            </button>
          </div>

          <button
            type="button"
            className="btn btn--outline btn--sm btn-icon"
            onClick={loadHistory}
            title="Perbarui Histori"
            disabled={loading}
          >
            <RefreshCw size={14} className={loading ? 'spin' : ''} />
          </button>
        </div>
      </div>

      {/* Summary KPI Cards */}
      <div className="savings-kpi-grid">
        <div className="kpi-card">
          <span className="kpi-label">Total Nilai Pasar Terkini</span>
          <div className="kpi-value text-primary">{mask(formatRupiah(totalCurrent))}</div>
          <span className="kpi-hint">Valuasi aktif setelah update</span>
        </div>

        <div className="kpi-card">
          <span className="kpi-label">Total Modal Diinvestasikan</span>
          <div className="kpi-value">{mask(formatRupiah(totalCapital))}</div>
          <span className="kpi-hint">Akumulasi modal awal aset</span>
        </div>

        <div className="kpi-card">
          <span className="kpi-label">Unrealized Gain / Loss</span>
          <div className="kpi-value-row">
            <span className={`kpi-value ${totalGainLoss >= 0 ? 'text-success' : 'text-danger'}`}>
              {totalGainLoss >= 0 ? '+' : ''}
              {mask(formatRupiah(totalGainLoss))}
            </span>
            <span className={`badge ${totalGainLoss >= 0 ? 'badge--success' : 'badge--danger'} kpi-trend-badge`}>
              {totalGainLoss >= 0 ? <TrendingUp size={12} /> : <TrendingDown size={12} />}
              {totalGainPct.toFixed(1)}%
            </span>
          </div>
          <span className="kpi-hint">Return portofolio saat ini</span>
        </div>

        <div className="kpi-card">
          <span className="kpi-label">Histori Update Valuasi</span>
          <div className="kpi-value">{history.length} Catatan</div>
          <span className="kpi-hint">
            {history.length > 0
              ? `Terakhir: ${new Date(history[0].recorded_at).toLocaleDateString('id-ID')}`
              : 'Belum ada update'}
          </span>
        </div>
      </div>

      {error && (
        <div className="auth-alert auth-alert--error" style={{ margin: '1rem 0' }}>
          <span>⚠️ {error}</span>
          <button className="btn btn--outline btn--sm" onClick={loadHistory} style={{ marginLeft: 'auto' }}>
            Coba Lagi
          </button>
        </div>
      )}

      {!hasInvestments && (
        <div className="empty-state" style={{ padding: '2rem 1rem' }}>
          <p>Belum ada aset investasi. Tambahkan investasi baru untuk mulai memantau valuasi.</p>
        </div>
      )}

      {hasInvestments && viewMode === 'comparison' && (
        <div className="valuation-comparison-view">
          <div className="chart-legend-row">
            <div className="legend-item">
              <span className="legend-dot" style={{ backgroundColor: '#64748b' }} />
              <span>Modal Awal (Capital)</span>
            </div>
            <div className="legend-item">
              <span className="legend-dot" style={{ backgroundColor: '#10b981' }} />
              <span>Nilai Pasar Terkini (Current Valuation)</span>
            </div>
          </div>

          <div className="chart-svg-container" style={{ position: 'relative' }}>
            <svg
              viewBox={`0 0 ${compSvgWidth} ${compSvgHeight}`}
              className="valuation-comparison-svg"
              style={{ width: '100%', height: 'auto', overflow: 'visible' }}
            >
              {displayedInvestments.map((inv, idx) => {
                const y = compPad.top + idx * 48;
                const capWidth = Math.max(4, (inv.capital / maxVal) * compChartW);
                const curWidth = Math.max(4, (inv.current_value / maxVal) * compChartW);
                const isProfit = inv.unrealized_gain >= 0;

                return (
                  <g key={inv.id} className="comp-bar-row">
                    {/* Investment Name */}
                    <text
                      x={compPad.left - 12}
                      y={y + 16}
                      textAnchor="end"
                      fontSize="12"
                      fontWeight="600"
                      fill="var(--text-color, #f1f5f9)"
                    >
                      {inv.name}
                    </text>
                    <text
                      x={compPad.left - 12}
                      y={y + 28}
                      textAnchor="end"
                      fontSize="10"
                      fill="var(--text-muted, #94a3b8)"
                    >
                      {inv.type}
                    </text>

                    {/* Capital Bar */}
                    <rect
                      x={compPad.left}
                      y={y + 4}
                      width={capWidth}
                      height="12"
                      rx="3"
                      fill="#64748b"
                      opacity="0.85"
                    />

                    {/* Current Value Bar */}
                    <rect
                      x={compPad.left}
                      y={y + 18}
                      width={curWidth}
                      height="12"
                      rx="3"
                      fill={isProfit ? '#10b981' : '#ef4444'}
                    />

                    {/* Return Badge */}
                    <g transform={`translate(${compPad.left + Math.max(capWidth, curWidth) + 12}, ${y + 18})`}>
                      <text
                        x="0"
                        y="0"
                        fontSize="11"
                        fontWeight="600"
                        fill={isProfit ? '#10b981' : '#ef4444'}
                      >
                        {isProfit ? '+' : ''}
                        {mask(formatRupiah(inv.unrealized_gain))} ({inv.unrealized_gain_percentage.toFixed(1)}%)
                      </text>
                    </g>
                  </g>
                );
              })}
            </svg>
          </div>
        </div>
      )}

      {/* Timeline View (Valuation Updates Over Time) */}
      {hasInvestments && viewMode === 'timeline' && (
        <div className="valuation-timeline-view">
          {sortedHistory.length === 0 ? (
            <div className="empty-state" style={{ padding: '2rem 1rem' }}>
              <History size={28} className="text-muted" style={{ marginBottom: '0.5rem' }} />
              <p>Belum ada riwayat perubahan valuasi yang tercatat.</p>
              <small className="text-muted">
                Riwayat otomatis tersimpan saat Anda memperbarui valuasi aset lewat tombol Edit / Update Valuasi.
              </small>
            </div>
          ) : (
            <>
              <div className="chart-svg-container" style={{ position: 'relative' }}>
                <svg
                  viewBox={`0 0 ${timelineSvgW} ${timelineSvgH}`}
                  className="valuation-timeline-svg"
                  style={{ width: '100%', height: 'auto', overflow: 'visible' }}
                >
                  <defs>
                    <linearGradient id="valTimelineGrad" x1="0" y1="0" x2="1" y2="0">
                      <stop offset="0%" stopColor="#3b82f6" />
                      <stop offset="100%" stopColor="#10b981" />
                    </linearGradient>
                  </defs>

                  {/* Horizontal Grid lines */}
                  {[0, 0.33, 0.66, 1].map((r) => {
                    const v = minTimelineVal + r * timelineRange;
                    const y = getTimelineY(v);
                    return (
                      <g key={r}>
                        <line
                          x1={tPad.left}
                          y1={y}
                          x2={timelineSvgW - tPad.right}
                          y2={y}
                          stroke="var(--border-color, #334155)"
                          strokeDasharray="3,3"
                          opacity="0.35"
                        />
                        <text
                          x={tPad.left - 8}
                          y={y + 4}
                          textAnchor="end"
                          fontSize="10"
                          fill="var(--text-muted, #94a3b8)"
                          fontFamily="monospace"
                        >
                          {mask(formatRupiah(Math.round(v / 1000) * 1000))}
                        </text>
                      </g>
                    );
                  })}

                  {/* Line */}
                  {sortedHistory.length > 1 && (
                    <polyline
                      fill="none"
                      stroke="url(#valTimelineGrad)"
                      strokeWidth="2.5"
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      points={timelinePoints}
                    />
                  )}

                  {/* Dots */}
                  {sortedHistory.map((h, i) => {
                    const cx = getTimelineX(i);
                    const cy = getTimelineY(h.valuation);
                    const isHovered = hoveredHistory?.id === h.id;

                    return (
                      <g key={h.id}>
                        <circle
                          cx={cx}
                          cy={cy}
                          r={isHovered ? 6 : 4}
                          fill="#10b981"
                          stroke="#ffffff"
                          strokeWidth={isHovered ? '2.5' : '1.5'}
                          style={{ cursor: 'pointer', transition: 'r 0.15s ease' }}
                          onMouseEnter={() => setHoveredHistory(h)}
                          onMouseLeave={() => setHoveredHistory(null)}
                        />
                        <text
                          x={cx}
                          y={timelineSvgH - tPad.bottom + 16}
                          textAnchor="middle"
                          fontSize="9"
                          fill="var(--text-muted, #94a3b8)"
                        >
                          {new Date(h.recorded_at).toLocaleDateString('id-ID', { month: 'short', day: 'numeric' })}
                        </text>
                      </g>
                    );
                  })}
                </svg>

                {/* Hover Tooltip */}
                {hoveredHistory && (
                  <div
                    className="chart-tooltip"
                    style={{
                      position: 'absolute',
                      left: '50%',
                      top: '10px',
                      transform: 'translateX(-50%)',
                      pointerEvents: 'none',
                      zIndex: 10,
                    }}
                  >
                    <div className="tooltip-title">
                      {hoveredHistory.investment_name} ({hoveredHistory.investment_type})
                    </div>
                    <div className="tooltip-row">
                      <span>Valuasi:</span>
                      <strong className="text-primary">{mask(hoveredHistory.formatted_valuation)}</strong>
                    </div>
                    <div className="tooltip-row">
                      <span>Gain / Loss:</span>
                      <strong className={hoveredHistory.unrealized_gain >= 0 ? 'text-success' : 'text-danger'}>
                        {hoveredHistory.unrealized_gain >= 0 ? '+' : ''}
                        {mask(hoveredHistory.formatted_unrealized_gain)} ({hoveredHistory.gain_percentage.toFixed(1)}%)
                      </strong>
                    </div>
                    {hoveredHistory.notes && (
                      <div className="tooltip-row">
                        <span>Catatan:</span>
                        <em>"{hoveredHistory.notes}"</em>
                      </div>
                    )}
                    <div className="tooltip-row">
                      <span>Waktu Catat:</span>
                      <small>{new Date(hoveredHistory.recorded_at).toLocaleString('id-ID')}</small>
                    </div>
                  </div>
                )}
              </div>

              {/* History Data Table */}
              <div className="chart-breakdown-section" style={{ marginTop: '1rem' }}>
                <div className="table-responsive">
                  <table className="data-table">
                    <thead>
                      <tr>
                        <th>Waktu Update</th>
                        <th>Aset</th>
                        <th style={{ textAlign: 'right' }}>Valuasi Baru</th>
                        <th style={{ textAlign: 'right' }}>Gain / Loss</th>
                        <th>Catatan</th>
                      </tr>
                    </thead>
                    <tbody>
                      {history.map((h) => (
                        <tr key={h.id}>
                          <td>
                            <div style={{ display: 'flex', alignItems: 'center', gap: '0.4rem' }}>
                              <Calendar size={13} className="text-muted" />
                              <span>{new Date(h.recorded_at).toLocaleString('id-ID')}</span>
                            </div>
                          </td>
                          <td>
                            <strong>{h.investment_name}</strong>
                            <span className="badge badge--neutral" style={{ marginLeft: '0.4rem', fontSize: '0.75rem' }}>
                              {h.investment_type}
                            </span>
                          </td>
                          <td style={{ textAlign: 'right' }}>
                            <strong>{mask(h.formatted_valuation)}</strong>
                          </td>
                          <td style={{ textAlign: 'right' }} className={h.unrealized_gain >= 0 ? 'text-success' : 'text-danger'}>
                            {h.unrealized_gain >= 0 ? '+' : ''}
                            {mask(h.formatted_unrealized_gain)} ({h.gain_percentage.toFixed(1)}%)
                          </td>
                          <td>
                            <span className="text-muted">{h.notes || '—'}</span>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              </div>
            </>
          )}
        </div>
      )}
    </div>
  );
};
