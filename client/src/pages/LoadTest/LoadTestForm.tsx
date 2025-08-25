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
  const [method, setMethod] = createSignal<
    "brute_force" | "batched" | "optimized" | "ludicrous" | "plaid"
  >("brute_force");

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

  const getMethodColor = (methodValue: string): string => {
    switch (methodValue) {
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

  const getEstimatedTime = () => {
    let rowsPerSecond;
    switch (method()) {
      case "plaid":
        rowsPerSecond = 200000; // PostgreSQL COPY should be the fastest
        break;
      case "ludicrous":
        rowsPerSecond = 160000; // Double the optimized performance
        break;
      case "optimized":
        rowsPerSecond = 80000; // Estimated higher throughput for optimized method
        break;
      case "batched":
        rowsPerSecond = 50000;
        break;
      case "brute_force":
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
              {(preset) => {
                const isSelected = () => rows() === preset.value;
                return (
                  <Button
                    variant={isSelected() ? "primary" : "secondary"}
                    size="sm"
                    onClick={() => handlePresetClick(preset.value)}
                    disabled={props.disabled}
                  >
                    {preset.label}
                  </Button>
                );
              }}
            </For>
          </div>
        </div>

        {/* Method Selection */}
        <div class={styles.section}>
          <h3>Insertion Method</h3>
          <div class={styles.methodOptions}>
            <div class={`${styles.methodOption} ${method() === "brute_force" ? styles.selected : ""}`}
                 style={method() === "brute_force" ? {"border-color": getMethodColor("brute_force")} : {}}
                 onClick={() => !props.disabled && setMethod("brute_force")}>
              <div class={styles.methodContent}>
                <div class={styles.methodHeader}>
                  <div 
                    class={styles.methodColorIndicator}
                    style={{"background-color": getMethodColor("brute_force")}}
                  />
                  <strong>Brute Force</strong>
                </div>
                <p>Single row inserts - slower but simple</p>
              </div>
            </div>

            <div class={`${styles.methodOption} ${method() === "batched" ? styles.selected : ""}`}
                 style={method() === "batched" ? {"border-color": getMethodColor("batched")} : {}}
                 onClick={() => !props.disabled && setMethod("batched")}>
              <div class={styles.methodContent}>
                <div class={styles.methodHeader}>
                  <div 
                    class={styles.methodColorIndicator}
                    style={{"background-color": getMethodColor("batched")}}
                  />
                  <strong>Batched</strong>
                </div>
                <p>Batch inserts with GORM - faster</p>
              </div>
            </div>

            <div class={`${styles.methodOption} ${method() === "optimized" ? styles.selected : ""}`}
                 style={method() === "optimized" ? {"border-color": getMethodColor("optimized")} : {}}
                 onClick={() => !props.disabled && setMethod("optimized")}>
              <div class={styles.methodContent}>
                <div class={styles.methodHeader}>
                  <div 
                    class={styles.methodColorIndicator}
                    style={{"background-color": getMethodColor("optimized")}}
                  />
                  <strong>Optimized</strong>
                </div>
                <p>Streaming pipeline with concurrent workers</p>
              </div>
            </div>

            <div class={`${styles.methodOption} ${method() === "ludicrous" ? styles.selected : ""}`}
                 style={method() === "ludicrous" ? {"border-color": getMethodColor("ludicrous")} : {}}
                 onClick={() => !props.disabled && setMethod("ludicrous")}>
              <div class={styles.methodContent}>
                <div class={styles.methodHeader}>
                  <div 
                    class={styles.methodColorIndicator}
                    style={{"background-color": getMethodColor("ludicrous")}}
                  />
                  <strong>Ludicrous Speed</strong>
                </div>
                <p>Raw SQL with minimal overhead - insanely fast</p>
              </div>
            </div>

            <div class={`${styles.methodOption} ${method() === "plaid" ? styles.selected : ""}`}
                 style={method() === "plaid" ? {"border-color": getMethodColor("plaid")} : {}}
                 onClick={() => !props.disabled && setMethod("plaid")}>
              <div class={styles.methodContent}>
                <div class={styles.methodHeader}>
                  <div 
                    class={styles.methodColorIndicator}
                    style={{"background-color": getMethodColor("plaid")}}
                  />
                  <strong>Plaid Speed</strong>
                </div>
                <p>PostgreSQL Streaming - ultimate performance</p>
              </div>
            </div>
          </div>
        </div>

        {/* Test Summary */}
        <div class={styles.summary}>
          <h3>Test Summary</h3>
          <div class={styles.summaryGrid}>
            <div class={styles.summaryItem}>
              <span class={styles.summaryLabel}>Rows: {rows().toLocaleString()}</span>
            </div>
            <div class={styles.summaryItem}>
              <span class={styles.summaryLabel}>
                Method: {method() === "batched"
                  ? "Batched"
                  : method() === "optimized"
                    ? "Optimized"
                    : method() === "ludicrous"
                      ? "Ludicrous Speed"
                      : method() === "plaid"
                        ? "Plaid (COPY)"
                        : "Brute Force"}
              </span>
            </div>
            <div class={styles.summaryItem}>
              <span class={styles.summaryLabel}>Est. Time: {getEstimatedTime()}</span>
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
