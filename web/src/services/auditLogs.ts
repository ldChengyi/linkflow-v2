import { apiRequest } from './http/client';
import type { PageResult } from './http/types';

export type AuditResult = 'success' | 'failure';

export interface AuditLog {
  id: string;
  actor_user_id: string;
  actor_role: string;
  action: string;
  resource_type: string;
  resource_id: string;
  result: AuditResult;
  error_code: string;
  method: string;
  path: string;
  status_code: number;
  duration_ms: number;
  ip: string;
  user_agent: string;
  request_id: string;
  trace_id: string;
  operation_id: string;
  metadata: Record<string, unknown>;
  created_at: string;
}

export interface AuditLogListParams extends Record<string, unknown> {
  page: number;
  page_size: number;
}

export const listAuditLogs = (params: AuditLogListParams) => {
  return apiRequest<PageResult<AuditLog>>('/api/v1/audit-logs', {
    method: 'GET',
    params,
  });
};
