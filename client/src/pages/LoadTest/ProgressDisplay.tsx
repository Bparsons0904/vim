import { Component, createSignal, createEffect, onCleanup, Show } from "solid-js";
import { Card } from "@components/common/ui/Card/Card";
import styles from "./ProgressDisplay.module.scss";
import { LoadTestResult } from "./LoadTest";

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
  const [intervalId, setIntervalId] = createSignal<NodeJS.Timeout | null>(null);

  // Mock progress simulation for now - will be replaced with WebSocket updates
  createEffect(() => {
    if (props.test.status === 'running') {
      setStartTime(Date.now());
      simulateProgress();
    }
  });

  const simulateProgress = () => {
    let currentProgress = 0;
    let phase: ProgressState['phase'] = 'csv_generation';
    
    const updateProgress = () => {
      const elapsed = Date.now() - startTime();
      currentProgress += Math.random() * 5 + 1; // Simulate variable progress
      
      if (currentProgress >= 100) {
        currentProgress = 100;
        phase = 'completed';
      } else if (currentProgress > 85) {
        phase = 'insertion';
      } else if (currentProgress > 25) {
        phase = 'parsing';
      } else {
        phase = 'csv_generation';
      }

      const rowsProcessed = Math.floor((currentProgress / 100) * props.test.rows);
      const rowsPerSecond = elapsed > 0 ? Math.floor((rowsProcessed / elapsed) * 1000) : 0;
      const remainingRows = props.test.rows - rowsProcessed;
      const etaSeconds = rowsPerSecond > 0 ? Math.ceil(remainingRows / rowsPerSecond) : 0;
      
      setProgress({
        phase,
        overallProgress: currentProgress,
        phaseProgress: getPhaseProgress(phase, currentProgress),
        currentPhase: getPhaseLabel(phase),
        rowsProcessed,
        rowsPerSecond,
        eta: formatETA(etaSeconds),
        message: getPhaseMessage(phase, rowsProcessed)
      });

      if (currentProgress >= 100) {
        // Test completed
        const completedTest: LoadTestResult = {
          ...props.test,
          status: 'completed',
          csvGenTime: Math.floor(Math.random() * 5000) + 1000,
          parseTime: Math.floor(Math.random() * 3000) + 500,
          insertTime: Math.floor(Math.random() * 10000) + 2000,
          totalTime: elapsed
        };
        props.onTestComplete(completedTest);
        
        if (intervalId()) {
          clearInterval(intervalId()!);
          setIntervalId(null);
        }
      }
    };

    const id = setInterval(updateProgress, 500);
    setIntervalId(id);
  };

  const getPhaseProgress = (phase: ProgressState['phase'], overall: number): number => {
    switch (phase) {
      case 'csv_generation':
        return Math.min(overall * 4, 100); // 0-25% overall = 0-100% phase
      case 'parsing':
        return Math.min((overall - 25) * (100 / 60), 100); // 25-85% overall = 0-100% phase
      case 'insertion':
        return Math.min((overall - 85) * (100 / 15), 100); // 85-100% overall = 0-100% phase
      default:
        return 100;
    }
  };

  const getPhaseLabel = (phase: ProgressState['phase']): string => {
    switch (phase) {
      case 'csv_generation':
        return 'Generating CSV Data';
      case 'parsing':
        return 'Parsing and Validating';
      case 'insertion':
        return 'Inserting into Database';
      case 'completed':
        return 'Test Completed';
      case 'failed':
        return 'Test Failed';
      default:
        return 'Unknown Phase';
    }
  };

  const getPhaseMessage = (phase: ProgressState['phase'], rowsProcessed: number): string => {
    switch (phase) {
      case 'csv_generation':
        return 'Generating realistic test data with faker library...';
      case 'parsing':
        return 'Validating date formats and parsing CSV data...';
      case 'insertion':
        return `Inserting data using ${props.test.method === 'optimized' ? 'batch' : 'single row'} method...`;
      case 'completed':
        return `Successfully processed ${rowsProcessed.toLocaleString()} rows!`;
      case 'failed':
        return 'Test failed. Check error details below.';
      default:
        return 'Processing...';
    }
  };

  const formatETA = (seconds: number): string => {
    if (seconds <= 0) return 'Done';
    if (seconds < 60) return `${seconds}s`;
    if (seconds < 3600) return `${Math.ceil(seconds / 60)}m`;
    return `${Math.ceil(seconds / 3600)}h`;
  };

  const getProgressBarColor = (): string => {
    if (progress().phase === 'failed') return 'var(--color-error)';
    if (progress().phase === 'completed') return 'var(--color-success)';
    return 'var(--color-primary)';
  };

  onCleanup(() => {
    if (intervalId()) {
      clearInterval(intervalId()!);
    }
  });

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