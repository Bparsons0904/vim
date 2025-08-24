import { getApi, postApi, deleteApi } from "../api.service";
import { LoadTestConfig, LoadTestResult } from "@pages/LoadTest/LoadTest";

export interface StartLoadTestRequest {
  rows: number;
  method: 'brute_force' | 'batched' | 'optimized' | 'ludicrous';
}

export interface StartLoadTestResponse {
  message: string;
  loadTest: LoadTestResult;
}

export interface GetLoadTestResponse {
  message: string;
  loadTest: LoadTestResult;
}

export interface GetLoadTestHistoryResponse {
  message: string;
  loadTests: LoadTestResult[];
}

export interface DeleteLoadTestResponse {
  message: string;
}

// Start a new load test
export const startLoadTest = async (config: LoadTestConfig): Promise<StartLoadTestResponse> => {
  return postApi<StartLoadTestResponse, StartLoadTestRequest>('load-tests', {
    rows: config.rows,
    method: config.method,
  });
};

// Get status of a specific load test
export const getLoadTestStatus = async (testId: string): Promise<GetLoadTestResponse> => {
  return getApi<GetLoadTestResponse>(`load-tests/${testId}`);
};

// Get load test history
export const getLoadTestHistory = async (params?: {
  page?: number;
  limit?: number;
  status?: 'running' | 'completed' | 'failed';
  method?: 'brute_force' | 'batched' | 'optimized' | 'ludicrous';
}): Promise<GetLoadTestHistoryResponse> => {
  return getApi<GetLoadTestHistoryResponse>('load-tests', params);
};

// Delete a load test and its data
export const deleteLoadTest = async (testId: string): Promise<DeleteLoadTestResponse> => {
  return deleteApi<DeleteLoadTestResponse>(`load-tests/${testId}`);
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