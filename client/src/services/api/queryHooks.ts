// Reusable query and mutation hooks with axios + TanStack Query integration

import { 
  useQuery, 
  useMutation, 
  useQueryClient, 
  UseQueryOptions, 
  UseMutationOptions,
  UseQueryResult,
  UseMutationResult,
} from '@tanstack/solid-query';
import { getApi, postApi, putApi, patchApi, deleteApi } from './api.service';
import { ApiClientError } from './apiTypes';
import { Accessor } from 'solid-js';
import { useToast } from '../../context/ToastContext';

// Enhanced query options with better defaults
export interface ApiQueryOptions<T> extends Omit<UseQueryOptions<T>, 'queryKey' | 'queryFn'> {
  enabled?: boolean | Accessor<boolean>;
  staleTime?: number;
}

// Enhanced mutation options
export interface ApiMutationOptions<T, V> extends Omit<UseMutationOptions<T, Error, V>, 'mutationFn'> {
  invalidateQueries?: readonly (readonly unknown[])[];
  onSuccessToast?: string | ((data: T, variables: V) => string);
  onErrorToast?: string | ((error: Error) => string);
  onSuccess?: (data: T, variables: V, context: unknown) => void;
  onError?: (error: Error, variables: V, context: unknown) => void;
}

// Generic API query hook
export function useApiQuery<T>(
  queryKey: readonly unknown[],
  url: string,
  params?: Record<string, unknown>,
  options?: ApiQueryOptions<T>
): UseQueryResult<T, Error> {
  return useQuery(() => ({
    queryKey,
    queryFn: () => getApi<T>(url, params),
    refetchOnWindowFocus: false,
    retry: (failureCount, error) => {
      // Don't retry on client errors (4xx)
      if (error instanceof ApiClientError && error.status && error.status >= 400 && error.status < 500) {
        return false;
      }
      // Retry up to 3 times for network errors and server errors
      return failureCount < 3;
    },
    staleTime: 5 * 60 * 1000, // 5 minutes
    ...options,
  }));
}

// Generic API mutation hook
export function useApiMutation<T, V = unknown>(
  method: 'POST' | 'PUT' | 'PATCH' | 'DELETE',
  url: string,
  options?: ApiMutationOptions<T, V>
): UseMutationResult<T, Error, V> {
  const queryClient = useQueryClient();
  const toast = useToast();

  return useMutation(() => ({
    mutationFn: (variables: V) => {
      switch (method) {
        case 'POST':
          return postApi<T, V>(url, variables);
        case 'PUT':
          return putApi<T, V>(url, variables);
        case 'PATCH':
          return patchApi<T, V>(url, variables);
        case 'DELETE':
          return deleteApi<T>(url);
        default:
          throw new Error(`Unsupported method: ${method}`);
      }
    },
    onSuccess: (data, variables, context) => {
      // Invalidate specified queries
      if (options?.invalidateQueries) {
        options.invalidateQueries.forEach(queryKey => {
          queryClient.invalidateQueries({ queryKey });
        });
      }

      // Handle success toast
      if (options?.onSuccessToast) {
        const message = typeof options.onSuccessToast === 'function' 
          ? options.onSuccessToast(data, variables)
          : options.onSuccessToast;
        toast.showSuccess(message);
      }

      // Call original onSuccess
      options?.onSuccess?.(data, variables, context);
    },
    onError: (error, variables, context) => {
      // Handle error toast
      if (options?.onErrorToast) {
        const message = typeof options.onErrorToast === 'function'
          ? options.onErrorToast(error)
          : options.onErrorToast;
        toast.showError(message);
      }

      // Call original onError
      options?.onError?.(error, variables, context);
    },
    ...options,
  }));
}

// Convenience hooks for common operations
export function useApiGet<T>(
  queryKey: readonly unknown[],
  url: string,
  params?: Record<string, unknown>,
  options?: ApiQueryOptions<T>
) {
  return useApiQuery<T>(queryKey, url, params, options);
}

export function useApiPost<T, V = unknown>(
  url: string,
  options?: ApiMutationOptions<T, V>
) {
  return useApiMutation<T, V>('POST', url, options);
}

export function useApiPut<T, V = unknown>(
  url: string,
  options?: ApiMutationOptions<T, V>
) {
  return useApiMutation<T, V>('PUT', url, options);
}

export function useApiPatch<T, V = unknown>(
  url: string,
  options?: ApiMutationOptions<T, V>
) {
  return useApiMutation<T, V>('PATCH', url, options);
}

export function useApiDelete<T>(
  url: string,
  options?: ApiMutationOptions<T, void>
) {
  return useApiMutation<T, void>('DELETE', url, options);
}

// Advanced hooks for common patterns

// Paginated query hook
export function useApiPaginatedQuery<T>(
  baseQueryKey: readonly unknown[],
  url: string,
  page: number,
  limit: number = 10,
  params?: Record<string, unknown>,
  options?: ApiQueryOptions<T>
) {
  const queryKey = [...baseQueryKey, 'paginated', { page, limit, ...params }];
  const queryParams = { page, limit, ...params };
  
  return useApiQuery<T>(queryKey, url, queryParams, options);
}

// Infinite query hook (for "load more" functionality)
export function useApiInfiniteQuery<T>(
  queryKey: readonly unknown[],
  url: string,
  options?: {
    pageParam?: string;
    initialPageParam?: unknown;
    getNextPageParam?: (lastPage: T) => unknown;
  } & ApiQueryOptions<T>
) {
  // This would need to be implemented with useInfiniteQuery from TanStack Query
  // For now, return a regular query
  return useApiQuery<T>(queryKey, url, undefined, options);
}

// Search hook with debouncing
export function useApiSearch<T>(
  baseQueryKey: readonly unknown[],
  url: string,
  searchQuery: Accessor<string>,
  _debounceMs?: number,
  options?: ApiQueryOptions<T>
) {
  // This would need debouncing logic - for now return basic search
  const queryKey = [...baseQueryKey, 'search', searchQuery()];
  
  return useApiQuery<T>(
    queryKey,
    url,
    { q: searchQuery() },
    {
      enabled: () => searchQuery().length > 2, // Only search with 3+ characters
      ...options,
    }
  );
}