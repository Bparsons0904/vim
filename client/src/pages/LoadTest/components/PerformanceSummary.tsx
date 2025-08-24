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
      default:
        return method;
    }
  };

  const getMethodColor = (method: string): string => {
    switch (method) {
      case 'brute_force':
        return 'var(--color-warning)';
      case 'batched':
        return 'var(--color-info)';
      case 'optimized':
        return 'var(--color-success)';
      case 'ludicrous':
        return 'var(--color-primary)';
      default:
        return 'var(--color-text-secondary)';
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
          <For each={props.summaries}>
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

                <div class={styles.metricsGrid}>
                  <div class={styles.metric}>
                    <span class={styles.metricLabel}>Avg Throughput</span>
                    <span class={styles.metricValue}>
                      {formatNumber(summary.avgRowsPerSec)} rows/sec
                    </span>
                  </div>

                  <div class={styles.metric}>
                    <span class={styles.metricLabel}>Max Throughput</span>
                    <span class={styles.metricValue}>
                      {formatNumber(summary.maxRowsPerSec)} rows/sec
                    </span>
                  </div>

                  <div class={styles.metric}>
                    <span class={styles.metricLabel}>Min Throughput</span>
                    <span class={styles.metricValue}>
                      {formatNumber(summary.minRowsPerSec)} rows/sec
                    </span>
                  </div>

                  <div class={styles.metric}>
                    <span class={styles.metricLabel}>P95 Throughput</span>
                    <span class={styles.metricValue}>
                      {formatNumber(summary.p95RowsPerSec)} rows/sec
                    </span>
                  </div>

                  <div class={styles.metric}>
                    <span class={styles.metricLabel}>Avg Total Time</span>
                    <span class={styles.metricValue}>
                      {formatTime(summary.avgTotalTime)}
                    </span>
                  </div>

                  <div class={styles.metric}>
                    <span class={styles.metricLabel}>Best Time</span>
                    <span class={styles.metricValue}>
                      {formatTime(summary.minTotalTime)}
                    </span>
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