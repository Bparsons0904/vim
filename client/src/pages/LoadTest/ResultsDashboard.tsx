import { Component, Show, For, createMemo } from "solid-js";
import { Card } from "@components/common/ui/Card/Card";
import { Button } from "@components/common/ui/Button/Button";
import styles from "./ResultsDashboard.module.scss";
import { LoadTestResult } from "./LoadTest";
import { ApiClientError, NetworkError } from "@services/api/apiTypes";

interface ResultsDashboardProps {
  currentTest: LoadTestResult | null;
  testHistory: LoadTestResult[];
  isHistoryLoading?: boolean;
  historyError?: ApiClientError | NetworkError | null;
}

export const ResultsDashboard: Component<ResultsDashboardProps> = (props) => {
  const completedTests = createMemo(() => 
    props.testHistory.filter(test => test.status === 'completed')
  );

  const sortedTestHistory = createMemo(() => {
    // Sort by creation time or ID (most recent first) to ensure chronological order
    return [...props.testHistory].sort((a, b) => {
      // Assuming the API returns tests with more recent tests having larger IDs or timestamps
      // If we have timestamps, use those; otherwise use ID comparison
      return b.id.localeCompare(a.id);
    });
  });

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

  const getBestPerformance = createMemo(() => {
    const completed = completedTests();
    if (completed.length === 0) return null;

    return completed.reduce((best, test) => {
      if (!test.totalTime || !best.totalTime) return test.totalTime ? test : best;
      return test.totalTime < best.totalTime ? test : best;
    });
  });

  const getAveragePerformance = createMemo(() => {
    const completed = completedTests();
    if (completed.length === 0) return null;

    const validTests = completed.filter(test => test.totalTime);
    if (validTests.length === 0) return null;

    const avgTime = validTests.reduce((sum, test) => sum + (test.totalTime || 0), 0) / validTests.length;
    const avgRows = validTests.reduce((sum, test) => sum + test.rows, 0) / validTests.length;
    
    return {
      avgTime: Math.round(avgTime),
      avgRows: Math.round(avgRows),
      avgRate: Math.round((avgRows / avgTime) * 1000)
    };
  });

  return (
    <div class={styles.dashboard}>
      {/* Current Test Summary */}
      <Show when={props.currentTest}>
        <Card class={styles.currentTestCard}>
          <h2>Current Test</h2>
          <div class={styles.testSummary}>
            <div class={styles.testHeader}>
              <div class={styles.testStatus}>
                <span 
                  class={styles.statusIcon}
                  style={{ color: getStatusColor(props.currentTest!.status) }}
                >
                  {getStatusIcon(props.currentTest!.status)}
                </span>
                <span class={styles.statusText}>
                  {props.currentTest!.status.charAt(0).toUpperCase() + props.currentTest!.status.slice(1)}
                </span>
              </div>
              <div class={styles.testMethod}>
                {props.currentTest!.method === 'optimized' ? 'Optimized' : 'Brute Force'}
              </div>
            </div>
            
            <div class={styles.testDetails}>
              <div class={styles.testDetail}>
                <span class={styles.detailLabel}>Rows:</span>
                <span class={styles.detailValue}>{props.currentTest!.rows.toLocaleString()}</span>
              </div>
              <div class={styles.testDetail}>
                <span class={styles.detailLabel}>Columns:</span>
                <span class={styles.detailValue}>
                  {props.currentTest!.columns} ({props.currentTest!.dateColumns} date)
                </span>
              </div>
            </div>

            <Show when={props.currentTest!.status === 'completed'}>
              <div class={styles.performanceMetrics}>
                <div class={styles.metric}>
                  <span class={styles.metricLabel}>CSV Generation</span>
                  <span class={styles.metricValue}>{formatTime(props.currentTest!.csvGenTime)}</span>
                </div>
                <div class={styles.metric}>
                  <span class={styles.metricLabel}>Parsing</span>
                  <span class={styles.metricValue}>{formatTime(props.currentTest!.parseTime)}</span>
                </div>
                <div class={styles.metric}>
                  <span class={styles.metricLabel}>Insertion</span>
                  <span class={styles.metricValue}>{formatTime(props.currentTest!.insertTime)}</span>
                </div>
                <div class={styles.metric}>
                  <span class={styles.metricLabel}>Total Time</span>
                  <span class={styles.metricValue}>{formatTime(props.currentTest!.totalTime)}</span>
                </div>
                <div class={styles.metric}>
                  <span class={styles.metricLabel}>Insert Rate</span>
                  <span class={styles.metricValue}>
                    {formatRate(props.currentTest!.rows, props.currentTest!.insertTime)}
                  </span>
                </div>
              </div>
            </Show>
          </div>
        </Card>
      </Show>

      {/* Performance Summary */}
      <Show when={completedTests().length > 0}>
        <Card class={styles.summaryCard}>
          <h2>Performance Summary</h2>
          <div class={styles.summaryGrid}>
            <div class={styles.summaryItem}>
              <h3>Tests Completed</h3>
              <div class={styles.summaryValue}>{completedTests().length}</div>
            </div>
            
            <Show when={getBestPerformance()}>
              <div class={styles.summaryItem}>
                <h3>Best Performance</h3>
                <div class={styles.summaryValue}>{formatTime(getBestPerformance()!.totalTime)}</div>
                <div class={styles.summaryDetail}>
                  {getBestPerformance()!.rows.toLocaleString()} rows
                </div>
              </div>
            </Show>
            
            <Show when={getAveragePerformance()}>
              <div class={styles.summaryItem}>
                <h3>Average Rate</h3>
                <div class={styles.summaryValue}>
                  {getAveragePerformance()!.avgRate.toLocaleString()} rows/sec
                </div>
                <div class={styles.summaryDetail}>
                  {getAveragePerformance()!.avgTime}ms avg
                </div>
              </div>
            </Show>
          </div>
        </Card>
      </Show>

      {/* Test History */}
      <Card class={styles.historyCard}>
        <div class={styles.historyHeader}>
          <h2>Test History</h2>
          <Show when={props.testHistory.length > 0}>
            <Button variant="secondary" size="sm">
              Export Results
            </Button>
          </Show>
        </div>
        
        <Show when={props.historyError}>
          <div class={styles.errorState}>
            <p>Failed to load test history</p>
            <p class={styles.errorSubtext}>
              {props.historyError instanceof ApiClientError 
                ? props.historyError.message 
                : 'Please check your connection and try again'}
            </p>
          </div>
        </Show>

        <Show when={props.isHistoryLoading}>
          <div class={styles.loadingState}>
            <p>Loading test history...</p>
          </div>
        </Show>

        <Show 
          when={!props.isHistoryLoading && !props.historyError && props.testHistory.length > 0}
          fallback={
            <Show when={!props.isHistoryLoading && !props.historyError}>
              <div class={styles.emptyState}>
                <p>No tests completed yet</p>
                <p class={styles.emptySubtext}>Start your first performance test to see results here</p>
              </div>
            </Show>
          }
        >
          <div class={styles.historyList}>
            <For each={sortedTestHistory().slice(0, 10)}>
              {(test, index) => (
                <div class={styles.historyItem}>
                  <div class={styles.historyHeader}>
                    <div class={styles.historyTitle}>
                      <span 
                        class={styles.statusIcon}
                        style={{ color: getStatusColor(test.status) }}
                      >
                        {getStatusIcon(test.status)}
                      </span>
                      Test #{index() + 1}
                    </div>
                    <div class={styles.historyMethod}>
                      {test.method === 'optimized' ? 'Optimized' : 'Brute Force'}
                    </div>
                  </div>
                  
                  <div class={styles.historyDetails}>
                    <span>{test.rows.toLocaleString()} rows</span>
                    <span>{test.columns} columns</span>
                    <Show when={test.status === 'completed' && test.totalTime}>
                      <span>{formatTime(test.totalTime)}</span>
                      <span>{formatRate(test.rows, test.totalTime)}</span>
                    </Show>
                  </div>
                  
                  <Show when={test.status === 'failed' && test.errorMessage}>
                    <div class={styles.errorPreview}>
                      Error: {test.errorMessage!.substring(0, 50)}
                      {test.errorMessage!.length > 50 && '...'}
                    </div>
                  </Show>
                </div>
              )}
            </For>
            
            <Show when={sortedTestHistory().length > 10}>
              <div class={styles.showMore}>
                <Button variant="ghost" size="sm">
                  Show {sortedTestHistory().length - 10} more tests
                </Button>
              </div>
            </Show>
          </div>
        </Show>
      </Card>

    </div>
  );
};