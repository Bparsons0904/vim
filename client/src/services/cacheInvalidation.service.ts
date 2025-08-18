import { useQueryClient } from '@tanstack/solid-query';
import { createEffect, onCleanup } from 'solid-js';
import { useWebSocket } from '../context/WebSocketContext';

export function useCacheInvalidation() {
  const queryClient = useQueryClient();
  const webSocket = useWebSocket();

  createEffect(() => {
    const cleanup = webSocket.onCacheInvalidation((resourceType: string, resourceId: string) => {
      console.log(`[CacheInvalidation] Invalidating ${resourceType}:${resourceId}`);
      
      switch (resourceType) {
        case 'story':
          // Invalidate specific story
          queryClient.invalidateQueries({
            queryKey: ['stories', resourceId]
          });
          
          // Invalidate user's stories list
          queryClient.invalidateQueries({
            queryKey: ['user', 'stories']
          });
          
          // Invalidate any story-related queries
          queryClient.invalidateQueries({
            predicate: (query) => {
              const key = query.queryKey as string[];
              return key.includes('stories') || (key.includes('story') && key.includes(resourceId));
            }
          });
          break;
          
        case 'user':
          // Invalidate user-specific queries
          queryClient.invalidateQueries({
            queryKey: ['user', resourceId]
          });
          
          // Invalidate current user queries if it's the current user
          queryClient.invalidateQueries({
            queryKey: ['user']
          });
          break;
          
        case 'section':
          // For sections, we might need to invalidate the parent story
          // This would need the story ID in the data, or we could invalidate all stories
          queryClient.invalidateQueries({
            predicate: (query) => {
              const key = query.queryKey as string[];
              return key.includes('stories') || key.includes('story');
            }
          });
          break;
          
        default:
          console.warn(`[CacheInvalidation] Unknown resource type: ${resourceType}`);
          break;
      }
    });

    onCleanup(() => {
      cleanup();
    });
  });
}

// Convenience hook for components that need cache invalidation
export function useAutoCacheInvalidation() {
  useCacheInvalidation();
}