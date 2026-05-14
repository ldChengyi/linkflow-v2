import { apiRequest } from './http/client';
import type { PageResult } from './http/types';

export type TenantStatus = 'active' | 'disabled';

export interface Tenant {
  id: string;
  owner_user_id: string;
  tenant_slug: string;
  tenant_name: string;
  status: TenantStatus;
  created_at: string;
  updated_at: string;
}

export interface TenantListParams extends Record<string, unknown> {
  page: number;
  page_size: number;
}

export interface TenantCreateInput {
  tenant_slug: string;
  tenant_name: string;
}

export interface TenantUpdateInput {
  tenant_name: string;
  status: TenantStatus;
}

export interface TenantDeleteResult {
  deleted: boolean;
}

export const listTenants = (params: TenantListParams) => {
  return apiRequest<PageResult<Tenant>>('/api/v1/tenants', {
    method: 'GET',
    params,
  });
};

export const createTenant = (body: TenantCreateInput) => {
  return apiRequest<Tenant>('/api/v1/tenants', {
    method: 'POST',
    data: body,
  });
};

export const updateTenant = (tenantId: string, body: TenantUpdateInput) => {
  return apiRequest<Tenant>(`/api/v1/tenants/${tenantId}`, {
    method: 'PUT',
    data: body,
  });
};

export const deleteTenant = (tenantId: string) => {
  return apiRequest<TenantDeleteResult>(`/api/v1/tenants/${tenantId}`, {
    method: 'DELETE',
  });
};
