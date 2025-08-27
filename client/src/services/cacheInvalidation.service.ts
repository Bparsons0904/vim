import { useQueryClient } from '@tanstack/solid-query';
import { createEffect, onCleanup } from 'solid-js';
import { useWebSocket } from '../context/WebSocketContext';

export function useCacheInvalidation() {
  const queryClient = useQueryClient();
  const webSocket = useWebSocket();

  createEffect(() => {
    const cleanup = webSocket.onCacheInvalidation((resourceType: string, resourceId: string) => {
      switch (resourceType) {
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
          
        default:
          // Unknown resource type - no action needed
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