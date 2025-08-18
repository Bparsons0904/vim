// User-specific API functions and hooks using the enhanced API layer

import { User } from "src/types/User";
import { useApiGet, useApiPost, useApiPut, useApiDelete } from "../queryHooks";
import { queryKeys } from "../queryKeys";
import { Accessor } from "solid-js";

// Type definitions for user operations
export interface LoginCredentials {
  login: string;
  password: string;
}

export interface RegisterData {
  firstName: string;
  lastName: string;
  email: string;
  username: string;
  password: string;
}

export interface UpdateUserData {
  firstName?: string;
  lastName?: string;
  email?: string;
}

export interface ChangePasswordData {
  currentPassword: string;
  newPassword: string;
  confirmPassword: string;
}

// Query hooks for user data
export function useCurrentUser(enabled?: boolean | Accessor<boolean>) {
  return useApiGet<{ user: User }>(
    queryKeys.userProfile(),
    'users',
    undefined,
    {
      enabled,
      staleTime: 10 * 60 * 1000, // 10 minutes for user data
    }
  );
}

export function useUser(userId: string, enabled?: boolean | Accessor<boolean>) {
  return useApiGet<User>(
    queryKeys.user(userId),
    `users/${userId}`,
    undefined,
    {
      enabled,
    }
  );
}

export function useUsersList(params?: { page?: number; limit?: number; search?: string }) {
  return useApiGet<{ users: User[]; total: number }>(
    queryKeys.filtered(queryKeys.users(), params || {}),
    'users/list',
    params
  );
}

// Mutation hooks for user operations
export function useLoginUser() {
  return useApiPost<User, LoginCredentials>('users/login', {
    invalidateQueries: [queryKeys.userProfile()],
    onSuccessToast: 'Successfully logged in!',
    onErrorToast: (error) => `Login failed: ${error.message}`,
  });
}

export function useRegisterUser() {
  return useApiPost<User, RegisterData>('users/register', {
    invalidateQueries: [queryKeys.userProfile()],
    onSuccessToast: 'Account created successfully!',
    onErrorToast: (error) => `Registration failed: ${error.message}`,
  });
}

export function useLogoutUser() {
  return useApiPost<void, Record<string, never>>('users/logout', {
    invalidateQueries: [queryKeys.all()],
    onSuccessToast: 'Successfully logged out!',
  });
}

export function useUpdateUser() {
  return useApiPut<User, UpdateUserData>('users/profile', {
    invalidateQueries: [queryKeys.userProfile(), queryKeys.users()],
    onSuccessToast: 'Profile updated successfully!',
    onErrorToast: (error) => `Failed to update profile: ${error.message}`,
  });
}

export function useChangePassword() {
  return useApiPost<void, ChangePasswordData>('users/change-password', {
    onSuccessToast: 'Password changed successfully!',
    onErrorToast: (error) => `Failed to change password: ${error.message}`,
  });
}

export function useDeleteUser() {
  return useApiDelete<void>('users/profile', {
    invalidateQueries: [queryKeys.all()],
    onSuccessToast: 'Account deleted successfully!',
    onErrorToast: (error) => `Failed to delete account: ${error.message}`,
  });
}

// Higher-level hooks that combine multiple operations
export function useAuthenticatedUser() {
  const userQuery = useCurrentUser();
  
  return {
    ...userQuery,
    isAuthenticated: userQuery.isSuccess && !!userQuery.data?.user,
    user: userQuery.data?.user || null,
  };
}

// Search hook with debouncing
export function useUserSearch(searchQuery: Accessor<string>) {
  return useApiGet<{ users: User[]; total: number }>(
    queryKeys.search(queryKeys.users(), searchQuery()),
    'users/search',
    { q: searchQuery() },
    {
      enabled: () => searchQuery().length > 2,
    }
  );
}