import { describe, it, expect, vi, afterEach } from "vitest";
import { render, screen, cleanup, fireEvent } from "@solidjs/testing-library";
import { Checkbox } from "./Checkbox";

describe("Checkbox", () => {
  afterEach(() => {
    cleanup();
  });

  it("is checked when checked prop is true", () => {
    render(() => (
      <Checkbox label="Checked Checkbox" name="checked" checked={true} />
    ));

    const checkbox = screen.getByRole("checkbox") as HTMLInputElement;
    expect(checkbox.checked).toBe(true);
  });

  it("is unchecked when checked prop is false", () => {
    render(() => (
      <Checkbox label="Unchecked Checkbox" name="unchecked" checked={false} />
    ));

    const checkbox = screen.getByRole("checkbox") as HTMLInputElement;
    expect(checkbox.checked).toBe(false);
  });

  it("calls onChange when clicked", () => {
    const mockOnChange = vi.fn();

    render(() => (
      <Checkbox label="Test Checkbox" name="test" onChange={mockOnChange} />
    ));

    const checkbox = screen.getByRole("checkbox");
    fireEvent.click(checkbox);

    expect(mockOnChange).toHaveBeenCalledWith(true);
  });

  it("calls onChange when label is clicked", () => {
    const mockOnChange = vi.fn();

    render(() => (
      <Checkbox label="Test Checkbox" name="test" onChange={mockOnChange} />
    ));

    const label = screen.getByText("Test Checkbox");
    fireEvent.click(label);

    expect(mockOnChange).toHaveBeenCalledWith(true);
  });

  it("calls onBlur when focus is lost", () => {
    const mockOnBlur = vi.fn();

    render(() => (
      <Checkbox label="Test Checkbox" name="test" onBlur={mockOnBlur} />
    ));

    const checkbox = screen.getByRole("checkbox");
    fireEvent.blur(checkbox);

    expect(mockOnBlur).toHaveBeenCalled();
  });

  it("does not call onChange when disabled and clicked", () => {
    const mockOnChange = vi.fn();

    render(() => (
      <Checkbox
        label="Disabled Checkbox"
        name="disabled"
        disabled
        onChange={mockOnChange}
      />
    ));

    const label = screen.getByText("Disabled Checkbox");
    fireEvent.click(label);

    expect(mockOnChange).not.toHaveBeenCalled();
  });

  it("toggles state correctly on multiple clicks", () => {
    const mockOnChange = vi.fn();

    render(() => (
      <Checkbox label="Toggle Checkbox" name="toggle" onChange={mockOnChange} />
    ));

    const checkbox = screen.getByRole("checkbox");

    // First click - should check
    fireEvent.click(checkbox);
    expect(mockOnChange).toHaveBeenCalledWith(true);

    // Second click - should uncheck
    fireEvent.click(checkbox);
    expect(mockOnChange).toHaveBeenCalledWith(false);
  });
});

