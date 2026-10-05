import { getToken } from './auth';
import { API_BASE_URL } from '../config';

export interface BudgetVsActualItem {
  category_id?: string;
  category: string;
  category_icon?: string;
  budget: number;
  actual: number;
  variance: number;
}

export interface AccountAssetItem {
  id: string;
  name: string;
  type: string;
  balance: number;
}

export interface ReceivableAssetItem {
  id: string;
  counterparty: string;
  principal: number;
  remaining_amount: number;
  status: string;
}

export interface InvestmentAssetItem {
  id: string;
  name: string;
  type: string;
  capital: number;
  current_value: number;
}

export interface AssetSnapshot {
  accounts: number;
  receivables: number;
  investments: number;
  total: number;
  account_list?: AccountAssetItem[];
  receivable_list?: ReceivableAssetItem[];
  investment_list?: InvestmentAssetItem[];
}

export interface ReportData {
  period: string;
  start_date: string;
  end_date: string;
  income: number;
  expense: number;
  transfer: number;
  net_cash_flow: number;
  budget_vs_actual: BudgetVsActualItem[];
  assets: AssetSnapshot;
  liabilities: number;
  net_worth: number;
}

async function authFetch(url: string, options: RequestInit = {}): Promise<Response> {
  const token = getToken();
  if (!token) {
    throw new Error('Not authenticated');
  }

  const headers = new Headers(options.headers || {});
  headers.set('Authorization', `Bearer ${token}`);
  if (!headers.has('Content-Type') && options.body) {
    headers.set('Content-Type', 'application/json');
  }

  return fetch(`${API_BASE_URL}${url}`, {
    ...options,
    headers,
  });
}

export async function fetchMonthlyReport(year: number, month: number): Promise<ReportData> {
  const paddedMonth = month.toString().padStart(2, '0');
  const res = await authFetch(`/api/v1/reports/monthly?year=${year}&month=${paddedMonth}`);
  const json = await res.json();
  if (!res.ok || !json.success) {
    throw new Error(json.message || 'Failed to fetch monthly report');
  }
  return json.data;
}

export async function fetchCycleReport(date?: string): Promise<ReportData> {
  const query = date ? `?date=${encodeURIComponent(date)}` : '';
  const res = await authFetch(`/api/v1/reports/cycle${query}`);
  const json = await res.json();
  if (!res.ok || !json.success) {
    throw new Error(json.message || 'Failed to fetch cycle report');
  }
  return json.data;
}
