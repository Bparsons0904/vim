import { Component, Show, For } from "solid-js";
import { Card } from "@components/common/ui/Card/Card";
import { LoadTestResult } from "../LoadTest";
import styles from "./TestHistory.module.scss";

interface TestHistoryProps {
  tests: LoadTestResult[];
  isLoading?: boolean;
  error?: Error | null;
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

  const formatTestDate = (createdAt: string): string => {
    const date = new Date(createdAt);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffMinutes = Math.floor(diffMs / (1000 * 60));
    const diffHours = Math.floor(diffMs / (1000 * 60 * 60));
    const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));

    if (diffMinutes < 1) return 'Just now';
    if (diffMinutes < 60) return `${diffMinutes}m ago`;
    if (diffHours < 24) return `${diffHours}h ago`;
    if (diffDays < 7) return `${diffDays}d ago`;
    
    // For older tests, show the actual date
    return date.toLocaleDateString(undefined, { 
      month: 'short', 
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
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

  // Server now returns tests in reverse chronological order (most recent first)
  const recentTests = () => props.tests;

  return (
    <Card class={styles.testHistoryCard}>
      <div class={styles.header}>
        <h2>Recent Test History</h2>
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
            {(test) => (
              <div class={styles.historyItem}>
                <div class={styles.itemHeader}>
                  <div class={styles.itemTitle}>
                    <span 
                      class={styles.statusIcon}
                      style={{ color: getStatusColor(test.status) }}
                    >
                      {getStatusIcon(test.status)}
                    </span>
                    {formatTestDate(test.createdAt)}
                  </div>
                  <div class={styles.itemMethod}>
                    <span 
                      class={styles.methodIcon}
                      style={{ 'background-color': getMethodColor(test.method) }}
                    />
                    {getMethodDisplayName(test.method)}
                  </div>
                </div>
                
                <div class={styles.itemDetails}>
                  <span class={styles.detail}>{test.rows.toLocaleString()} rows</span>
                  <Show when={test.status === 'completed' && test.parseTime && test.insertTime}>
                    <span class={styles.detail}>{formatTime((test.parseTime || 0) + (test.insertTime || 0))}</span>
                    <span class={styles.detail}>{formatRate(test.rows, (test.parseTime || 0) + (test.insertTime || 0))}</span>
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
        </div>
      </Show>
    </Card>
  );
};