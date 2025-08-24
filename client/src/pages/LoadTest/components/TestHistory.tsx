import { Component, Show, For } from "solid-js";
import { Card } from "@components/common/ui/Card/Card";
import { Button } from "@components/common/ui/Button/Button";
import { LoadTestResult } from "../LoadTest";
import styles from "./TestHistory.module.scss";

interface TestHistoryProps {
  tests: LoadTestResult[];
  isLoading?: boolean;
  error?: any;
}

export const TestHistory: Component<TestHistoryProps> = (props) => {
  const formatTime = (ms: number | undefined): string => {
    if (!ms) return 'N/A';
    if (ms < 1000) return `${ms}ms`;
    return `${(ms / 1000).toFixed(2)}s`;
  };

  const formatRate = (rows: number, timeMs: number | undefined): string => {
    if (!timeMs || timeMs === 0) return 'N/A';
    const rowsPerSecond = Math.round((rows / timeMs) * 1000);
    return `${rowsPerSecond.toLocaleString()} rows/sec`;
  };

  const getStatusColor = (status: string): string => {
    switch (status) {
      case 'completed':
        return 'var(--color-success)';
      case 'running':
        return 'var(--color-primary)';
      case 'failed':
        return 'var(--color-error)';
      default:
        return 'var(--color-text-secondary)';
    }
  };

  const getStatusIcon = (status: string): string => {
    switch (status) {
      case 'completed':
        return '✓';
      case 'running':
        return '⟳';
      case 'failed':
        return '✗';
      default:
        return '?';
    }
  };

  const getMethodDisplayName = (method: string): string => {
    switch (method) {
      case 'brute_force':
        return 'Brute Force';
      case 'batched':
        return 'Batched';
      case 'optimized':
        return 'Optimized';
      case 'ludicrous':
        return 'Ludicrous Speed';
      default:
        return method;
    }
  };

  // Take only the last 10 tests
  const recentTests = () => props.tests.slice(0, 10);

  return (
    <Card class={styles.testHistoryCard}>
      <div class={styles.header}>
        <h2>Recent Test History</h2>
        <Show when={props.tests.length > 0}>
          <Button variant="secondary" size="sm">
            Export Results
          </Button>
        </Show>
      </div>
      
      <Show when={props.error}>
        <div class={styles.errorState}>
          <p>Failed to load test history</p>
          <p class={styles.errorSubtext}>
            Please check your connection and try again
          </p>
        </div>
      </Show>

      <Show when={props.isLoading}>
        <div class={styles.loadingState}>
          <p>Loading test history...</p>
        </div>
      </Show>

      <Show 
        when={!props.isLoading && !props.error && props.tests.length > 0}
        fallback={
          <Show when={!props.isLoading && !props.error}>
            <div class={styles.emptyState}>
              <p>No tests completed yet</p>
              <p class={styles.emptySubtext}>Start your first performance test to see results here</p>
            </div>
          </Show>
        }
      >
        <div class={styles.historyList}>
          <For each={recentTests()}>
            {(test, index) => (
              <div class={styles.historyItem}>
                <div class={styles.itemHeader}>
                  <div class={styles.itemTitle}>
                    <span 
                      class={styles.statusIcon}
                      style={{ color: getStatusColor(test.status) }}
                    >
                      {getStatusIcon(test.status)}
                    </span>
                    Test #{index() + 1}
                  </div>
                  <div class={styles.itemMethod}>
                    {getMethodDisplayName(test.method)}
                  </div>
                </div>
                
                <div class={styles.itemDetails}>
                  <span class={styles.detail}>{test.rows.toLocaleString()} rows</span>
                  <span class={styles.detail}>{test.columns} columns</span>
                  <Show when={test.status === 'completed' && test.totalTime}>
                    <span class={styles.detail}>{formatTime(test.totalTime)}</span>
                    <span class={styles.detail}>{formatRate(test.rows, test.totalTime)}</span>
                  </Show>
                </div>
                
                <Show when={test.status === 'failed' && test.errorMessage}>
                  <div class={styles.errorPreview}>
                    Error: {test.errorMessage!.substring(0, 80)}
                    {test.errorMessage!.length > 80 && '...'}
                  </div>
                </Show>
              </div>
            )}
          </For>
          
          <Show when={props.tests.length > 10}>
            <div class={styles.showMore}>
              <Button variant="ghost" size="sm">
                Show {props.tests.length - 10} more tests
              </Button>
            </div>
          </Show>
        </div>
      </Show>
    </Card>
  );
};