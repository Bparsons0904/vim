import { Component, Show, For } from "solid-js";
import { Card } from "@components/common/ui/Card/Card";
import { PerformanceSummary as PerformanceSummaryData } from "@services/api/endpoints/loadtest.api";
import styles from "./PerformanceSummary.module.scss";

interface PerformanceSummaryProps {
  summaries: PerformanceSummaryData[];
  isLoading?: boolean;
  error?: any;
}

export const PerformanceSummary: Component<PerformanceSummaryProps> = (props) => {
  const formatNumber = (num: number): string => {
    return num.toLocaleString();
  };

  const formatTime = (ms: number): string => {
    if (ms < 1000) return `${ms}ms`;
    return `${(ms / 1000).toFixed(2)}s`;
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
      case 'plaid':
        return 'Plaid';
      default:
        return method;
    }
  };

  const getMethodColor = (method: string): string => {
    switch (method) {
      case 'brute_force':
        return '#f39c12'; // Orange/yellow for warning
      case 'batched':
        return '#3498db'; // Blue for info
      case 'optimized':
        return '#27ae60'; // Green for success
      case 'ludicrous':
        return '#e74c3c'; // Red for primary/intense
      case 'plaid':
        return '#6c5ce7'; // Purple for Plaid
      default:
        return '#95a5a6'; // Gray for unknown
    }
  };

  return (
    <Card class={styles.performanceSummaryCard}>
      <div class={styles.header}>
        <h2>Performance Summary</h2>
        <p class={styles.subtitle}>Performance metrics grouped by test method</p>
      </div>
      
      <Show when={props.error}>
        <div class={styles.errorState}>
          <p>Failed to load performance summary</p>
          <p class={styles.errorSubtext}>
            Please check your connection and try again
          </p>
        </div>
      </Show>

      <Show when={props.isLoading}>
        <div class={styles.loadingState}>
          <p>Loading performance summary...</p>
        </div>
      </Show>

      <Show 
        when={!props.isLoading && !props.error && props.summaries.length > 0}
        fallback={
          <Show when={!props.isLoading && !props.error}>
            <div class={styles.emptyState}>
              <p>No completed tests yet</p>
              <p class={styles.emptySubtext}>Run some performance tests to see summary statistics</p>
            </div>
          </Show>
        }
      >
        <div class={styles.summariesGrid}>
          <For each={props.summaries.slice().sort((a, b) => b.avgRowsPerSec - a.avgRowsPerSec)}>
            {(summary) => (
              <div class={styles.summaryCard}>
                <div class={styles.summaryHeader}>
                  <div class={styles.methodName}>
                    <span 
                      class={styles.methodIcon}
                      style={{ 'background-color': getMethodColor(summary.method) }}
                    />
                    {getMethodDisplayName(summary.method)}
                  </div>
                  <div class={styles.testCount}>
                    {summary.testCount} test{summary.testCount !== 1 ? 's' : ''}
                  </div>
                </div>

                <div class={styles.metricsLayout}>
                  <div class={styles.primaryMetric}>
                    <span class={styles.primaryLabel}>Average Throughput</span>
                    <div class={styles.primaryValueContainer}>
                      <span class={styles.primaryValue}>
                        {formatNumber(summary.avgRowsPerSec)}
                      </span>
                      <span class={styles.primaryUnit}>rows/sec</span>
                    </div>
                  </div>

                  <div class={styles.secondaryMetrics}>
                    <div class={styles.metric}>
                      <span class={styles.metricLabel}>Max</span>
                      <span class={styles.metricValue}>
                        {formatNumber(summary.maxRowsPerSec)} rows/sec
                      </span>
                    </div>

                    <div class={styles.metric}>
                      <span class={styles.metricLabel}>Min</span>
                      <span class={styles.metricValue}>
                        {formatNumber(summary.minRowsPerSec)} rows/sec
                      </span>
                    </div>
                  </div>
                </div>
              </div>
            )}
          </For>
        </div>
      </Show>
    </Card>
  );
};