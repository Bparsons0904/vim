import { Component, createContext, useContext, JSX } from "solid-js";
import { createStore } from "solid-js/store";

interface FieldValidation {
  isValid: boolean;
  errorMessage?: string;
  isRequired: boolean;
  value: string;
}

interface FormContextValue {
  registerField: (fieldId: string, validation: FieldValidation) => void;
  unregisterField: (fieldId: string) => void;
  updateField: (fieldId: string, validation: FieldValidation) => void;
  isFormValid: () => boolean;
  getFieldValidation: (fieldId: string) => FieldValidation | undefined;
  formData: Record<string, string>;
  setFormData: (fieldId: string, value: string) => void;
}

const FormContext = createContext<FormContextValue>();

interface FormProviderProps {
  children: JSX.Element;
  onSubmit?: (formData: Record<string, string>) => void;
  class?: string;
}

export const FormProvider: Component<FormProviderProps> = (props) => {
  const [fields, setFields] = createStore<Record<string, FieldValidation>>({});
  const [formData, setFormDataStore] = createStore<Record<string, string>>({});

  const registerField = (fieldId: string, validation: FieldValidation) => {
    setFields(fieldId, validation);
    setFormDataStore(fieldId, validation.value);
  };

  const unregisterField = (fieldId: string) => {
    setFields(fieldId, undefined!);
    setFormDataStore(fieldId, undefined!);
  };

  const updateField = (fieldId: string, validation: FieldValidation) => {
    setFields(fieldId, validation);
    setFormDataStore(fieldId, validation.value);
  };

  const isFormValid = () => {
    const fieldEntries = Object.entries(fields);

    // Check if all required fields are filled and valid
    return fieldEntries.every(([, field]) => {
      if (!field) return true; // Skip undefined fields

      // If field is required, it must have a value and be valid
      if (field.isRequired) {
        return field.value.trim().length > 0 && field.isValid;
      }

      // If field is not required but has a value, it must be valid
      if (field.value.trim().length > 0) {
        return field.isValid;
      }

      // Empty non-required fields are valid
      return true;
    });
  };

  const getFieldValidation = (fieldId: string) => {
    return fields[fieldId];
  };

  const setFormData = (fieldId: string, value: string) => {
    setFormDataStore(fieldId, value);
  };

  const handleSubmit = (e: Event) => {
    e.preventDefault();
    if (props.onSubmit && isFormValid()) {
      props.onSubmit(formData);
    }
  };

  const contextValue: FormContextValue = {
    registerField,
    unregisterField,
    updateField,
    isFormValid,
    getFieldValidation,
    formData,
    setFormData,
  };

  return (
    <FormContext.Provider value={contextValue}>
      <form onSubmit={handleSubmit} class={props.class}>
        {props.children}
      </form>
    </FormContext.Provider>
  );
};

export const useForm = () => {
  const context = useContext(FormContext);
  if (!context) {
    throw new Error("useForm must be used within a FormProvider");
  }
  return context;
};

