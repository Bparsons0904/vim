import { Component, createSignal, createUniqueId, For, Show, createEffect } from "solid-js";
import { useFormField } from "@hooks/useFormField";
import { useValidation } from "@hooks/useValidation";
import { ValidatorFunction } from "../../../../utils/validation";
import clsx from "clsx";
import styles from "./MultiSelect.module.scss";

export interface MultiSelectOption {
  value: string;
  label: string;
  disabled?: boolean;
}

export interface MultiSelectProps {
  name?: string;
  label?: string;
  placeholder?: string;
  options: MultiSelectOption[];
  value?: string[];
  required?: boolean;
  disabled?: boolean;
  customValidators?: ValidatorFunction[];
  class?: string;
  onChange?: (values: string[]) => void;
  onBlur?: (event: FocusEvent) => void;
  defaultValue?: string[];
  maxSelections?: number;
}

export const MultiSelect: Component<MultiSelectProps> = (props) => {
  const id = createUniqueId();
  const [selectedValues, setSelectedValues] = createSignal<string[]>(
    props.value || props.defaultValue || [],
  );
  const [isOpen, setIsOpen] = createSignal(false);
  const [focusedIndex, setFocusedIndex] = createSignal(-1);

  // Sync internal state with external value prop changes
  createEffect(() => {
    if (props.value !== undefined) {
      setSelectedValues(props.value);
    }
  });

  const validation = useValidation({
    initialValue: (props.value || props.defaultValue)?.join(","),
    required: props.required,
    customValidators: props.customValidators,
    fieldName: props.label,
  });

  const formField = useFormField({
    name: props.name,
    required: props.required,
    initialValue: (props.value || props.defaultValue)?.join(","),
  });

  const handleOptionToggle = (optionValue: string) => {
    if (props.disabled) return;

    const current = selectedValues();
    const isSelected = current.includes(optionValue);

    let newValues: string[];
    if (isSelected) {
      newValues = current.filter((v) => v !== optionValue);
    } else {
      if (props.maxSelections && current.length >= props.maxSelections) {
        return;
      }
      newValues = [...current, optionValue];
    }

    setSelectedValues(newValues);
    validation.setValue(newValues.join(","));
    props.onChange?.(newValues);

    if (formField.isConnectedToForm) {
      const validationResult = validation.validate(newValues.join(","), true);
      formField.updateFormField({
        isValid: validationResult.isValid,
        errorMessage: validationResult.errorMessage,
        value: newValues.join(","),
      });
    }
  };

  const handleBlur = (event: FocusEvent) => {
    // Don't close if focus is moving to an option within the dropdown
    const relatedTarget = event.relatedTarget as HTMLElement;
    if (relatedTarget?.closest('[data-multiselect-options]')) {
      return;
    }

    setIsOpen(false);
    setFocusedIndex(-1);
    
    const values = selectedValues();
    const stringValue = values.join(",");

    validation.markAsBlurred();
    const validationResult = validation.validate(stringValue, true);

    if (formField.isConnectedToForm) {
      formField.updateFormField({
        isValid: validationResult.isValid,
        errorMessage: validationResult.errorMessage,
        value: stringValue,
      });
    }

    props.onBlur?.(event);
  };

  const handleKeyDown = (event: KeyboardEvent) => {
    if (props.disabled) return;

    const availableOptions = props.options.filter(option => !option.disabled);
    const maxIndex = availableOptions.length - 1;

    switch (event.key) {
      case 'Enter':
      case ' ':
        event.preventDefault();
        if (!isOpen()) {
          setIsOpen(true);
          setFocusedIndex(0);
        } else if (focusedIndex() >= 0 && focusedIndex() <= maxIndex) {
          const option = availableOptions[focusedIndex()];
          handleOptionToggle(option.value);
        }
        break;
        
      case 'Escape':
        event.preventDefault();
        setIsOpen(false);
        setFocusedIndex(-1);
        break;
        
      case 'ArrowDown':
        event.preventDefault();
        if (!isOpen()) {
          setIsOpen(true);
          setFocusedIndex(0);
        } else {
          setFocusedIndex(prev => Math.min(prev + 1, maxIndex));
        }
        break;
        
      case 'ArrowUp':
        event.preventDefault();
        if (isOpen()) {
          setFocusedIndex(prev => Math.max(prev - 1, 0));
        }
        break;
        
      case 'Tab':
        if (isOpen()) {
          setIsOpen(false);
          setFocusedIndex(-1);
        }
        break;
    }
  };

  const getSelectedLabels = () => {
    const selected = selectedValues();
    return props.options
      .filter((option) => selected.includes(option.value))
      .map((option) => option.label);
  };

  const getDisplayText = () => {
    const selectedLabels = getSelectedLabels();
    if (selectedLabels.length === 0) {
      return props.placeholder || "Select options...";
    }
    if (selectedLabels.length === 1) {
      return selectedLabels[0];
    }
    return `${selectedLabels.length} items selected`;
  };

  return (
    <div class={clsx(styles.multiSelectContainer, props.class)}>
      <Show when={props.label}>
        <label for={id} class={styles.label}>
          {props.label}
          <Show when={props.required}>
            <span class={styles.required}>*</span>
          </Show>
        </label>
      </Show>

      <div class={styles.multiSelectWrapper}>
        <div
          class={clsx(styles.multiSelectTrigger, {
            [styles.selectError]: !validation.isValid(),
            [styles.open]: isOpen(),
          })}
          onClick={() => !props.disabled && setIsOpen(!isOpen())}
          onBlur={handleBlur}
          onKeyDown={handleKeyDown}
          tabIndex={0}
          role="combobox"
          aria-expanded={isOpen()}
          aria-haspopup="listbox"
          aria-owns={`${id}-listbox`}
        >
          <span
            class={clsx({
              [styles.placeholder]: selectedValues().length === 0,
              [styles.selectedText]: selectedValues().length > 0,
            })}
          >
            {getDisplayText()}
          </span>

          <div class={styles.selectIcon}>
            <svg
              width="16"
              height="16"
              viewBox="0 0 16 16"
              fill="none"
              xmlns="http://www.w3.org/2000/svg"
              class={clsx({ [styles.iconRotated]: isOpen() })}
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

        <Show when={isOpen()}>
          <div 
            class={styles.optionsContainer}
            data-multiselect-options
            id={`${id}-listbox`}
            role="listbox"
            aria-multiselectable="true"
          >
            <For each={props.options}>
              {(option) => {
                const availableOptions = props.options.filter(opt => !opt.disabled);
                const availableIndex = availableOptions.findIndex(opt => opt.value === option.value);
                const isFocused = availableIndex === focusedIndex() && !option.disabled;
                
                return (
                  <div
                    class={clsx(styles.option, {
                      [styles.selected]: selectedValues().includes(option.value),
                      [styles.disabled]: option.disabled,
                      [styles.focused]: isFocused,
                    })}
                    onClick={() =>
                      !option.disabled && handleOptionToggle(option.value)
                    }
                    role="option"
                    aria-selected={selectedValues().includes(option.value)}
                    aria-disabled={option.disabled}
                    tabIndex={-1}
                  >
                    <div class={styles.checkbox}>
                      <Show when={selectedValues().includes(option.value)}>
                        <svg
                          width="12"
                          height="12"
                          viewBox="0 0 12 12"
                          fill="none"
                          xmlns="http://www.w3.org/2000/svg"
                        >
                          <path
                            d="M2 6L5 9L10 3"
                            stroke="currentColor"
                            stroke-width="2"
                            stroke-linecap="round"
                            stroke-linejoin="round"
                          />
                        </svg>
                      </Show>
                    </div>
                    <span class={styles.optionLabel}>{option.label}</span>
                  </div>
                );
              }}
            </For>
          </div>
        </Show>
      </div>

      <Show when={!validation.isValid() && validation.errorMessage()}>
        <div class={styles.errorMessage} role="alert">
          {validation.errorMessage()}
        </div>
      </Show>

      {/* Hidden input for form submission */}
      <input
        type="hidden"
        name={props.name}
        value={selectedValues().join(",")}
      />
    </div>
  );
};
