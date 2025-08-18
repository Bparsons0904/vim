// Section-specific API functions using the correct server types and endpoints

import { useApiPost, useApiPatch, useApiDelete } from "../queryHooks";
import { queryKeys } from "../queryKeys";
import { useQueryClient } from '@tanstack/solid-query';
import { ContentRequest, StoryApiResponse } from '../apiTypes';

// Using shared types - all section operations return a story response
// and create/update operations take content as input

// Mutation hook for creating a new section
export function useCreateSection(storyId: string) {
  const queryClient = useQueryClient();
  
  return useApiPost<StoryApiResponse, ContentRequest>(
    `stories/${storyId}/sections`,
    {
      onSuccessToast: "Section added successfully!",
      onErrorToast: "Failed to add section",
      onSuccess: (data) => {
        // Optimistic update with the returned story
        queryClient.setQueryData(queryKeys.story(storyId), {
          message: "success",
          story: data.story
        });
      },
    },
  );
}

// Mutation hook for updating a section
export function useUpdateSection(storyId: string, sectionId: string) {
  const queryClient = useQueryClient();
  
  return useApiPatch<StoryApiResponse, ContentRequest>(
    `stories/${storyId}/sections/${sectionId}`,
    {
      onSuccessToast: "Section updated successfully!",
      onErrorToast: "Failed to update section",
      onSuccess: (data) => {
        // Optimistic update with the returned story
        queryClient.setQueryData(queryKeys.story(storyId), {
          message: "success",
          story: data.story
        });
      },
    },
  );
}

// Mutation hook for deleting a section
export function useDeleteSection(storyId: string, sectionId: string) {
  const queryClient = useQueryClient();
  
  return useApiDelete<StoryApiResponse>(
    `stories/${storyId}/sections/${sectionId}`,
    {
      onSuccessToast: "Section deleted successfully!",
      onErrorToast: "Failed to delete section",
      onSuccess: (data) => {
        // Optimistic update with the returned story
        queryClient.setQueryData(queryKeys.story(storyId), {
          message: "success",
          story: data.story
        });
      },
    },
  );
}
