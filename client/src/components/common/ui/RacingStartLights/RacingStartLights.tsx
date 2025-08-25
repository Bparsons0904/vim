import { Component, createSignal, onMount, onCleanup } from "solid-js";
import styles from "./RacingStartLights.module.scss";

interface RacingStartLightsProps {
  onRaceStart?: () => void;
}

export const RacingStartLights: Component<RacingStartLightsProps> = (props) => {
  const [currentStep, setCurrentStep] = createSignal(0);

  let intervalId: number | undefined;

  const runContinuousSequence = () => {
    let step = 0;
    
    intervalId = window.setInterval(() => {
      step++;
      
      if (step <= 5) {
        // Light up lights 1-5 sequentially
        setCurrentStep(step);
      } else if (step === 6) {
        // All lights green moment (GO!)
        setCurrentStep(7);
        // Trigger racing animation
        props.onRaceStart?.();
      } else if (step === 7) {
        // Turn off lights
        setCurrentStep(0);
      } else if (step >= 8 && step < 26) {
        // Extended pause - 15 seconds (19 steps * 800ms = ~15s)
        // Keep lights off during this time
        setCurrentStep(0);
      } else if (step === 26) {
        // Reset for next loop
        setCurrentStep(0);
        step = 0; // Reset for next loop
      }
    }, 800); // 800ms intervals
  };

  onMount(() => {
    // Start the animation immediately when component mounts
    runContinuousSequence();
  });

  onCleanup(() => {
    if (intervalId) {
      clearInterval(intervalId);
    }
  });

  return (
    <div class={styles.racingLights}>
      <div class={styles.lightContainer}>
        <div class={styles.lightGrid}>
          {/* Single row of 5 lights that turn from red to green */}
          <div class={styles.lightRow}>
            <div
              class={`${styles.light} ${
                currentStep() >= 1 && currentStep() < 6 ? styles.red : ""
              } ${currentStep() === 7 ? styles.green : ""} ${
                currentStep() >= 1 ? styles.active : ""
              }`}
            />
            <div
              class={`${styles.light} ${
                currentStep() >= 2 && currentStep() < 6 ? styles.red : ""
              } ${currentStep() === 7 ? styles.green : ""} ${
                currentStep() >= 2 ? styles.active : ""
              }`}
            />
            <div
              class={`${styles.light} ${
                currentStep() >= 3 && currentStep() < 6 ? styles.red : ""
              } ${currentStep() === 7 ? styles.green : ""} ${
                currentStep() >= 3 ? styles.active : ""
              }`}
            />
            <div
              class={`${styles.light} ${
                currentStep() >= 4 && currentStep() < 6 ? styles.red : ""
              } ${currentStep() === 7 ? styles.green : ""} ${
                currentStep() >= 4 ? styles.active : ""
              }`}
            />
            <div
              class={`${styles.light} ${
                currentStep() >= 5 && currentStep() < 6 ? styles.red : ""
              } ${currentStep() === 7 ? styles.green : ""} ${
                currentStep() >= 5 ? styles.active : ""
              }`}
            />
          </div>
        </div>
      </div>
    </div>
  );
};