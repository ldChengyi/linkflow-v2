import { apiRequest } from './http/client';
import type { PageResult } from './http/types';

export type DeviceStatus = 'active' | 'disabled';
export type DeviceConnectionStatus = 'online' | 'offline';

export interface Device {
  id: string;
  tenant_id: string;
  product_id: string;
  device_slug: string;
  device_name: string;
  description: string;
  status: DeviceStatus;
  connection_status: DeviceConnectionStatus;
  gateway_device_id?: string;
  firmware_version?: string;
  ip_address?: string;
  last_seen_at?: string;
  created_at: string;
  updated_at: string;
}

export interface DeviceCreateResult {
  device: Device;
  device_secret?: string;
}

export interface DeviceLatestProperties {
  id: string;
  tenant_id: string;
  product_id: string;
  product_key: string;
  device_slug: string;
  reported: boolean;
  properties: Record<string, unknown>;
  occurred_at?: string;
  received_at?: string;
}

export interface DeviceEventEntry {
  event_id: string;
  tenant_id: string;
  product_id: string;
  product_key: string;
  device_slug: string;
  event_name: string;
  params: Record<string, unknown>;
  occurred_at: string;
  received_at: string;
}

export type DeviceServiceAckStatus = 'success' | 'failed' | 'pending';

export interface DeviceServiceCallResult {
  command_id: string;
  topic: string;
  service_name: string;
  input: Record<string, unknown>;
  sent_at: string;
}

export interface DeviceServiceCallHistoryEntry {
  command_id: string;
  tenant_id: string;
  product_id: string;
  product_key: string;
  device_slug: string;
  service_name: string;
  topic: string;
  input: Record<string, unknown>;
  occurred_at: string;
  ack_deadline_at: string;
  ack_status: DeviceServiceAckStatus;
  ack_event_id?: string;
  ack_success?: boolean;
  ack_code?: string;
  ack_message?: string;
  ack_output?: Record<string, unknown>;
  ack_occurred_at?: string;
  ack_received_at?: string;
  ack_deadline_seconds: number;
}

export interface DeviceListParams extends Record<string, unknown> {
  tenant_id: string;
  product_id?: string;
  page: number;
  page_size: number;
}

export interface DeviceEventHistoryParams extends Record<string, unknown> {
  event_name?: string;
  page: number;
  page_size: number;
}

export interface DeviceServiceCallHistoryParams extends Record<string, unknown> {
  service_name?: string;
  ack_deadline_seconds: number;
  page: number;
  page_size: number;
}

export interface DeviceCreateInput {
  tenant_id: string;
  product_id: string;
  device_slug: string;
  device_name: string;
  description: string;
  gateway_device_id: string;
}

export interface DeviceUpdateInput {
  device_name: string;
  description: string;
  status: DeviceStatus;
  gateway_device_id: string;
}

export interface DeviceServiceCallInput {
  input: Record<string, unknown>;
}

export interface DeviceDeleteResult {
  deleted: boolean;
}

export const listDevices = (params: DeviceListParams) => {
  return apiRequest<PageResult<Device>>('/api/v1/devices', {
    method: 'GET',
    params,
  });
};

export const getDeviceLatestProperties = (deviceId: string) => {
  return apiRequest<DeviceLatestProperties>(
    `/api/v1/devices/${deviceId}/properties/latest`,
    {
      method: 'GET',
    },
  );
};

export const listDeviceEvents = (
  deviceId: string,
  params: DeviceEventHistoryParams,
) => {
  return apiRequest<PageResult<DeviceEventEntry>>(
    `/api/v1/devices/${deviceId}/events`,
    {
      method: 'GET',
      params,
    },
  );
};

export const callDeviceService = (
  deviceId: string,
  serviceName: string,
  body: DeviceServiceCallInput,
) => {
  return apiRequest<DeviceServiceCallResult>(
    `/api/v1/devices/${deviceId}/services/${serviceName}/call`,
    {
      method: 'POST',
      data: body,
    },
  );
};

export const listDeviceServiceCalls = (
  deviceId: string,
  params: DeviceServiceCallHistoryParams,
) => {
  return apiRequest<PageResult<DeviceServiceCallHistoryEntry>>(
    `/api/v1/devices/${deviceId}/services/history`,
    {
      method: 'GET',
      params,
    },
  );
};

export const createDevice = (body: DeviceCreateInput) => {
  return apiRequest<DeviceCreateResult>('/api/v1/devices', {
    method: 'POST',
    data: body,
  });
};

export const updateDevice = (deviceId: string, body: DeviceUpdateInput) => {
  return apiRequest<Device>(`/api/v1/devices/${deviceId}`, {
    method: 'PUT',
    data: body,
  });
};

export const deleteDevice = (deviceId: string) => {
  return apiRequest<DeviceDeleteResult>(`/api/v1/devices/${deviceId}`, {
    method: 'DELETE',
  });
};
