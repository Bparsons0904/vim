import { createSignal } from "solid-js";
import { ValidationResult, ValidatorFunction, runValidators } from "../utils/validation";

interface UseValidationOptions {
  initialValue?: string;
  required?: boolean;
  minLength?: number;
  maxLength?: number;
  pattern?: string;
  customValidators?: ValidatorFunction[];
  fieldName?: string;
}

export const useValidation = (options: UseValidationOptions = {}) => {
  const [value, setValue] = createSignal(options.initialValue || "");
  const [isValid, setIsValid] = createSignal(true);
  const [errorMessage, setErrorMessage] = createSignal<string>("");
  const [hasBlurred, setHasBlurred] = createSignal(false);

  const buildValidators = (): ValidatorFunction[] => {
    const validators: ValidatorFunction[] = [];
    const fieldName = options.fieldName || 'Field';

    // Required validation
    if (options.required) {
      validators.push((val: string) => {
        if (!val || val.trim().length === 0) {
          return { isValid: false, errorMessage: `${fieldName} is required` };
        }
        return { isValid: true };
      });
    }

    // Skip other validations if field is empty and not required
    const nonEmptyValidators: ValidatorFunction[] = [];

    // Min length validation
    if (options.minLength !== undefined) {
      nonEmptyValidators.push((val: string) => {
        if (val.length < options.minLength!) {
          return { 
            isValid: false, 
            errorMessage: `${fieldName} must be at least ${options.minLength} characters` 
          };
        }
        return { isValid: true };
      });
    }

    // Max length validation
    if (options.maxLength !== undefined) {
      nonEmptyValidators.push((val: string) => {
        if (val.length > options.maxLength!) {
          return { 
            isValid: false, 
            errorMessage: `${fieldName} must be no more than ${options.maxLength} characters` 
          };
        }
        return { isValid: true };
      });
    }

    // Pattern validation
    if (options.pattern) {
      nonEmptyValidators.push((val: string) => {
        const regex = new RegExp(options.pattern!);
        if (!regex.test(val)) {
          return { 
            isValid: false, 
            errorMessage: `${fieldName} format is invalid` 
          };
        }
        return { isValid: true };
      });
    }

    // Add custom validators
    if (options.customValidators) {
      nonEmptyValidators.push(...options.customValidators);
    }

    // Combine validators with empty check logic
    if (nonEmptyValidators.length > 0) {
      validators.push((val: string) => {
        // Skip other validations if field is empty and not required
        if ((!val || val.trim().length === 0) && !options.required) {
          return { isValid: true };
        }
        
        return runValidators(val, nonEmptyValidators);
      });
    }

    return validators;
  };

  const validate = (newValue: string, forceValidation = false): ValidationResult => {
    const validators = buildValidators();
    
    // Only show validation errors after blur or if forced
    if (!hasBlurred() && !forceValidation) {
      // Still run validation for form validity checks, but don't show errors
      const result = runValidators(newValue, validators);
      return result;
    }
    
    const result = runValidators(newValue, validators);
    
    setIsValid(result.isValid);
    setErrorMessage(result.errorMessage || "");
    
    return result;
  };

  const updateValue = (newValue: string) => {
    setValue(newValue);
    return validate(newValue);
  };

  const getValidationState = () => ({
    isValid: isValid(),
    errorMessage: errorMessage(),
    value: value(),
  });

  const markAsBlurred = () => {
    setHasBlurred(true);
  };

  return {
    value,
    setValue: updateValue,
    isValid,
    errorMessage,
    validate,
    getValidationState,
    markAsBlurred,
  };
};