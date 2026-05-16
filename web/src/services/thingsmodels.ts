import { apiRequest } from './http/client';
import type { PageResult } from './http/types';

export type ThingsModelStatus = 'draft' | 'published' | 'deprecated';
export type ThingsModelObject = Record<string, unknown>;

export interface ThingsModel {
  id: string;
  tenant_id: string;
  product_id: string;
  model_version: number;
  model_name: string;
  description: string;
  status: ThingsModelStatus;
  is_current: boolean;
  properties: ThingsModelObject;
  events: ThingsModelObject;
  services: ThingsModelObject;
  published_at?: string;
  created_at: string;
  updated_at: string;
}

export interface ThingsModelListParams extends Record<string, unknown> {
  tenant_id: string;
  product_id?: string;
  page: number;
  page_size: number;
}

export interface ThingsModelCreateInput {
  tenant_id: string;
  product_id: string;
  model_version: number;
  model_name: string;
  description: string;
  status: ThingsModelStatus;
  is_current: boolean;
  properties: ThingsModelObject;
  events: ThingsModelObject;
  services: ThingsModelObject;
}

export interface ThingsModelUpdateInput {
  model_name: string;
  description: string;
  status: ThingsModelStatus;
  is_current: boolean;
  properties: ThingsModelObject;
  events: ThingsModelObject;
  services: ThingsModelObject;
}

export interface ThingsModelDeleteResult {
  deleted: boolean;
}

export const listThingsModels = (params: ThingsModelListParams) => {
  return apiRequest<PageResult<ThingsModel>>('/api/v1/thingsmodels', {
    method: 'GET',
    params,
  });
};

export const createThingsModel = (body: ThingsModelCreateInput) => {
  return apiRequest<ThingsModel>('/api/v1/thingsmodels', {
    method: 'POST',
    data: body,
  });
};

export const updateThingsModel = (
  thingsModelId: string,
  body: ThingsModelUpdateInput,
) => {
  return apiRequest<ThingsModel>(`/api/v1/thingsmodels/${thingsModelId}`, {
    method: 'PUT',
    data: body,
  });
};

export const deleteThingsModel = (thingsModelId: string) => {
  return apiRequest<ThingsModelDeleteResult>(
    `/api/v1/thingsmodels/${thingsModelId}`,
    {
      method: 'DELETE',
    },
  );
};
