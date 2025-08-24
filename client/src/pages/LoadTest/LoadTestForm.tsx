import { Component, createSignal, For } from "solid-js";
import { Button } from "@components/common/ui/Button/Button";
import { TextInput } from "@components/common/forms/TextInput/TextInput";
import { Card } from "@components/common/ui/Card/Card";
import styles from "./LoadTestForm.module.scss";
import { LoadTestConfig } from "./LoadTest";

interface LoadTestFormProps {
  onStartTest: (config: LoadTestConfig) => void;
  isLoading: boolean;
  disabled: boolean;
}

const ROW_PRESETS = [
  { label: "10K", value: 10000 },
  { label: "50K", value: 50000 },
  { label: "100K", value: 100000 },
  { label: "500K", value: 500000 },
  { label: "1M", value: 1000000 },
  { label: "5M", value: 5000000 },
  { label: "10M", value: 10000000 },
];

export const LoadTestForm: Component<LoadTestFormProps> = (props) => {
  const [rows, setRows] = createSignal(10000);
  const [method, setMethod] = createSignal<'brute_force' | 'batched' | 'optimized'>('batched');

  const handlePresetClick = (value: number) => {
    setRows(value);
  };

  const handleSubmit = (e: Event) => {
    e.preventDefault();
    
    const config: LoadTestConfig = {
      rows: rows(),
      method: method(),
    };

    props.onStartTest(config);
  };

  const getEstimatedTime = () => {
    let rowsPerSecond;
    switch (method()) {
      case 'optimized':
        rowsPerSecond = 80000; // Estimated higher throughput for optimized method
        break;
      case 'batched':
        rowsPerSecond = 50000;
        break;
      case 'brute_force':
      default:
        rowsPerSecond = 100;
        break;
    }
    
    const estimatedSeconds = Math.ceil(rows() / rowsPerSecond);
    
    if (estimatedSeconds < 60) {
      return `~${estimatedSeconds}s`;
    } else if (estimatedSeconds < 3600) {
      return `~${Math.ceil(estimatedSeconds / 60)}m`;
    } else {
      return `~${Math.ceil(estimatedSeconds / 3600)}h`;
    }
  };

  return (
    <Card class={styles.formCard}>
      <h2>Test Configuration</h2>
      
      <form onSubmit={handleSubmit} class={styles.form}>
        {/* Row Count Selection */}
        <div class={styles.section}>
          <h3>Row Count</h3>
          <div class={styles.presetButtons}>
            <For each={ROW_PRESETS}>
              {(preset) => (
                <Button
                  variant={rows() === preset.value ? "primary" : "secondary"}
                  size="sm"
                  onClick={() => handlePresetClick(preset.value)}
                  disabled={props.disabled}
                >
                  {preset.label}
                </Button>
              )}
            </For>
          </div>
        </div>

        {/* Fixed Column Configuration Info */}
        <div class={styles.section}>
          <h3>Column Configuration</h3>
          <div class={styles.fixedConfig}>
            <p>Uses a fixed set of 25 meaningful demographic, employment, and health columns plus 125 additional columns for testing (150 total).</p>
            <div class={styles.configDetails}>
              <span><strong>25</strong> meaningful columns (names, addresses, employment, insurance)</span>
              <span><strong>125</strong> ignored filler columns</span>
              <span><strong>6</strong> date columns with validation</span>
            </div>
          </div>
        </div>

        {/* Method Selection */}
        <div class={styles.section}>
          <h3>Insertion Method</h3>
          <div class={styles.methodOptions}>
            <label class={styles.radioOption}>
              <input
                type="radio"
                name="method"
                value="brute_force"
                checked={method() === 'brute_force'}
                onChange={() => setMethod('brute_force')}
                disabled={props.disabled}
              />
              <div class={styles.radioContent}>
                <strong>Brute Force</strong>
                <p>Single row inserts - slower but simple</p>
              </div>
            </label>
            
            <label class={styles.radioOption}>
              <input
                type="radio"
                name="method"
                value="batched"
                checked={method() === 'batched'}
                onChange={() => setMethod('batched')}
                disabled={props.disabled}
              />
              <div class={styles.radioContent}>
                <strong>Batched</strong>
                <p>Batch inserts with transactions - faster</p>
              </div>
            </label>
            
            <label class={styles.radioOption}>
              <input
                type="radio"
                name="method"
                value="optimized"
                checked={method() === 'optimized'}
                onChange={() => setMethod('optimized')}
                disabled={props.disabled}
              />
              <div class={styles.radioContent}>
                <strong>Optimized</strong>
                <p>Streaming pipeline with concurrent workers - fastest</p>
              </div>
            </label>
          </div>
        </div>

        {/* Test Summary */}
        <div class={styles.summary}>
          <h3>Test Summary</h3>
          <div class={styles.summaryGrid}>
            <div class={styles.summaryItem}>
              <span class={styles.summaryLabel}>Rows:</span>
              <span class={styles.summaryValue}>{rows().toLocaleString()}</span>
            </div>
            <div class={styles.summaryItem}>
              <span class={styles.summaryLabel}>Method:</span>
              <span class={styles.summaryValue}>
                {method() === 'batched' ? 'Batched' : method() === 'optimized' ? 'Optimized' : 'Brute Force'}
              </span>
            </div>
            <div class={styles.summaryItem}>
              <span class={styles.summaryLabel}>Est. Time:</span>
              <span class={styles.summaryValue}>{getEstimatedTime()}</span>
            </div>
          </div>
        </div>

        <Button
          type="submit"
          variant="primary"
          size="lg"
          disabled={props.disabled || props.isLoading}
          class={styles.submitButton}
        >
          {props.isLoading ? "Starting Test..." : "Start Performance Test"}
        </Button>
      </form>
    </Card>
  );
};