import { Component, createSignal } from "solid-js";
import styles from "./LoadTest.module.scss";
import { LoadTestForm } from "./LoadTestForm";
import { ProgressDisplay } from "./ProgressDisplay";
import { ResultsDashboard } from "./ResultsDashboard";
import { ProcessInfo } from "./ProcessInfo";
import {
  useStartLoadTest,
  useLoadTestHistory,
} from "@services/api/hooks/loadtest.hooks";
import { useQueryClient } from "@tanstack/solid-query";
import { queryKeys } from "@services/api/queryKeys";

export interface LoadTestConfig {
  rows: number;
  method: "brute_force" | "batched" | "optimized" | "ludicrous" | "plaid";
}

export interface LoadTestResult {
  id: string;
  rows: number;
  columns: number;
  dateColumns: number;
  method: string;
  status: "running" | "completed" | "failed";
  csvGenTime?: number;
  parseTime?: number;
  insertTime?: number;
  totalTime?: number;
  errorMessage?: string;
  createdAt: string;
}

const LoadTest: Component = () => {
  const [currentTest, setCurrentTest] = createSignal<LoadTestResult | null>(
    null,
  );

  // API hooks
  const startTestMutation = useStartLoadTest();
  const historyQuery = useLoadTestHistory();
  const queryClient = useQueryClient();

  const handleStartTest = async (config: LoadTestConfig) => {
    try {
      const result = await startTestMutation.mutateAsync(config);
      setCurrentTest(result.loadTest);
    } catch {
      // Error handling is done in the mutation hook
    }
  };

  const handleTestComplete = (result: LoadTestResult) => {
    setCurrentTest(result);

    // Invalidate queries to refresh data
    queryClient.invalidateQueries({
      queryKey: queryKeys.loadTestHistory(),
    });
    queryClient.invalidateQueries({
      queryKey: ["performance-summary"],
    });
  };

  return (
    <div class={styles.loadTest}>
      <div class={styles.container}>
        <header class={styles.header}>
          <h1>Database Load Performance Tester</h1>
          <p>
            Test and benchmark database insertion performance with various
            implementation methods
          </p>
        </header>

        <ProcessInfo />

        <div class={styles.content}>
          <div class={styles.leftColumn}>
            <LoadTestForm
              onStartTest={handleStartTest}
              isLoading={startTestMutation.isPending}
              disabled={currentTest()?.status === "running"}
            />

            {currentTest() && (
              <ProgressDisplay
                test={currentTest()!}
                onTestComplete={handleTestComplete}
              />
            )}
          </div>

          <div class={styles.rightColumn}>
            <ResultsDashboard
              currentTest={currentTest()}
              testHistory={historyQuery.data?.loadTests || []}
              isHistoryLoading={historyQuery.isLoading}
              historyError={historyQuery.error}
            />
          </div>
        </div>
      </div>
    </div>
  );
};

export default LoadTest;

