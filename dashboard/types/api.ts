export interface ApiResponse<T> {
  data: T;
  error?: ApiError;
  meta?: ApiMeta;
}

export interface ApiError {
  code: string;
  message: string;
  field?: string;
}

export interface ApiMeta {
  limit?: number;
  offset?: number;
  total?: number;
  page?: number;
}

export interface PaginatedItems<T> {
  items: T[];
  meta: ApiMeta;
}
