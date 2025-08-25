import { Component, JSX } from "solid-js";
import styles from "./Button.module.scss";
import clsx from "clsx";

export type ButtonVariant = "primary" | "secondary" | "tertiary" | "danger" | "gradient" | "ghost" | "warning";
export type ButtonSize = "sm" | "md" | "lg";

interface ButtonProps {
  variant?: ButtonVariant;
  size?: ButtonSize;
  type?: "button" | "submit" | "reset";
  disabled?: boolean;
  onClick?: (event: MouseEvent) => void;
  children: JSX.Element;
  class?: string;
  className?: string;
}

export const Button: Component<ButtonProps> = (props) => {
  return (
    <button
      class={clsx(
        styles.button,
        styles[props.variant || "primary"],
        styles[props.size || "md"],
        props.class || props.className
      )}
      type={props.type || "button"}
      disabled={props.disabled}
      onClick={props.onClick}
    >
      {props.children}
    </button>
  );
};
