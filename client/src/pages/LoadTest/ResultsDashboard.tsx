import { Component, Show, For, createMemo } from "solid-js";
import { Card } from "@components/common/ui/Card/Card";
import { Button } from "@components/common/ui/Button/Button";
import { TestHistory } from "./components/TestHistory";
import { PerformanceSummary as PerformanceSummaryComponent } from "./components/PerformanceSummary";
import { usePerformanceSummary } from "@services/api/hooks/loadtest.hooks";
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
  const performanceSummaryQuery = usePerformanceSummary();
  
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
                  style={{ color: (() => {
                    switch (props.currentTest!.status) {
                      case 'completed': return 'var(--color-success)';
                      case 'running': return 'var(--color-primary)';
                      case 'failed': return 'var(--color-error)';
                      default: return 'var(--color-text-secondary)';
                    }
                  })() }}
                >
                  {(() => {
                    switch (props.currentTest!.status) {
                      case 'completed': return '✓';
                      case 'running': return '⟳';
                      case 'failed': return '✗';
                      default: return '?';
                    }
                  })()}
                </span>
                <span class={styles.statusText}>
                  {props.currentTest!.status.charAt(0).toUpperCase() + props.currentTest!.status.slice(1)}
                </span>
              </div>
              <div class={styles.testMethod}>
                {(() => {
                  switch (props.currentTest!.method) {
                    case 'brute_force':
                      return 'Brute Force';
                    case 'batched':
                      return 'Batched';
                    case 'optimized':
                      return 'Optimized';
                    case 'ludicrous':
                      return 'Ludicrous Speed';
                    default:
                      return props.currentTest!.method;
                  }
                })()}
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

      {/* Performance Summary */}
      <PerformanceSummaryComponent 
        summaries={performanceSummaryQuery.data?.performanceSummary || []}
        isLoading={performanceSummaryQuery.isLoading}
        error={performanceSummaryQuery.error}
      />

      {/* Test History */}
      <TestHistory 
        tests={sortedTestHistory()}
        isLoading={props.isHistoryLoading}
        error={props.historyError}
      />

    </div>
  );
};