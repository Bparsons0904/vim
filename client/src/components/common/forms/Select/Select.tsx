import { Component, createUniqueId, Show } from "solid-js";
import { useFormField } from "@hooks/useFormField";
import { useValidation } from "@hooks/useValidation";
import { ValidatorFunction } from "../../../../utils/validation";
import styles from "./Select.module.scss";

export interface SelectOption {
  value: string;
  label: string;
  disabled?: boolean;
}

export interface SelectProps {
  name?: string;
  label?: string;
  placeholder?: string;
  options: SelectOption[];
  value?: string;
  required?: boolean;
  disabled?: boolean;
  customValidators?: ValidatorFunction[];
  class?: string;
  onChange?: (value: string) => void;
  onBlur?: (event: FocusEvent) => void;
  defaultValue?: string;
}

export const Select: Component<SelectProps> = (props) => {
  const id = createUniqueId();
  
  const validation = useValidation({
    initialValue: props.defaultValue,
    required: props.required,
    customValidators: props.customValidators,
    fieldName: props.label,
  });

  const formField = useFormField({
    name: props.name,
    required: props.required,
    initialValue: props.defaultValue,
  });

  const handleChange = (event: Event) => {
    const target = event.target as HTMLSelectElement;
    const newValue = target.value;
    
    validation.setValue(newValue);
    props.onChange?.(newValue);
  };

  const handleBlur = (event: FocusEvent) => {
    const target = event.target as HTMLSelectElement;
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

  return (
    <div class={`${styles.selectContainer} ${props.class || ""}`}>
      <Show when={props.label}>
        <label for={id} class={styles.label}>
          {props.label}
          <Show when={props.required}>
            <span class={styles.required}>*</span>
          </Show>
        </label>
      </Show>
      
      <div class={styles.selectWrapper}>
        <select
          id={id}
          name={props.name}
          class={`${styles.select} ${!validation.isValid() ? styles.selectError : ""}`}
          value={props.value || validation.value()}
          disabled={props.disabled}
          required={props.required}
          onChange={handleChange}
          onBlur={handleBlur}
        >
          <Show when={props.placeholder}>
            <option value="" disabled>
              {props.placeholder}
            </option>
          </Show>
          
          {props.options.map((option) => (
            <option 
              value={option.value} 
              disabled={option.disabled}
            >
              {option.label}
            </option>
          ))}
        </select>

        <div class={styles.selectIcon}>
          <svg
            width="16"
            height="16"
            viewBox="0 0 16 16"
            fill="none"
            xmlns="http://www.w3.org/2000/svg"
          >
            <path
              d="M4 6L8 10L12 6"
              stroke="currentColor"
              stroke-width="1.5"
              stroke-linecap="round"
              stroke-linejoin="round"
            />
          </svg>
        </div>
      </div>

      <Show when={!validation.isValid() && validation.errorMessage()}>
        <div class={styles.errorMessage} role="alert">
          {validation.errorMessage()}
        </div>
      </Show>
    </div>
  );
};