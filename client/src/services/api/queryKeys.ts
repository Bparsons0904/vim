// Query key factories for consistent cache management
// Following TanStack Query best practices for hierarchical query keys

export const queryKeys = {
  // Root level keys
  all: () => ['api'] as const,

  // User-related queries
  users: () => [...queryKeys.all(), 'users'] as const,
  user: (id?: string) => [...queryKeys.users(), 'user', id] as const,
  userProfile: () => [...queryKeys.users(), 'profile'] as const,
  userSettings: () => [...queryKeys.users(), 'settings'] as const,

  // Session-related queries
  sessions: () => [...queryKeys.all(), 'sessions'] as const,
  session: (id: string) => [...queryKeys.sessions(), 'session', id] as const,

  // LoadTest-related queries
  loadTests: () => [...queryKeys.all(), 'loadtests'] as const,
  loadTest: (id: string) => [...queryKeys.loadTests(), 'loadtest', id] as const,
  loadTestHistory: (filters?: Record<string, unknown>) => 
    [...queryKeys.loadTests(), 'history', filters] as const,

  // Pagination helper
  paginated: (baseKey: readonly unknown[], page: number, limit: number) =>
    [...baseKey, 'paginated', { page, limit }] as const,

  // Search helper
  search: (baseKey: readonly unknown[], query: string) =>
    [...baseKey, 'search', query] as const,

  // Filter helper
  filtered: (baseKey: readonly unknown[], filters: Record<string, unknown>) =>
    [...baseKey, 'filtered', filters] as const,
} as const;

// Utility type for extracting query key types
export type QueryKey = ReturnType<typeof queryKeys[keyof typeof queryKeys]>;

// Helper functions for query invalidation
export const invalidationHelpers = {
  // Invalidate all user-related queries
  invalidateUsers: () => queryKeys.users(),
  
  // Invalidate specific user data
  invalidateUser: (id?: string) => queryKeys.user(id),
  
  // Invalidate all loadtest-related queries
  invalidateLoadTests: () => queryKeys.loadTests(),
  
  // Invalidate specific loadtest data
  invalidateLoadTest: (id: string) => queryKeys.loadTest(id),
  
  // Invalidate all API queries
  invalidateAll: () => queryKeys.all(),
} as const;