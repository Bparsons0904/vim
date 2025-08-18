import { Component, createUniqueId, Show } from "solid-js";
import { useFormField } from "@hooks/useFormField";
import { useValidation } from "@hooks/useValidation";
import { ValidatorFunction } from "../../../../utils/validation";
import styles from "./Textarea.module.scss";

export interface TextareaProps {
  name?: string;
  label?: string;
  placeholder?: string;
  value?: string;
  required?: boolean;
  disabled?: boolean;
  rows?: number;
  maxLength?: number;
  showCharacterCount?: boolean;
  customValidators?: ValidatorFunction[];
  class?: string;
  onChange?: (value: string) => void;
  onBlur?: (event: FocusEvent) => void;
  onInput?: (event: InputEvent) => void;
  defaultValue?: string;
}

export const Textarea: Component<TextareaProps> = (props) => {
  const id = createUniqueId();

  const validation = useValidation({
    initialValue: props.defaultValue,
    required: props.required,
    maxLength: props.maxLength,
    customValidators: props.customValidators,
    fieldName: props.label,
  });

  const formField = useFormField({
    name: props.name,
    required: props.required,
    initialValue: props.defaultValue,
  });

  const handleInput = (event: InputEvent) => {
    const target = event.target as HTMLTextAreaElement;
    const newValue = target.value;

    validation.setValue(newValue);
    props.onChange?.(newValue);
    props.onInput?.(event);
  };

  const handleBlur = (event: FocusEvent) => {
    const target = event.target as HTMLTextAreaElement;
    const value = target.value;

    validation.markAsBlurred();
    const validationResult = validation.validate(value, true);

    if (formField.isConnectedToForm) {
      formField.updateFormField({
        isValid: validationResult.isValid,
        errorMessage: validationResult.errorMessage,
        value: value,
      });
    }

    props.onBlur?.(event);
  };

  const currentLength = () => (props.value || validation.value() || "").length;
  const isNearLimit = () =>
    props.maxLength && currentLength() > props.maxLength * 0.8;
  const isOverLimit = () =>
    props.maxLength && currentLength() > props.maxLength;

  return (
    <div class={`${styles.textareaContainer} ${props.class || ""}`}>
      <Show when={props.label}>
        <label for={id} class={styles.label}>
          {props.label}
          <Show when={props.required}>
            <span class={styles.required}>*</span>
          </Show>
        </label>
      </Show>

      <div class={styles.textareaWrapper}>
        <textarea
          id={id}
          name={props.name}
          class={`${styles.textarea} ${!validation.isValid() ? styles.textareaError : ""}`}
          placeholder={props.placeholder}
          value={props.value || validation.value()}
          rows={props.rows || 3}
          maxLength={props.maxLength}
          disabled={props.disabled}
          required={props.required}
          onInput={handleInput}
          onBlur={handleBlur}
        />
      </div>

      <Show when={props.showCharacterCount && props.maxLength}>
        <div
          class={`${styles.characterCount} ${isNearLimit() ? styles.characterCountWarning : ""} ${isOverLimit() ? styles.characterCountError : ""}`}
        >
          {currentLength()}/{props.maxLength}
        </div>
      </Show>

      <Show when={!validation.isValid() && validation.errorMessage()}>
        <div class={styles.errorMessage} role="alert">
          {validation.errorMessage()}
        </div>
      </Show>
    </div>
  );
};

