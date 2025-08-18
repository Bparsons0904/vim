import { Component } from "solid-js";
import { Modal, ModalSize } from "../Modal/Modal";
import { Button, ButtonVariant } from "../Button/Button";
import clsx from "clsx";
import styles from "./ConfirmPopup.module.scss";

export interface ConfirmPopupProps {
  isOpen: boolean;
  onClose: () => void;
  onConfirm: () => void;
  title?: string;
  description?: string;
  confirmText?: string;
  cancelText?: string;
  variant?: ButtonVariant;
  size?: ModalSize;
}

export const ConfirmPopup: Component<ConfirmPopupProps> = (props) => {
  const getIcon = () => {
    switch (props.variant) {
      case "danger":
        return (
          <svg class={styles.icon} viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="10" />
            <line x1="15" y1="9" x2="9" y2="15" />
            <line x1="9" y1="9" x2="15" y2="15" />
          </svg>
        );
      case "warning":
        return (
          <svg class={styles.icon} viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3Z" />
            <line x1="12" y1="9" x2="12" y2="13" />
            <line x1="12" y1="17" x2="12.01" y2="17" />
          </svg>
        );
      default:
        return (
          <svg class={styles.icon} viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="10" />
            <path d="m9 12 2 2 4-4" />
          </svg>
        );
    }
  };

  return (
    <Modal
      isOpen={props.isOpen}
      onClose={props.onClose}
      title={props.title || "Confirm Action"}
      size={props.size || ModalSize.Small}
    >
      <div class={styles.confirmPopup}>
        <div class={clsx(styles.iconContainer, styles[props.variant || "primary"])}>
          {getIcon()}
        </div>
        
        <div class={styles.content}>
          <p class={styles.description}>
            {props.description || "Are you sure you want to proceed with this action?"}
          </p>
        </div>

        <div class={styles.actions}>
          <Button
            variant="tertiary"
            onClick={props.onClose}
          >
            {props.cancelText || "Cancel"}
          </Button>
          <Button
            variant={props.variant || "primary"}
            onClick={props.onConfirm}
          >
            {props.confirmText || "Confirm"}
          </Button>
        </div>
      </div>
    </Modal>
  );
};