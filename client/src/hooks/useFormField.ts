import { onMount, onCleanup } from "solid-js";
import { useForm } from "@context/FormContext";

interface FieldValidation {
  isValid: boolean;
  errorMessage?: string;
  isRequired: boolean;
  value: string;
}

interface UseFormFieldOptions {
  name?: string;
  required?: boolean;
  initialValue?: string;
}

export const useFormField = (options: UseFormFieldOptions = {}) => {
  // Try to get form context, but don't require it
  let formContext: ReturnType<typeof useForm> | null = null;
  try {
    formContext = useForm();
  } catch {
    // Not inside a form context, which is fine
    formContext = null;
  }

  const fieldId = options.name || `field-${Math.random().toString(36).substring(2, 11)}`;
  const isConnectedToForm = formContext && options.name;

  // Register with form context on mount
  onMount(() => {
    if (isConnectedToForm) {
      formContext!.registerField(fieldId, {
        isValid: true,
        errorMessage: undefined,
        isRequired: options.required || false,
        value: options.initialValue || "",
      });
    }
  });

  // Unregister from form context on cleanup
  onCleanup(() => {
    if (isConnectedToForm) {
      formContext!.unregisterField(fieldId);
    }
  });

  const updateFormField = (validation: Omit<FieldValidation, 'isRequired'>) => {
    if (isConnectedToForm) {
      formContext!.updateField(fieldId, {
        ...validation,
        isRequired: options.required || false,
      });
    }
  };

  const getFormData = () => {
    return formContext?.formData || {};
  };

  return {
    isConnectedToForm: !!isConnectedToForm,
    updateFormField,
    getFormData,
    fieldId,
  };
};