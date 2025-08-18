import { describe, it, expect, vi, afterEach } from "vitest";
import { render, screen, cleanup, fireEvent } from "@solidjs/testing-library";
import { Form } from "./Form";
import { TextInput } from "../TextInput/TextInput";
import { Button } from "../../ui/Button/Button";

describe("Form", () => {
  afterEach(() => {
    cleanup();
  });

  it("renders form element with children", () => {
    render(() => (
      <Form>
        <div data-testid="form-child">Form Content</div>
      </Form>
    ));
    
    const form = document.querySelector("form");
    const child = screen.getByTestId("form-child");
    
    expect(form).toBeTruthy();
    expect(child).toBeInTheDocument();
  });

  it("applies custom class", () => {
    render(() => (
      <Form class="custom-form-class">
        <div>Content</div>
      </Form>
    ));
    
    const form = document.querySelector("form");
    expect(form).toHaveClass("custom-form-class");
  });

  it("handles form submission with valid data", () => {
    const handleSubmit = vi.fn();
    
    render(() => (
      <Form onSubmit={handleSubmit}>
        <TextInput label="Name" name="name" required />
        <TextInput label="Email" name="email" type="email" required />
        <Button type="submit">Submit</Button>
      </Form>
    ));
    
    const form = document.querySelector("form");
    const nameInput = screen.getByLabelText(/Name/);
    const emailInput = screen.getByLabelText(/Email/);
    void screen.getByRole("button", { name: /submit/i });
    
    // Fill in valid data
    fireEvent.input(nameInput, { target: { value: "John Doe" } });
    fireEvent.blur(nameInput);
    fireEvent.input(emailInput, { target: { value: "john@example.com" } });
    fireEvent.blur(emailInput);
    
    // Submit the form
    fireEvent.submit(form);
    
    expect(handleSubmit).toHaveBeenCalledWith({
      name: "John Doe",
      email: "john@example.com"
    });
  });

  it("prevents submission with invalid data", () => {
    const handleSubmit = vi.fn();
    
    render(() => (
      <Form onSubmit={handleSubmit}>
        <TextInput label="Name" name="name" required />
        <TextInput label="Email" name="email" type="email" required />
        <Button type="submit">Submit</Button>
      </Form>
    ));
    
    const form = document.querySelector("form");
    const emailInput = screen.getByLabelText(/Email/);
    
    // Fill in invalid email
    fireEvent.input(emailInput, { target: { value: "invalid-email" } });
    fireEvent.blur(emailInput);
    
    // Submit the form
    fireEvent.submit(form);
    
    expect(handleSubmit).not.toHaveBeenCalled();
  });

  it("prevents submission when required fields are empty", () => {
    const handleSubmit = vi.fn();
    
    render(() => (
      <Form onSubmit={handleSubmit}>
        <TextInput label="Required Field" name="required" required />
        <Button type="submit">Submit</Button>
      </Form>
    ));
    
    const form = document.querySelector("form");
    
    // Submit without filling required field
    fireEvent.submit(form);
    
    expect(handleSubmit).not.toHaveBeenCalled();
  });

  it("handles form submission with optional fields", () => {
    const handleSubmit = vi.fn();
    
    render(() => (
      <Form onSubmit={handleSubmit}>
        <TextInput label="Required Field" name="required" required />
        <TextInput label="Optional Field" name="optional" />
        <Button type="submit">Submit</Button>
      </Form>
    ));
    
    const form = document.querySelector("form");
    const requiredInput = screen.getByLabelText(/Required Field/);
    const optionalInput = screen.getByLabelText(/Optional Field/);
    
    // Fill in required field only
    fireEvent.input(requiredInput, { target: { value: "Required Value" } });
    fireEvent.blur(requiredInput);
    
    // Leave optional field empty
    fireEvent.input(optionalInput, { target: { value: "" } });
    
    // Submit the form
    fireEvent.submit(form);
    
    expect(handleSubmit).toHaveBeenCalledWith({
      required: "Required Value",
      optional: ""
    });
  });

  it("validates optional fields when they have values", () => {
    const handleSubmit = vi.fn();
    
    render(() => (
      <Form onSubmit={handleSubmit}>
        <TextInput label="Required Field" name="required" required />
        <TextInput label="Optional Email" name="optionalEmail" type="email" />
        <Button type="submit">Submit</Button>
      </Form>
    ));
    
    const form = document.querySelector("form");
    const requiredInput = screen.getByLabelText(/Required Field/);
    const optionalEmailInput = screen.getByLabelText(/Optional Email/);
    
    // Fill in required field
    fireEvent.input(requiredInput, { target: { value: "Required Value" } });
    fireEvent.blur(requiredInput);
    
    // Fill in optional field with invalid email
    fireEvent.input(optionalEmailInput, { target: { value: "invalid-email" } });
    fireEvent.blur(optionalEmailInput);
    
    // Submit the form
    fireEvent.submit(form);
    
    // Should not submit due to invalid optional email
    expect(handleSubmit).not.toHaveBeenCalled();
  });

  it("works without onSubmit prop", () => {
    expect(() => {
      render(() => (
        <Form>
          <TextInput label="Name" name="name" />
          <Button type="submit">Submit</Button>
        </Form>
      ));
    }).not.toThrow();
    
    const form = document.querySelector("form");
    expect(form).toBeTruthy();
  });
});