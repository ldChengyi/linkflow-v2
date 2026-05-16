import { apiRequest } from './http/client';
import type { PageResult } from './http/types';

export type ProductNodeType = 'direct' | 'gateway' | 'sub_device';
export type ProductAuthType = 'secret' | 'certificate' | 'anonymous';
export type ProductProtocolType =
  | 'mqtt'
  | 'http'
  | 'coap'
  | 'modbus'
  | 'opcua'
  | 'lora';
export type ProductStatus = 'active' | 'disabled';

export interface Product {
  id: string;
  tenant_id: string;
  product_key: string;
  product_name: string;
  description: string;
  node_type: ProductNodeType;
  auth_type: ProductAuthType;
  protocol_type: ProductProtocolType;
  status: ProductStatus;
  created_at: string;
  updated_at: string;
}

export interface ProductListParams extends Record<string, unknown> {
  tenant_id: string;
  page: number;
  page_size: number;
}

export interface ProductCreateInput {
  tenant_id: string;
  product_key: string;
  product_name: string;
  description: string;
  node_type: ProductNodeType;
  auth_type: ProductAuthType;
  protocol_type: ProductProtocolType;
}

export interface ProductUpdateInput {
  product_name: string;
  description: string;
  node_type: ProductNodeType;
  auth_type: ProductAuthType;
  protocol_type: ProductProtocolType;
  status: ProductStatus;
}

export interface ProductDeleteResult {
  deleted: boolean;
}

export const listProducts = (params: ProductListParams) => {
  return apiRequest<PageResult<Product>>('/api/v1/products', {
    method: 'GET',
    params,
  });
};

export const createProduct = (body: ProductCreateInput) => {
  return apiRequest<Product>('/api/v1/products', {
    method: 'POST',
    data: body,
  });
};

export const updateProduct = (productId: string, body: ProductUpdateInput) => {
  return apiRequest<Product>(`/api/v1/products/${productId}`, {
    method: 'PUT',
    data: body,
  });
};

export const deleteProduct = (productId: string) => {
  return apiRequest<ProductDeleteResult>(`/api/v1/products/${productId}`, {
    method: 'DELETE',
  });
};
