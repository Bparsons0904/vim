import { Component, createUniqueId, JSX, Show } from "solid-js";
import { useFormField } from "@hooks/useFormField";
import { useValidation } from "@hooks/useValidation";
import { ValidatorFunction } from "../../../../utils/validation";
import styles from "./Checkbox.module.scss";

export interface CheckboxProps {
  name?: string;
  label?: string;
  checked?: boolean;
  required?: boolean;
  disabled?: boolean;
  customValidators?: ValidatorFunction[];
  class?: string;
  onChange?: (checked: boolean) => void;
  onBlur?: (event: FocusEvent) => void;
  children?: JSX.Element;
  defaultChecked?: boolean;
}

export const Checkbox: Component<CheckboxProps> = (props) => {
  const id = createUniqueId();

  const validation = useValidation({
    initialValue: props.defaultChecked ? "true" : "false",
    required: props.required,
    customValidators: props.customValidators,
    fieldName: props.label,
  });

  const formField = useFormField({
    name: props.name,
    required: props.required,
    initialValue: props.defaultChecked ? "true" : "false",
  });

  const handleChange = (event: Event) => {
    const target = event.target as HTMLInputElement;
    const newValue = target.checked;

    validation.setValue(newValue ? "true" : "false");
    props.onChange?.(newValue);
  };

  const handleBlur = (event: FocusEvent) => {
    const target = event.target as HTMLInputElement;
    const value = target.checked ? "true" : "false";

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

  const handleLabelClick = () => {
    if (!props.disabled) {
      const newValue = !props.checked;
      validation.setValue(newValue ? "true" : "false");
      props.onChange?.(newValue);
    }
  };

  return (
    <div class={`${styles.checkboxContainer} ${props.class || ""}`}>
      <div class={styles.checkboxWrapper}>
        <input
          id={id}
          type="checkbox"
          name={props.name}
          class={`${styles.checkbox} ${!validation.isValid() ? styles.checkboxError : ""}`}
          checked={props.checked || false}
          disabled={props.disabled}
          required={props.required}
          onChange={handleChange}
          onBlur={handleBlur}
        />

        <div
          class={`${styles.checkboxCustom} ${props.checked ? styles.checkboxChecked : ""} ${props.disabled ? styles.checkboxDisabled : ""}`}
          onClick={handleLabelClick}
        >
          <Show when={props.checked}>
            <svg
              class={styles.checkboxIcon}
              width="12"
              height="12"
              viewBox="0 0 12 12"
              fill="none"
              xmlns="http://www.w3.org/2000/svg"
            >
              <path
                d="M10 3L4.5 8.5L2 6"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
              />
            </svg>
          </Show>
        </div>

        <Show when={props.label || props.children}>
          <label
            for={id}
            class={`${styles.label} ${props.disabled ? styles.labelDisabled : ""}`}
            onClick={handleLabelClick}
          >
            {props.label || props.children}
            <Show when={props.required}>
              <span class={styles.required}>*</span>
            </Show>
          </label>
        </Show>
      </div>

      <Show when={!validation.isValid() && validation.errorMessage()}>
        <div class={styles.errorMessage} role="alert">
          {validation.errorMessage()}
        </div>
      </Show>
    </div>
  );
};

