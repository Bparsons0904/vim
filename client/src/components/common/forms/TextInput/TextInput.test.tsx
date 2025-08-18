import { describe, it, expect, vi, afterEach } from "vitest";
import { render, screen, cleanup, fireEvent } from "@solidjs/testing-library";
import { TextInput } from "./TextInput";
import { FormProvider } from "@context/FormContext";

describe("TextInput", () => {
  afterEach(() => {
    cleanup();
  });

  it("renders with basic props", () => {
    render(() => <TextInput label="Test Input" name="test" />);
    
    const input = screen.getByRole("textbox");
    const label = screen.getByText("Test Input");
    
    expect(input).toBeInTheDocument();
    expect(label).toBeInTheDocument();
    expect(input).toHaveAttribute("name", "test");
  });

  it("shows required asterisk when required", () => {
    render(() => <TextInput label="Required Field" name="required" required />);
    
    // Check that the label contains both the text and asterisk
    const labelElement = screen.getByLabelText("Required Field *");
    expect(labelElement).toBeInTheDocument();
  });

  it("handles different input types", () => {
    const { unmount } = render(() => <TextInput label="Email" name="email" type="email" />);
    expect(screen.getByRole("textbox")).toHaveAttribute("type", "email");
    unmount();

    render(() => <TextInput label="Password" name="password" type="password" />);
    expect(screen.getByDisplayValue("")).toHaveAttribute("type", "password");
  });

  it("shows password toggle for password inputs", () => {
    render(() => <TextInput label="Password" name="password" type="password" />);
    
    const toggleButton = screen.getByRole("button");
    expect(toggleButton).toBeInTheDocument();
  });

  it("toggles password visibility", () => {
    render(() => <TextInput label="Password" name="password" type="password" />);
    
    const input = screen.getByDisplayValue("");
    const toggleButton = screen.getByRole("button");
    
    expect(input).toHaveAttribute("type", "password");
    
    fireEvent.click(toggleButton);
    expect(input).toHaveAttribute("type", "text");
    
    fireEvent.click(toggleButton);
    expect(input).toHaveAttribute("type", "password");
  });

  it("validates required fields on blur", () => {
    render(() => <TextInput label="Required Field" name="required" required />);
    
    const input = screen.getByRole("textbox");
    
    // Initially no error should be shown
    expect(screen.queryByText("Required Field is required")).not.toBeInTheDocument();
    
    // Blur without entering text should show error
    fireEvent.blur(input);
    expect(screen.getByText("Required Field is required")).toBeInTheDocument();
  });

  it("validates email format", () => {
    render(() => <TextInput label="Email" name="email" type="email" required />);
    
    const input = screen.getByRole("textbox");
    
    // Enter invalid email and blur
    fireEvent.input(input, { target: { value: "invalid-email" } });
    fireEvent.blur(input);
    
    expect(screen.getByText("Please enter a valid email address")).toBeInTheDocument();
  });

  it("validates minimum length", () => {
    render(() => <TextInput label="Username" name="username" minLength={3} required />);
    
    const input = screen.getByRole("textbox");
    
    // Enter text shorter than minimum and blur
    fireEvent.input(input, { target: { value: "ab" } });
    fireEvent.blur(input);
    
    expect(screen.getByText("Username must be at least 3 characters")).toBeInTheDocument();
  });

  it("validates maximum length", () => {
    render(() => <TextInput label="Code" name="code" maxLength={5} required />);
    
    const input = screen.getByRole("textbox");
    
    // Enter text longer than maximum and blur
    fireEvent.input(input, { target: { value: "toolong" } });
    fireEvent.blur(input);
    
    expect(screen.getByText("Code must be no more than 5 characters")).toBeInTheDocument();
  });

  it("calls onBlur callback", () => {
    const handleBlur = vi.fn();
    render(() => <TextInput label="Test" name="test" onBlur={handleBlur} />);
    
    const input = screen.getByRole("textbox");
    fireEvent.input(input, { target: { value: "test value" } });
    fireEvent.blur(input);
    
    expect(handleBlur).toHaveBeenCalledWith("test value", input, expect.any(Object));
  });

  it("integrates with form context", () => {
    const TestForm = () => (
      <FormProvider>
        <form>
          <TextInput label="Form Field" name="formField" required />
        </form>
      </FormProvider>
    );

    render(() => <TestForm />);
    
    const input = screen.getByRole("textbox");
    expect(input).toBeInTheDocument();
    expect(input).toHaveAttribute("name", "formField");
  });

  it("handles custom validation function", () => {
    const customValidator = (value: string) => {
      if (value === "forbidden") {
        return { isValid: false, errorMessage: "This value is not allowed" };
      }
      return { isValid: true };
    };

    render(() => (
      <TextInput 
        label="Custom" 
        name="custom" 
        validationFunction={customValidator}
        required 
      />
    ));
    
    const input = screen.getByRole("textbox");
    
    fireEvent.input(input, { target: { value: "forbidden" } });
    fireEvent.blur(input);
    
    expect(screen.getByText("This value is not allowed")).toBeInTheDocument();
  });
});