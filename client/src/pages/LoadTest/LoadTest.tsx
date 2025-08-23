import { Component, createSignal, createEffect, onCleanup } from "solid-js";
import styles from "./LoadTest.module.scss";
import { LoadTestForm } from "./LoadTestForm";
import { ProgressDisplay } from "./ProgressDisplay";
import { ResultsDashboard } from "./ResultsDashboard";
import { useStartLoadTest, useLoadTestHistory } from "@services/api/hooks/loadtest.hooks";

export interface LoadTestConfig {
  rows: number;
  columns: number;
  dateColumns: number;
  method: 'brute_force' | 'optimized';
}

export interface LoadTestResult {
  id: string;
  rows: number;
  columns: number;
  dateColumns: number;
  method: string;
  status: 'running' | 'completed' | 'failed';
  csvGenTime?: number;
  parseTime?: number;
  insertTime?: number;
  totalTime?: number;
  errorMessage?: string;
}

const LoadTest: Component = () => {
  const [currentTest, setCurrentTest] = createSignal<LoadTestResult | null>(null);
  
  // API hooks
  const startTestMutation = useStartLoadTest();
  const historyQuery = useLoadTestHistory({ limit: 20 });

  const handleStartTest = async (config: LoadTestConfig) => {
    try {
      const result = await startTestMutation.mutateAsync(config);
      setCurrentTest(result.data.test);
    } catch (error) {
      console.error("Failed to start test:", error);
      // Error handling is done in the mutation hook
    }
  };

  const handleTestComplete = (result: LoadTestResult) => {
    setCurrentTest(result);
    // History will be automatically updated via query invalidation
  };

  return (
    <div class={styles.loadTest}>
      <div class={styles.container}>
        <header class={styles.header}>
          <h1>Database Load Performance Tester</h1>
          <p>Test and benchmark database insertion performance with configurable parameters</p>
        </header>

        <div class={styles.content}>
          <div class={styles.leftColumn}>
            <LoadTestForm 
              onStartTest={handleStartTest}
              isLoading={startTestMutation.isPending}
              disabled={currentTest()?.status === 'running'}
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
              testHistory={historyQuery.data?.data?.tests || []}
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