import { request } from '@umijs/max';
import { ApiError } from './error';
import type { ApiRequestOptions, ApiResponse } from './types';

const isSuccessCode = (code: number) => code >= 200 && code < 300;

export async function apiRequest<T>(
  url: string,
  options: ApiRequestOptions = {},
): Promise<T> {
  const response = await request<ApiResponse<T>>(url, options);

  if (!isSuccessCode(response.code)) {
    throw new ApiError(response.msg, response.code);
  }

  if (response.data === undefined) {
    const message = 'Empty response data';
    throw new ApiError(message, response.code);
  }

  return response.data;
}
