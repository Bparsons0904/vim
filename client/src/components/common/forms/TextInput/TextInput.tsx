import { Component, createSignal, Show } from "solid-js";
import clsx from "clsx";
import styles from "./TextInput.module.scss";
import { useValidation } from "@hooks/useValidation";
import { useFormField } from "@hooks/useFormField";
import { PasswordToggle } from "../PasswordToggle/PasswordToggle";
import { ValidatorFunction } from "../../../../utils/validation";

type TextInputType = "text" | "password" | "email";

interface TextInputProps {
  label?: string;
  defaultValue?: string;
  value?: string;
  placeholder?: string;
  onBlur?: (
    value: string,
    target: HTMLInputElement,
    event: FocusEvent & { target: HTMLInputElement },
  ) => void;
  onInput?: (value: string, event: InputEvent & { target: HTMLInputElement }) => void;
  type?: TextInputType;
  autoComplete?: string;
  onInvalid?: (event: Event & { target: HTMLInputElement }) => void;
  validationFunction?: ValidatorFunction;
  required?: boolean;
  minLength?: number;
  maxLength?: number;
  pattern?: string;
  customValidators?: ValidatorFunction[];
  name?: string;
  class?: string;
}

export const TextInput: Component<TextInputProps> = (props) => {
  const [showPassword, setShowPassword] = createSignal(false);
  
  // Generate unique ID for input-label association
  const inputId = `input-${props.name || Math.random().toString(36).substr(2, 9)}`;
  
  // Build custom validators including email validation
  const buildCustomValidators = () => {
    const validators = [];
    
    // Add email validation if type is email
    if (props.type === "email") {
      const emailValidator: ValidatorFunction = (value: string) => {
        const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
        if (value && !emailRegex.test(value)) {
          return { isValid: false, errorMessage: 'Please enter a valid email address' };
        }
        return { isValid: true };
      };
      validators.push(emailValidator);
    }
    
    // Add custom validation function
    if (props.validationFunction) {
      validators.push(props.validationFunction);
    }
    
    // Add any additional custom validators
    if (props.customValidators) {
      validators.push(...props.customValidators);
    }
    
    return validators;
  };

  // Determine if this is a controlled component
  const isControlled = () => props.value !== undefined;
  
  // Use custom hooks
  const validation = useValidation({
    initialValue: props.value ?? props.defaultValue,
    required: props.required,
    minLength: props.minLength,
    maxLength: props.maxLength,
    pattern: props.pattern,
    customValidators: buildCustomValidators(),
    fieldName: props.label,
  });

  const formField = useFormField({
    name: props.name,
    required: props.required,
    initialValue: props.defaultValue,
  });

  const isPasswordField = () => props.type === "password";
  
  const getInputType = () => {
    if (isPasswordField() && showPassword()) {
      return "text";
    }
    return props.type ?? "text";
  };


  const handleUpdate = (event: FocusEvent & { target: HTMLInputElement }) => {
    const value = event.target.value;
    
    // Mark field as blurred and force validation
    validation.markAsBlurred();
    const validationResult = validation.validate(value, true);

    // Update form context if connected
    if (formField.isConnectedToForm) {
      formField.updateFormField({
        isValid: validationResult.isValid,
        errorMessage: validationResult.errorMessage,
        value: value,
      });
    }

    if (!validationResult.isValid) {
      props.onInvalid?.(event);
      return;
    }

    props.onBlur?.(value, event.target, event);
  };

  const togglePasswordVisibility = () => {
    setShowPassword(!showPassword());
  };

  return (
    <div class={styles.textInput}>
      <label class={styles.label} for={inputId}>
        <span class={styles.labelText}>
          {props.label} <Show when={props.required}>*</Show>
        </span>
      </label>
      <div class={styles.inputWrapper}>
        <div class={styles.inputContainer}>
          <input
            id={inputId}
            type={getInputType()}
            class={clsx(
              styles.input,
              isPasswordField() && styles.passwordInput,
              !validation.isValid() && styles.inputError,
              props.class
            )}
            value={isControlled() ? props.value : validation.value()}
            placeholder={props.placeholder}
            onInput={(e) => {
              const value = e.target.value;
              if (isControlled()) {
                props.onInput?.(value, e);
              } else {
                validation.setValue(value);
              }
            }}
            onBlur={handleUpdate}
            autocomplete={props.autoComplete ?? "off"}
            name={props.name}
          />
          <Show when={isPasswordField()}>
            <PasswordToggle 
              showPassword={showPassword()} 
              onToggle={togglePasswordVisibility} 
            />
          </Show>
        </div>
        <Show when={!validation.isValid() && validation.errorMessage()}>
          <span class={styles.error}>{validation.errorMessage()}</span>
        </Show>
      </div>
    </div>
  );
};