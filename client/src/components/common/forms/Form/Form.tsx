import { Component, JSX } from "solid-js";
import { FormProvider } from "@context/FormContext";

interface FormProps {
  children: JSX.Element;
  onSubmit?: (formData: Record<string, unknown>) => void;
  class?: string;
}

export const Form: Component<FormProps> = (props) => {
  return (
    <FormProvider onSubmit={props.onSubmit} class={props.class}>
      {props.children}
    </FormProvider>
  );
};

