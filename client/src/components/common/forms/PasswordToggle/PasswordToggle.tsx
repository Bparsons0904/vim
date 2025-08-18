import { Component, Show } from "solid-js";
import styles from "./PasswordToggle.module.scss";

interface PasswordToggleProps {
  showPassword: boolean;
  onToggle: () => void;
}

export const PasswordToggle: Component<PasswordToggleProps> = (props) => {
  return (
    <button
      type="button"
      class={styles.toggleButton}
      onClick={props.onToggle}
      aria-label={props.showPassword ? "Hide password" : "Show password"}
    >
      <Show
        when={props.showPassword}
        fallback={
          <svg
            class={styles.icon}
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
          >
            <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" />
            <circle cx="12" cy="12" r="3" />
          </svg>
        }
      >
        <svg
          class={styles.icon}
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
        >
          <path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24" />
          <line x1="1" y1="1" x2="23" y2="23" />
        </svg>
      </Show>
    </button>
  );
};