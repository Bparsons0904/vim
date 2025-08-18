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

  // Story-related queries  
  stories: () => [...queryKeys.all(), 'stories'] as const,
  story: (id: string) => [...queryKeys.stories(), 'story', id] as const,
  storyList: (filters?: Record<string, unknown>) => [...queryKeys.stories(), 'list', filters] as const,
  storyParticipants: (storyId: string) => [...queryKeys.stories(), 'participants', storyId] as const,
  storyMembers: (storyId: string) => [...queryKeys.stories(), 'members', storyId] as const,
  createFormData: () => [...queryKeys.stories(), 'createFormData'] as const,
  
  // Invite-related queries
  knownContacts: () => [...queryKeys.all(), 'invites', 'knownContacts'] as const,

  // Session-related queries
  sessions: () => [...queryKeys.all(), 'sessions'] as const,
  session: (id: string) => [...queryKeys.sessions(), 'session', id] as const,

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
  
  // Invalidate all story-related queries
  invalidateStories: () => queryKeys.stories(),
  
  // Invalidate specific story data
  invalidateStory: (id: string) => queryKeys.story(id),
  
  // Invalidate all API queries
  invalidateAll: () => queryKeys.all(),
} as const;