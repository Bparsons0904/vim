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
  { label: "100K", value: 100000 },
  { label: "1M", value: 1000000 },
  { label: "10M", value: 10000000 },
];

export const LoadTestForm: Component<LoadTestFormProps> = (props) => {
  const [rows, setRows] = createSignal(10000);
  const [customRows, setCustomRows] = createSignal("");
  const [columns, setColumns] = createSignal(50);
  const [datePercentage, setDatePercentage] = createSignal(20);
  const [method, setMethod] = createSignal<'brute_force' | 'optimized'>('optimized');
  const [useCustomRows, setUseCustomRows] = createSignal(false);

  const handlePresetClick = (value: number) => {
    setRows(value);
    setUseCustomRows(false);
    setCustomRows("");
  };

  const handleCustomRowsChange = (value: string) => {
    setCustomRows(value);
    const numValue = parseInt(value);
    if (!isNaN(numValue) && numValue > 0) {
      setRows(numValue);
      setUseCustomRows(true);
    }
  };

  const calculateDateColumns = () => {
    return Math.floor((columns() * datePercentage()) / 100);
  };

  const handleSubmit = (e: Event) => {
    e.preventDefault();
    
    const config: LoadTestConfig = {
      rows: rows(),
      columns: columns(),
      dateColumns: calculateDateColumns(),
      method: method(),
    };

    props.onStartTest(config);
  };

  const getEstimatedTime = () => {
    const rowsPerSecond = method() === 'optimized' ? 20000 : 100;
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
                  variant={!useCustomRows() && rows() === preset.value ? "primary" : "secondary"}
                  size="sm"
                  onClick={() => handlePresetClick(preset.value)}
                  disabled={props.disabled}
                >
                  {preset.label}
                </Button>
              )}
            </For>
          </div>
          
          <div class={styles.customInput}>
            <TextInput
              label="Custom Row Count"
              placeholder="Enter custom number"
              value={customRows()}
              onInput={handleCustomRowsChange}
              type="text"
              pattern="[0-9]*"
              class={styles.numberInput}
            />
          </div>
          
          <div class={styles.currentValue}>
            Selected: <strong>{rows().toLocaleString()}</strong> rows
          </div>
        </div>

        {/* Column Count */}
        <div class={styles.section}>
          <h3>Column Configuration</h3>
          <div class={styles.sliderGroup}>
            <label class={styles.sliderLabel}>
              <span>Total Columns: <strong>{columns()}</strong></span>
              <input
                type="range"
                min="10"
                max="200"
                step="10"
                value={columns()}
                onInput={(e) => setColumns(parseInt(e.target.value))}
                class={styles.slider}
                disabled={props.disabled}
              />
              <div class={styles.sliderRange}>
                <span>10</span>
                <span>200</span>
              </div>
            </label>
          </div>

          <div class={styles.sliderGroup}>
            <label class={styles.sliderLabel}>
              <span>Date Columns: <strong>{calculateDateColumns()}</strong> ({datePercentage()}%)</span>
              <input
                type="range"
                min="0"
                max="50"
                step="5"
                value={datePercentage()}
                onInput={(e) => setDatePercentage(parseInt(e.target.value))}
                class={styles.slider}
                disabled={props.disabled}
              />
              <div class={styles.sliderRange}>
                <span>0%</span>
                <span>50%</span>
              </div>
            </label>
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
                value="optimized"
                checked={method() === 'optimized'}
                onChange={() => setMethod('optimized')}
                disabled={props.disabled}
              />
              <div class={styles.radioContent}>
                <strong>Optimized</strong>
                <p>Batch inserts with transactions - faster</p>
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
              <span class={styles.summaryLabel}>Columns:</span>
              <span class={styles.summaryValue}>{columns()} ({calculateDateColumns()} date)</span>
            </div>
            <div class={styles.summaryItem}>
              <span class={styles.summaryLabel}>Method:</span>
              <span class={styles.summaryValue}>
                {method() === 'optimized' ? 'Optimized' : 'Brute Force'}
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