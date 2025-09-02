import { useQuery, useMutation, useQueryClient } from "@tanstack/solid-query";
import { queryKeys, invalidationHelpers } from "../queryKeys";
import { 
  startLoadTest, 
  getLoadTestStatus, 
  getLoadTestHistory, 
  deleteLoadTest,
  getPerformanceSummary,
  getOverallSummary,
  StartLoadTestResponse,
  GetLoadTestResponse,
  DeleteLoadTestResponse
} from "../endpoints/loadtest.api";
import { LoadTestConfig, LoadTestResult } from "@pages/LoadTest/LoadTest";
import { ApiClientError, NetworkError } from "../apiTypes";
import { useToast } from "@context/ToastContext";

// Query hook for getting load test status
export const useLoadTestStatus = (testId: string, enabled = true) => {
  return useQuery(() => ({
    queryKey: queryKeys.loadTest(testId),
    queryFn: () => getLoadTestStatus(testId),
    enabled,
    refetchInterval: 2000, // Poll every 2 seconds
    refetchIntervalInBackground: true,
  }));
};

// Query hook for getting load test history
export const useLoadTestHistory = (filters?: {
  page?: number;
  limit?: number;
  status?: 'running' | 'completed' | 'failed';
  method?: 'brute_force' | 'batched' | 'optimized' | 'ludicrous';
}) => {
  return useQuery(() => ({
    queryKey: queryKeys.loadTestHistory(filters),
    queryFn: () => getLoadTestHistory(filters),
    staleTime: 5 * 60 * 1000, // 5 minutes
  }));
};

// Query hook for getting performance summary
export const usePerformanceSummary = () => {
  return useQuery(() => ({
    queryKey: ['performance-summary'],
    queryFn: () => getPerformanceSummary(),
    staleTime: 10 * 60 * 1000, // 10 minutes - summary data changes less frequently
  }));
};

// Query hook for getting overall summary statistics
export const useOverallSummary = () => {
  return useQuery(() => ({
    queryKey: ['overall-summary'],
    queryFn: () => getOverallSummary(),
    staleTime: 10 * 60 * 1000, // 10 minutes - summary data changes less frequently
  }));
};

// Mutation hook for starting a load test
export const useStartLoadTest = () => {
  const queryClient = useQueryClient();
  const toast = useToast();

  return useMutation(() => ({
    mutationFn: (config: LoadTestConfig) => startLoadTest(config),
    onSuccess: (data: StartLoadTestResponse) => {
      toast.showSuccess('Load test started successfully');

      // Invalidate and refetch history to include new test
      queryClient.invalidateQueries({ 
        queryKey: invalidationHelpers.invalidateLoadTests() 
      });

      // Set the new test data in cache
      queryClient.setQueryData(
        queryKeys.loadTest(data.loadTest.id),
        data
      );
    },
    onError: (error: ApiClientError | NetworkError) => {
      const message = error instanceof ApiClientError 
        ? error.message 
        : 'Failed to start load test. Please check your connection.';
      
      toast.showError(message);
    },
  }));
};

// Mutation hook for deleting a load test
export const useDeleteLoadTest = () => {
  const queryClient = useQueryClient();
  const toast = useToast();

  return useMutation(() => ({
    mutationFn: (testId: string) => deleteLoadTest(testId),
    onSuccess: (data: DeleteLoadTestResponse, testId: string) => {
      toast.showSuccess('Load test deleted successfully');

      // Remove the specific test from cache
      queryClient.removeQueries({ 
        queryKey: queryKeys.loadTest(testId) 
      });

      // Invalidate history to reflect deletion
      queryClient.invalidateQueries({ 
        queryKey: invalidationHelpers.invalidateLoadTests() 
      });
    },
    onError: (error: ApiClientError | NetworkError) => {
      const message = error instanceof ApiClientError 
        ? error.message 
        : 'Failed to delete load test. Please try again.';
      
      toast.showError(message);
    },
  }));
};

// Hook for managing WebSocket updates to load test data
export const useLoadTestWebSocketUpdates = () => {
  const queryClient = useQueryClient();

  const updateTestProgress = (testId: string, updatedTest: Partial<LoadTestResult>) => {
    queryClient.setQueryData(
      queryKeys.loadTest(testId),
      (oldData: GetLoadTestResponse | undefined) => {
        if (!oldData) return oldData;
        
        return {
          ...oldData,
          loadTest: {
            ...oldData.loadTest,
            ...updatedTest,
          },
        };
      }
    );
  };

  const markTestComplete = (testId: string, completedTest: LoadTestResult) => {
    // Update the specific test query
    queryClient.setQueryData(
      queryKeys.loadTest(testId),
      {
        message: "success",
        loadTest: completedTest,
      }
    );

    // Invalidate history to include completed test with final metrics
    queryClient.invalidateQueries({ 
      queryKey: invalidationHelpers.invalidateLoadTests() 
    });

    // Invalidate performance summary since new completed test affects statistics
    queryClient.invalidateQueries({ 
      queryKey: ['performance-summary'] 
    });

    // Invalidate overall summary since new completed test affects statistics
    queryClient.invalidateQueries({ 
      queryKey: ['overall-summary'] 
    });
  };

  return {
    updateTestProgress,
    markTestComplete,
  };
};

// Utility hook for load test statistics
export const useLoadTestStats = () => {
  const historyQuery = useLoadTestHistory();
  
  const stats = () => {
    const tests = historyQuery.data?.loadTests || [];
    const completedTests = tests.filter(test => test.status === 'completed');
    const runningTests = tests.filter(test => test.status === 'running');
    const failedTests = tests.filter(test => test.status === 'failed');

    const bestPerformance = completedTests
      .filter(test => test.totalTime)
      .reduce((best, test) => {
        if (!best.totalTime || !test.totalTime) return test.totalTime ? test : best;
        return test.totalTime < best.totalTime ? test : best;
      }, {} as LoadTestResult);

    const averageTime = completedTests
      .filter(test => test.totalTime)
      .reduce((sum, test) => sum + (test.totalTime || 0), 0) / 
      (completedTests.filter(test => test.totalTime).length || 1);

    return {
      total: tests.length,
      completed: completedTests.length,
      running: runningTests.length,
      failed: failedTests.length,
      bestPerformance: Object.keys(bestPerformance).length > 0 ? bestPerformance : null,
      averageTime: averageTime > 0 ? averageTime : null,
    };
  };

  return {
    stats,
    isLoading: historyQuery.isLoading,
    error: historyQuery.error,
    refetch: historyQuery.refetch,
  };
};