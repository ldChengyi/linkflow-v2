export interface ApiResponse<T> {
  msg: string;
  code: number;
  data?: T;
}

export interface ApiRequestOptions {
  method?: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE';
  data?: unknown;
  params?: Record<string, unknown>;
}
