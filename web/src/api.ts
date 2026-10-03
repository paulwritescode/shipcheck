// Typed client for the ShipCheck backend. All calls go to /api, which Vite
// proxies to the Go dev server (and which API Gateway serves in production).

import type {
  AdvisorAction,
  AdvisorResult,
  ChecklistItem,
  Launch,
  PublicReport,
  Risk,
  ValidationError,
} from './types';

/** ApiError carries the HTTP status and any field-level validation errors. */
export class ApiError extends Error {
  status: number;
  fields: ValidationError[];
  constructor(status: number, message: string, fields: ValidationError[] = []) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
    this.fields = fields;
  }
}

async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
  const res = await fetch(path, {
    method,
    headers: body !== undefined ? { 'Content-Type': 'application/json' } : undefined,
    body: body !== undefined ? JSON.stringify(body) : undefined,
  });

  if (res.status === 204) {
    return undefined as T;
  }

  let data: unknown = null;
  const text = await res.text();
  if (text) {
    try {
      data = JSON.parse(text);
    } catch {
      data = null;
    }
  }

  if (!res.ok) {
    const d = (data ?? {}) as { error?: string; fields?: ValidationError[] };
    throw new ApiError(res.status, d.error ?? `request failed (${res.status})`, d.fields ?? []);
  }
  return data as T;
}

export const api = {
  getLaunch: (id: string) => request<Launch>('GET', `/api/launches/${id}`),

  createLaunch: (input: {
    name: string;
    description?: string;
    targetDate: string;
    owner?: string;
    productArea?: string;
    repositoryUrl?: string;
    documentationUrl?: string;
  }) => request<Launch>('POST', '/api/launches', input),

  seedLaunch: () => request<Launch>('POST', '/api/launches', { seed: true }),

  updateLaunch: (id: string, patch: Record<string, unknown>) =>
    request<Launch>('PATCH', `/api/launches/${id}`, patch),

  addItem: (launchId: string, item: Partial<ChecklistItem>) =>
    request<Launch>('POST', `/api/launches/${launchId}/items`, item),

  updateItem: (launchId: string, item: ChecklistItem) =>
    request<Launch>('PATCH', `/api/launches/${launchId}/items/${item.id}`, item),

  deleteItem: (launchId: string, itemId: string) =>
    request<Launch>('DELETE', `/api/launches/${launchId}/items/${itemId}`),

  addRisk: (launchId: string, risk: Partial<Risk>) =>
    request<Launch>('POST', `/api/launches/${launchId}/risks`, risk),

  updateRisk: (launchId: string, risk: Risk) =>
    request<Launch>('PATCH', `/api/launches/${launchId}/risks/${risk.id}`, risk),

  deleteRisk: (launchId: string, riskId: string) =>
    request<Launch>('DELETE', `/api/launches/${launchId}/risks/${riskId}`),

  share: (launchId: string, action: 'enable' | 'regenerate' | 'disable') =>
    request<Launch>('POST', `/api/launches/${launchId}/share`, { action }),

  analyze: (launchId: string, action: AdvisorAction) =>
    request<AdvisorResult>('POST', `/api/launches/${launchId}/analyze`, { action }),

  getPublicReport: (token: string) => request<PublicReport>('GET', `/api/r/${token}`),
};
