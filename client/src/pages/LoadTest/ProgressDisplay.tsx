import { Component, createSignal, createEffect, onCleanup, Show } from "solid-js";
import { Card } from "@components/common/ui/Card/Card";
import styles from "./ProgressDisplay.module.scss";
import { LoadTestResult } from "./LoadTest";
import { useWebSocket } from "@context/WebSocketContext";

interface ProgressDisplayProps {
  test: LoadTestResult;
  onTestComplete: (result: LoadTestResult) => void;
}

interface ProgressState {
  phase: 'csv_generation' | 'parsing' | 'insertion' | 'completed' | 'failed';
  overallProgress: number;
  phaseProgress: number;
  currentPhase: string;
  rowsProcessed: number;
  rowsPerSecond: number;
  eta: string;
  message: string;
}

export const ProgressDisplay: Component<ProgressDisplayProps> = (props) => {
  const { onLoadTestProgress, onLoadTestComplete, onLoadTestError } = useWebSocket();
  
  const [progress, setProgress] = createSignal<ProgressState>({
    phase: 'csv_generation',
    overallProgress: 0,
    phaseProgress: 0,
    currentPhase: 'Initializing...',
    rowsProcessed: 0,
    rowsPerSecond: 0,
    eta: 'Calculating...',
    message: 'Starting test...'
  });

  const [startTime, setStartTime] = createSignal(Date.now());

  // Real WebSocket-based progress updates
  createEffect(() => {
    if (props.test.status === 'running') {
      setStartTime(Date.now());

      // Register WebSocket progress handlers
      const unsubscribeProgress = onLoadTestProgress((testId, data) => {
        if (testId === props.test.id) {
          setProgress({
            phase: data.phase as ProgressState['phase'] || 'csv_generation',
            overallProgress: Number(data.overallProgress) || 0,
            phaseProgress: Number(data.phaseProgress) || 0,
            currentPhase: String(data.currentPhase) || 'Processing...',
            rowsProcessed: Number(data.rowsProcessed) || 0,
            rowsPerSecond: Number(data.rowsPerSecond) || 0,
            eta: String(data.eta) || 'Calculating...',
            message: String(data.message) || 'Processing...'
          });
        }
      });

      const unsubscribeComplete = onLoadTestComplete((testId, data) => {
        if (testId === props.test.id) {
          const completedTest: LoadTestResult = {
            ...props.test,
            status: 'completed',
            csvGenTime: Number(data.csvGenTime),
            parseTime: Number(data.parseTime),
            insertTime: Number(data.insertTime),
            totalTime: Number(data.totalTime)
          };
          props.onTestComplete(completedTest);
          
          setProgress(prev => ({
            ...prev,
            phase: 'completed',
            overallProgress: 100,
            phaseProgress: 100,
            currentPhase: 'Test Completed',
            message: `Successfully processed ${props.test.rows.toLocaleString()} rows!`
          }));
        }
      });

      const unsubscribeError = onLoadTestError((testId, error) => {
        if (testId === props.test.id) {
          const errorTest: LoadTestResult = {
            ...props.test,
            status: 'failed',
            errorMessage: error
          };
          props.onTestComplete(errorTest);
          
          setProgress(prev => ({
            ...prev,
            phase: 'failed',
            message: 'Test failed. Check error details below.'
          }));
        }
      });

      // Cleanup function
      onCleanup(() => {
        unsubscribeProgress();
        unsubscribeComplete();
        unsubscribeError();
      });
    }
  });



  const getProgressBarColor = (): string => {
    if (progress().phase === 'failed') return 'var(--color-error)';
    if (progress().phase === 'completed') return 'var(--color-success)';
    return 'var(--color-primary)';
  };


  return (
    <Card class={styles.progressCard}>
      <h2>Test Progress</h2>
      
      <div class={styles.progressContainer}>
        {/* Overall Progress */}
        <div class={styles.progressSection}>
          <div class={styles.progressHeader}>
            <h3>Overall Progress</h3>
            <span class={styles.percentage}>{Math.round(progress().overallProgress)}%</span>
          </div>
          <div class={styles.progressBarContainer}>
            <div 
              class={styles.progressBar}
              style={{ 
                width: `${progress().overallProgress}%`,
                'background-color': getProgressBarColor()
              }}
            />
          </div>
        </div>

        {/* Current Phase */}
        <div class={styles.progressSection}>
          <div class={styles.progressHeader}>
            <h3>{progress().currentPhase}</h3>
            <span class={styles.percentage}>{Math.round(progress().phaseProgress)}%</span>
          </div>
          <div class={styles.progressBarContainer}>
            <div 
              class={styles.progressBar}
              style={{ 
                width: `${progress().phaseProgress}%`,
                'background-color': getProgressBarColor()
              }}
            />
          </div>
        </div>

        {/* Status Message */}
        <div class={styles.statusMessage}>
          {progress().message}
        </div>

        {/* Performance Metrics */}
        <Show when={props.test.status === 'running'}>
          <div class={styles.metricsGrid}>
            <div class={styles.metric}>
              <span class={styles.metricLabel}>Rows Processed</span>
              <span class={styles.metricValue}>
                {progress().rowsProcessed.toLocaleString()} / {props.test.rows.toLocaleString()}
              </span>
            </div>
            <div class={styles.metric}>
              <span class={styles.metricLabel}>Processing Rate</span>
              <span class={styles.metricValue}>
                {progress().rowsPerSecond.toLocaleString()} rows/sec
              </span>
            </div>
            <div class={styles.metric}>
              <span class={styles.metricLabel}>Time Elapsed</span>
              <span class={styles.metricValue}>
                {Math.floor((Date.now() - startTime()) / 1000)}s
              </span>
            </div>
            <div class={styles.metric}>
              <span class={styles.metricLabel}>ETA</span>
              <span class={styles.metricValue}>{progress().eta}</span>
            </div>
          </div>
        </Show>

        {/* Phase Timeline */}
        <div class={styles.timeline}>
          <div class={`${styles.timelineItem} ${progress().phase === 'csv_generation' || progress().overallProgress > 25 ? styles.active : ''} ${progress().overallProgress > 25 ? styles.completed : ''}`}>
            <div class={styles.timelineIcon}>1</div>
            <div class={styles.timelineContent}>
              <h4>CSV Generation</h4>
              <p>Generate test data</p>
            </div>
          </div>
          <div class={`${styles.timelineItem} ${progress().phase === 'parsing' || progress().overallProgress > 85 ? styles.active : ''} ${progress().overallProgress > 85 ? styles.completed : ''}`}>
            <div class={styles.timelineIcon}>2</div>
            <div class={styles.timelineContent}>
              <h4>Data Parsing</h4>
              <p>Validate and parse CSV</p>
            </div>
          </div>
          <div class={`${styles.timelineItem} ${progress().phase === 'insertion' || progress().overallProgress >= 100 ? styles.active : ''} ${progress().overallProgress >= 100 ? styles.completed : ''}`}>
            <div class={styles.timelineIcon}>3</div>
            <div class={styles.timelineContent}>
              <h4>Database Insertion</h4>
              <p>Insert into database</p>
            </div>
          </div>
        </div>

        {/* Error Display */}
        <Show when={props.test.status === 'failed' && props.test.errorMessage}>
          <div class={styles.errorContainer}>
            <h4>Error Details</h4>
            <p class={styles.errorMessage}>{props.test.errorMessage}</p>
          </div>
        </Show>
      </div>
    </Card>
  );
};