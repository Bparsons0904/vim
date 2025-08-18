import { Component, JSX, Show, createEffect, onCleanup } from "solid-js";
import { Portal } from "solid-js/web";
import styles from "./Modal.module.scss";

export enum ModalSize {
  Small = "sm",
  Medium = "md",
  Large = "lg",
  ExtraLarge = "xl"
}

interface ModalProps {
  isOpen: boolean;
  onClose: () => void;
  children: JSX.Element;
  size?: ModalSize;
  showCloseButton?: boolean;
  closeOnBackdropClick?: boolean;
  closeOnEscape?: boolean;
  title?: string;
  className?: string;
  backdropClassName?: string;
}

export const Modal: Component<ModalProps> = (props) => {
  let modalRef: HTMLDivElement | undefined;

  const handleKeyDown = (e: KeyboardEvent) => {
    if (e.key === "Escape" && props.closeOnEscape !== false) {
      props.onClose();
    }
  };

  const handleBackdropClick = (e: MouseEvent) => {
    if (e.target === e.currentTarget && props.closeOnBackdropClick !== false) {
      props.onClose();
    }
  };

  const trapFocus = (e: KeyboardEvent) => {
    if (!modalRef || e.key !== "Tab") return;

    const focusableElements = modalRef.querySelectorAll(
      'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
    );
    const firstElement = focusableElements[0] as HTMLElement;
    const lastElement = focusableElements[focusableElements.length - 1] as HTMLElement;

    if (e.shiftKey) {
      if (document.activeElement === firstElement) {
        lastElement?.focus();
        e.preventDefault();
      }
    } else {
      if (document.activeElement === lastElement) {
        firstElement?.focus();
        e.preventDefault();
      }
    }
  };

  createEffect(() => {
    if (props.isOpen) {
      document.addEventListener("keydown", handleKeyDown);
      document.addEventListener("keydown", trapFocus);
      document.body.style.overflow = "hidden";
      
      setTimeout(() => {
        const focusableElement = modalRef?.querySelector(
          'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
        ) as HTMLElement;
        focusableElement?.focus();
      }, 100);
    } else {
      document.removeEventListener("keydown", handleKeyDown);
      document.removeEventListener("keydown", trapFocus);
      document.body.style.overflow = "";
    }
  });

  onCleanup(() => {
    document.removeEventListener("keydown", handleKeyDown);
    document.removeEventListener("keydown", trapFocus);
    document.body.style.overflow = "";
  });

  const sizeClass = () => {
    switch (props.size) {
      case ModalSize.Small:
        return styles.modalSmall;
      case ModalSize.Large:
        return styles.modalLarge;
      case ModalSize.ExtraLarge:
        return styles.modalExtraLarge;
      default:
        return styles.modalMedium;
    }
  };

  return (
    <Show when={props.isOpen}>
      <Portal>
        <div
          class={`${styles.backdrop} ${props.backdropClassName || ""}`}
          onClick={handleBackdropClick}
          role="dialog"
          aria-modal="true"
          aria-labelledby={props.title ? "modal-title" : undefined}
        >
          <div
            ref={modalRef}
            class={`${styles.modal} ${sizeClass()} ${props.className || ""}`}
            onClick={(e) => e.stopPropagation()}
          >
            <Show when={props.title || props.showCloseButton !== false}>
              <div class={styles.header}>
                <Show when={props.title}>
                  <h2 id="modal-title" class={styles.title}>
                    {props.title}
                  </h2>
                </Show>
                <Show when={props.showCloseButton !== false}>
                  <button
                    class={styles.closeButton}
                    onClick={props.onClose}
                    aria-label="Close modal"
                    type="button"
                  >
                    <svg
                      width="24"
                      height="24"
                      viewBox="0 0 24 24"
                      fill="none"
                      stroke="currentColor"
                      stroke-width="2"
                      stroke-linecap="round"
                      stroke-linejoin="round"
                    >
                      <line x1="18" y1="6" x2="6" y2="18" />
                      <line x1="6" y1="6" x2="18" y2="18" />
                    </svg>
                  </button>
                </Show>
              </div>
            </Show>
            <div class={styles.content}>
              {props.children}
            </div>
          </div>
        </div>
      </Portal>
    </Show>
  );
};