// Story-specific API functions using the correct server types and endpoints

import { useApiGet } from "../queryHooks";
import { queryKeys } from "../queryKeys";
import {
  StoriesApiResponse,
  StoryApiResponse,
  CreateFormDataApiResponse,
  BaseApiResponse,
  KnownContactsApiResponse,
} from "../apiTypes";
import { useMutation, useQueryClient } from "@tanstack/solid-query";
import { postApi, deleteApi } from "../api.service";
import { User } from "../../../types/User";

// Using shared types for consistent API responses

// Query hook for getting user's stories
export function useUserStories() {
  return useApiGet<StoriesApiResponse>(queryKeys.stories(), "stories");
}

// Query hook for getting individual story by ID
export function useStory(storyId: string) {
  return useApiGet<StoryApiResponse>(
    queryKeys.story(storyId),
    `stories/${storyId}`,
  );
}

// Query hook for getting create form data (genres and order types)
export function useCreateFormData() {
  return useApiGet<CreateFormDataApiResponse>(
    queryKeys.createFormData(),
    "stories/createFormData",
  );
}

// Request type matching server CreateStoryRequest (without inviteEmails)
export interface CreateStoryRequest {
  title: string;
  description: string;
  genreId: number;
  storyOrderTypeId: number;
  allowOthersToInvite: boolean;
}

// Mutation hook for creating stories
export function useCreateStory() {
  const queryClient = useQueryClient();

  return useMutation(() => ({
    mutationFn: (data: CreateStoryRequest) =>
      postApi<StoryApiResponse, CreateStoryRequest>("stories", data),
    onSuccess: () => {
      // Invalidate stories cache to refetch user's stories
      queryClient.invalidateQueries({ queryKey: queryKeys.stories() });
    },
  }));
}

// Invitation-related types and API methods

export interface CreateInvitesRequest {
  storyId: string;
  emails?: string[];
  userIds?: string[];
}

export interface InviteUsersRequest {
  storyId: string;
  userIds: string[];
}

export interface StoryMembersApiResponse extends BaseApiResponse {
  members: User[];
}

// Query hook for getting story members
export function useStoryMembers(storyId: string) {
  return useApiGet<StoryMembersApiResponse>(
    queryKeys.storyMembers(storyId),
    `stories/${storyId}/members`,
  );
}

// Mutation hook for creating invitations
export function useCreateInvites() {
  const queryClient = useQueryClient();

  return useMutation(() => ({
    mutationFn: ({ storyId, emails, userIds }: { storyId: string; emails?: string[]; userIds?: string[] }) =>
      postApi<BaseApiResponse, CreateInvitesRequest>(
        "invites/create",
        { storyId, emails, userIds },
      ),
    onSuccess: () => {
      // Invalidate stories cache to refetch user's invitations
      queryClient.invalidateQueries({ queryKey: queryKeys.stories() });
    },
  }));
}

// Mutation hook for inviting users by ID
export function useInviteUsers() {
  const queryClient = useQueryClient();

  return useMutation(() => ({
    mutationFn: ({ storyId, userIds }: { storyId: string; userIds: string[] }) =>
      postApi<BaseApiResponse, InviteUsersRequest>(
        "invites/inviteUsers",
        { storyId, userIds },
      ),
    onSuccess: () => {
      // Invalidate stories cache to refetch user's invitations
      queryClient.invalidateQueries({ queryKey: queryKeys.stories() });
    },
  }));
}

// Mutation hook for accepting invitations
export function useAcceptInvite() {
  const queryClient = useQueryClient();

  return useMutation(() => ({
    mutationFn: (inviteId: string) =>
      postApi<BaseApiResponse>(`invites/${inviteId}/accept`, {}),
    onSuccess: () => {
      // Invalidate stories cache to refetch user's stories and invitations
      queryClient.invalidateQueries({ queryKey: queryKeys.stories() });
    },
  }));
}

// Mutation hook for declining invitations
export function useDeclineInvite() {
  const queryClient = useQueryClient();

  return useMutation(() => ({
    mutationFn: (inviteId: string) =>
      postApi<BaseApiResponse>(`invites/${inviteId}/decline`, {}),
    onSuccess: () => {
      // Invalidate stories cache to refetch user's invitations
      queryClient.invalidateQueries({ queryKey: queryKeys.stories() });
    },
  }));
}

// Mutation hook for revoking invitations
export function useRevokeInvite() {
  const queryClient = useQueryClient();

  return useMutation(() => ({
    mutationFn: (inviteId: string) =>
      deleteApi<BaseApiResponse>(`invites/${inviteId}`),
    onSuccess: () => {
      // Invalidate stories cache to refetch user's invitations
      queryClient.invalidateQueries({ queryKey: queryKeys.stories() });
    },
  }));
}

// Mutation hook for removing story members
export function useRemoveStoryMember() {
  const queryClient = useQueryClient();

  return useMutation(() => ({
    mutationFn: ({ storyId, userId }: { storyId: string; userId: string }) =>
      deleteApi<BaseApiResponse>(`stories/${storyId}/members/${userId}`),
    onSuccess: (_, { storyId }) => {
      // Invalidate story members cache
      queryClient.invalidateQueries({
        queryKey: queryKeys.storyMembers(storyId),
      });
      // Also invalidate stories cache in case membership affects story access
      queryClient.invalidateQueries({ queryKey: queryKeys.stories() });
    },
  }));
}

// Query hook for getting known contacts (users previously collaborated with)
export function useKnownContacts() {
  return useApiGet<KnownContactsApiResponse>(
    queryKeys.knownContacts(),
    "invites/knownContacts",
  );
}
