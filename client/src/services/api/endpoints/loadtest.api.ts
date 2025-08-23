import { getApi, postApi, deleteApi } from "../api.service";
import { LoadTestConfig, LoadTestResult } from "@pages/LoadTest/LoadTest";

export interface StartLoadTestRequest {
  rows: number;
  columns: number;
  dateColumns: number;
  method: 'brute_force' | 'optimized';
}

export interface StartLoadTestResponse {
  success: boolean;
  data: {
    testId: string;
    test: LoadTestResult;
  };
}

export interface GetLoadTestResponse {
  success: boolean;
  data: {
    test: LoadTestResult;
  };
}

export interface GetLoadTestHistoryResponse {
  success: boolean;
  data: {
    tests: LoadTestResult[];
    pagination?: {
      page: number;
      limit: number;
      total: number;
      hasMore: boolean;
    };
  };
}

export interface DeleteLoadTestResponse {
  success: boolean;
  data: {
    message: string;
  };
}

// Start a new load test
export const startLoadTest = async (config: LoadTestConfig): Promise<StartLoadTestResponse> => {
  return postApi<StartLoadTestResponse, StartLoadTestRequest>('loadtest/start', {
    rows: config.rows,
    columns: config.columns,
    dateColumns: config.dateColumns,
    method: config.method,
  });
};

// Get status of a specific load test
export const getLoadTestStatus = async (testId: string): Promise<GetLoadTestResponse> => {
  return getApi<GetLoadTestResponse>(`loadtest/${testId}`);
};

// Get load test history
export const getLoadTestHistory = async (params?: {
  page?: number;
  limit?: number;
  status?: 'running' | 'completed' | 'failed';
  method?: 'brute_force' | 'optimized';
}): Promise<GetLoadTestHistoryResponse> => {
  return getApi<GetLoadTestHistoryResponse>('loadtest', params);
};

// Delete a load test and its data
export const deleteLoadTest = async (testId: string): Promise<DeleteLoadTestResponse> => {
  return deleteApi<DeleteLoadTestResponse>(`loadtest/${testId}`);
};

// WebSocket event types for real-time updates
export interface LoadTestProgressEvent {
  type: 'progress';
  testId: string;
  phase: 'csv_generation' | 'parsing' | 'insertion';
  overallProgress: number;
  phaseProgress: number;
  currentPhase: string;
  rowsProcessed: number;
  rowsPerSecond: number;
  eta: string;
  message: string;
}

export interface LoadTestCompleteEvent {
  type: 'complete';
  testId: string;
  test: LoadTestResult;
}

export interface LoadTestErrorEvent {
  type: 'error';
  testId: string;
  error: string;
}

export type LoadTestWebSocketEvent = 
  | LoadTestProgressEvent 
  | LoadTestCompleteEvent 
  | LoadTestErrorEvent;