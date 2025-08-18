# API Implementation Guide

This guide walks through the comprehensive API layer implementation using Axios + TanStack Query with TypeScript support, designed for minimal boilerplate and excellent developer experience.

## 🏗️ Architecture Overview

The API layer is built with these core principles:
- **Type Safety**: Full TypeScript support with generics
- **Minimal Boilerplate**: Reusable hooks and utilities
- **Consistent Patterns**: Standardized approach across all endpoints
- **Error Handling**: Centralized error management with custom error classes
- **Cache Management**: Hierarchical query keys for efficient invalidation
- **Developer Experience**: Toast notifications, loading states, and automatic retries

## 📁 File Structure

```
src/services/api/
├── apiTypes.ts           # Core type definitions and error classes
├── api.service.ts        # Enhanced axios client with generic methods
├── queryKeys.ts          # Hierarchical query key factories
├── queryHooks.ts         # Reusable TanStack Query hooks
└── endpoints/
    ├── users.api.ts      # User-specific API functions
    └── stories.api.ts    # Story-specific API functions
```

## 🔧 Core Components

### 1. Type Definitions (`apiTypes.ts`)

Provides foundational types and error handling:

```typescript
// Generic API response wrapper
export interface ApiResponse<T = unknown> {
  data?: T;
  error?: ApiError;
  message?: string;
}

// Custom error classes for better error handling
export class ApiClientError extends Error {
  constructor(
    message: string, 
    public status?: number, 
    public code?: string, 
    public details?: Record<string, unknown>
  ) {
    super(message);
    this.name = 'ApiClientError';
  }
}

export class NetworkError extends Error {
  constructor(message: string = 'Network request failed') {
    super(message);
    this.name = 'NetworkError';
  }
}
```

### 2. Enhanced Axios Client (`api.service.ts`)

Generic HTTP client with comprehensive error handling:

```typescript
// Generic API request function
export const apiRequest = async <T>(
  method: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE',
  url: string,
  data?: unknown,
  config?: RequestConfig
): Promise<T> => {
  try {
    const response = await axiosInstance.request<T>({
      method,
      url,
      data,
      ...config,
    });
    return response.data;
  } catch (error) {
    // Comprehensive error handling with custom error classes
    throw handleApiError(error);
  }
};

// Convenience methods for each HTTP verb
export const apiGet = <T>(url: string, config?: RequestConfig) => 
  apiRequest<T>('GET', url, undefined, config);

export const apiPost = <T>(url: string, data?: unknown, config?: RequestConfig) => 
  apiRequest<T>('POST', url, data, config);

// ... other HTTP methods
```

### 3. Query Key Management (`queryKeys.ts`)

Hierarchical factory system for consistent cache management:

```typescript
export const queryKeys = {
  // Root level
  all: () => ['api'] as const,
  
  // User queries
  users: () => [...queryKeys.all(), 'users'] as const,
  user: (id?: string) => [...queryKeys.users(), 'user', id] as const,
  userProfile: () => [...queryKeys.users(), 'profile'] as const,
  
  // Story queries  
  stories: () => [...queryKeys.all(), 'stories'] as const,
  story: (id?: string) => [...queryKeys.stories(), 'story', id] as const,
  userStories: (userId?: string) => [...queryKeys.stories(), 'user', userId] as const,
  
  // Pagination support
  storiesPaginated: (page?: number, limit?: number) => 
    [...queryKeys.stories(), 'paginated', { page, limit }] as const,
} as const;
```

**Benefits:**
- **Consistent Keys**: All query keys follow the same hierarchical pattern
- **Easy Invalidation**: Invalidate related queries easily (`queryClient.invalidateQueries(queryKeys.users())`)
- **Type Safety**: All keys are strongly typed with `as const`
- **Scalable**: Easy to add new query key families

### 4. Reusable Query Hooks (`queryHooks.ts`)

Generic hooks that eliminate boilerplate across the application:

```typescript
// Generic query hook with built-in error handling and loading states
export function useApiQuery<T>(
  queryKey: readonly unknown[],
  url: string,
  params?: Record<string, unknown>,
  options?: ApiQueryOptions<T>
): UseQueryResult<T, Error> {
  return createQuery({
    queryKey: params ? [...queryKey, params] : queryKey,
    queryFn: () => apiGet<T>(url, params ? { params } : undefined),
    throwOnError: false,
    retry: 3,
    ...options?.queryOptions,
  });
}

// Generic mutation hook with toast notifications and cache invalidation
export function useApiPost<TResponse, TRequest = unknown>(
  url: string,
  options?: ApiMutationOptions<TResponse, TRequest>
): CreateMutationResult<TResponse, Error, TRequest> {
  const queryClient = useQueryClient();
  
  return createMutation({
    mutationFn: (data: TRequest) => apiPost<TResponse>(url, data),
    onSuccess: (data, variables) => {
      // Automatic cache invalidation
      if (options?.invalidateQueries) {
        options.invalidateQueries.forEach(queryKey => {
          queryClient.invalidateQueries({ queryKey });
        });
      }
      
      // Success toast notification
      if (options?.onSuccessToast) {
        const message = typeof options.onSuccessToast === 'function' 
          ? options.onSuccessToast(data, variables)
          : options.onSuccessToast;
        // Show toast (implementation depends on your toast library)
        console.log('Success:', message);
      }
      
      options?.onSuccess?.(data, variables);
    },
    onError: (error) => {
      // Error toast notification
      if (options?.onErrorToast) {
        const message = typeof options.onErrorToast === 'function'
          ? options.onErrorToast(error)
          : options.onErrorToast;
        console.error('Error:', message);
      }
      
      options?.onError?.(error);
    },
    ...options?.mutationOptions,
  });
}
```

### 5. Endpoint-Specific API Functions (`endpoints/users.api.ts`)

Domain-specific functions using the reusable hooks:

```typescript
// Login mutation with automatic auth context updates
export function useLoginUser() {
  return useApiPost<User, LoginCredentials>('users/login', {
    invalidateQueries: [queryKeys.userProfile()],
    onSuccessToast: 'Successfully logged in!',
    onErrorToast: (error) => `Login failed: ${error.message}`,
  });
}

// User profile query with automatic refetching
export function useUserProfile() {
  return useApiQuery<User>(
    queryKeys.userProfile(),
    'users/profile',
    undefined,
    {
      staleTime: 5 * 60 * 1000, // 5 minutes
      refetchOnWindowFocus: true,
    }
  );
}

// User registration with validation
export function useRegisterUser() {
  return useApiPost<User, RegisterData>('users/register', {
    invalidateQueries: [queryKeys.users()],
    onSuccessToast: 'Account created successfully!',
    onErrorToast: (error) => `Registration failed: ${error.message}`,
  });
}
```

## 🚀 Usage Examples

### Basic Query Usage

```typescript
function UserProfile() {
  const { data: user, isLoading, error } = useUserProfile();
  
  if (isLoading) return <div>Loading...</div>;
  if (error) return <div>Error: {error.message}</div>;
  
  return <div>Welcome, {user?.name}!</div>;
}
```

### Mutation with Form Handling

```typescript
function LoginForm() {
  const loginMutation = useLoginUser();
  
  const handleSubmit = (credentials: LoginCredentials) => {
    loginMutation.mutate(credentials);
  };
  
  return (
    <form onSubmit={handleSubmit}>
      {/* form fields */}
      <button 
        type="submit" 
        disabled={loginMutation.isPending}
      >
        {loginMutation.isPending ? 'Logging in...' : 'Login'}
      </button>
    </form>
  );
}
```

### Complex Query with Pagination

```typescript
function StoriesList() {
  const [page, setPage] = createSignal(1);
  
  const { data: stories, isLoading } = useApiQuery(
    queryKeys.storiesPaginated(page(), 10),
    'stories',
    { page: page(), limit: 10 }
  );
  
  return (
    <div>
      {stories?.map(story => (
        <StoryCard key={story.id} story={story} />
      ))}
      <button onClick={() => setPage(p => p + 1)}>
        Load More
      </button>
    </div>
  );
}
```

## 🎯 Key Benefits

### 1. **Minimal Boilerplate**
- No need to write repetitive axios calls
- Automatic error handling and loading states
- Built-in cache invalidation patterns

### 2. **Type Safety**
- Full TypeScript support with generics
- Compile-time error checking
- IntelliSense autocomplete for API responses

### 3. **Developer Experience**
- Automatic toast notifications
- Consistent error handling
- Easy cache management
- Hot module replacement support

### 4. **Performance**
- Efficient caching with TanStack Query
- Automatic background refetching
- Request deduplication
- Optimistic updates support

### 5. **Maintainability**
- Centralized API configuration
- Consistent patterns across all endpoints
- Easy to add new endpoints
- Clear separation of concerns

## 🔄 Migration from Direct Axios

Before (old pattern):
```typescript
// Lots of boilerplate and manual error handling
const [user, setUser] = createSignal<User | null>(null);
const [loading, setLoading] = createSignal(false);
const [error, setError] = createSignal<string | null>(null);

const fetchUser = async () => {
  setLoading(true);
  setError(null);
  try {
    const response = await axios.get<User>('/users/profile');
    setUser(response.data);
  } catch (err) {
    setError(err instanceof Error ? err.message : 'Unknown error');
  } finally {
    setLoading(false);
  }
};
```

After (new pattern):
```typescript
// Clean, minimal, and type-safe
const { data: user, isLoading, error } = useUserProfile();
```

## 📈 Scalability

The architecture easily scales by:

1. **Adding new endpoints**: Create new files in `endpoints/` following the same pattern
2. **Extending query keys**: Add new families to `queryKeys.ts`
3. **Customizing behavior**: Override options in specific hooks
4. **Adding middleware**: Extend the axios client with interceptors

## 🛠️ Development Tools Integration

The API layer integrates seamlessly with development tools:

- **TypeScript**: Full compile-time type checking
- **ESLint**: Consistent code style enforcement  
- **Vitest**: Comprehensive test coverage with mocked API calls
- **TanStack Query DevTools**: Real-time cache inspection and debugging

This implementation provides a robust, scalable, and developer-friendly foundation for all API interactions in your SolidJS application.