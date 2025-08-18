export interface ValidationResult {
  isValid: boolean;
  errorMessage?: string;
}

export type ValidatorFunction = (value: string) => ValidationResult;

// Built-in validators
export const validators = {
  required: (fieldName: string = 'Field'): ValidatorFunction => (value: string) => {
    if (!value || value.trim().length === 0) {
      return { isValid: false, errorMessage: `${fieldName} is required` };
    }
    return { isValid: true };
  },

  minLength: (min: number, fieldName: string = 'Field'): ValidatorFunction => (value: string) => {
    if (value.length < min) {
      return { 
        isValid: false, 
        errorMessage: `${fieldName} must be at least ${min} characters` 
      };
    }
    return { isValid: true };
  },

  maxLength: (max: number, fieldName: string = 'Field'): ValidatorFunction => (value: string) => {
    if (value.length > max) {
      return { 
        isValid: false, 
        errorMessage: `${fieldName} must be no more than ${max} characters` 
      };
    }
    return { isValid: true };
  },

  email: (): ValidatorFunction => (value: string) => {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!emailRegex.test(value)) {
      return { isValid: false, errorMessage: 'Please enter a valid email address' };
    }
    return { isValid: true };
  },

  pattern: (regex: RegExp, errorMessage: string): ValidatorFunction => (value: string) => {
    if (!regex.test(value)) {
      return { isValid: false, errorMessage };
    }
    return { isValid: true };
  },

  passwordStrength: (): ValidatorFunction => (value: string) => {
    if (value.length < 8) {
      return { isValid: false, errorMessage: "Password must be at least 8 characters long" };
    }
    if (!/(?=.*[a-z])/.test(value)) {
      return { isValid: false, errorMessage: "Password must contain at least one lowercase letter" };
    }
    if (!/(?=.*[A-Z])/.test(value)) {
      return { isValid: false, errorMessage: "Password must contain at least one uppercase letter" };
    }
    if (!/(?=.*\d)/.test(value)) {
      return { isValid: false, errorMessage: "Password must contain at least one number" };
    }
    return { isValid: true };
  },
};

// Validation runner that combines multiple validators
export const runValidators = (
  value: string, 
  validatorList: ValidatorFunction[]
): ValidationResult => {
  for (const validator of validatorList) {
    const result = validator(value);
    if (!result.isValid) {
      return result;
    }
  }
  return { isValid: true };
};

// Helper to build validator chain
export class ValidationChain {
  private validators: ValidatorFunction[] = [];

  required(fieldName?: string) {
    this.validators.push(validators.required(fieldName));
    return this;
  }

  minLength(min: number, fieldName?: string) {
    this.validators.push(validators.minLength(min, fieldName));
    return this;
  }

  maxLength(max: number, fieldName?: string) {
    this.validators.push(validators.maxLength(max, fieldName));
    return this;
  }

  email() {
    this.validators.push(validators.email());
    return this;
  }

  pattern(regex: RegExp, errorMessage: string) {
    this.validators.push(validators.pattern(regex, errorMessage));
    return this;
  }

  passwordStrength() {
    this.validators.push(validators.passwordStrength());
    return this;
  }

  custom(validator: ValidatorFunction) {
    this.validators.push(validator);
    return this;
  }

  validate(value: string): ValidationResult {
    return runValidators(value, this.validators);
  }

  getValidators(): ValidatorFunction[] {
    return [...this.validators];
  }
}

// Factory function for creating validation chains
export const createValidation = () => new ValidationChain();