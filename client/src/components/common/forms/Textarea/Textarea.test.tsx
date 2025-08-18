import { describe, it, expect, vi, afterEach } from "vitest";
import { render, screen, cleanup, fireEvent } from "@solidjs/testing-library";
import { Textarea } from "./Textarea";

describe("Textarea", () => {
  afterEach(() => {
    cleanup();
  });

  it("calls onChange when text is input", () => {
    const mockOnChange = vi.fn();

    render(() => (
      <Textarea label="Test Textarea" name="test" onChange={mockOnChange} />
    ));

    const textarea = screen.getByRole("textbox");
    fireEvent.input(textarea, { target: { value: "Hello world" } });

    expect(mockOnChange).toHaveBeenCalledWith("Hello world");
  });

  it("calls onInput when text is input", () => {
    const mockOnInput = vi.fn();

    render(() => (
      <Textarea label="Test Textarea" name="test" onInput={mockOnInput} />
    ));

    const textarea = screen.getByRole("textbox");
    fireEvent.input(textarea, { target: { value: "Hello" } });

    expect(mockOnInput).toHaveBeenCalled();
  });

  it("calls onBlur when focus is lost", () => {
    const mockOnBlur = vi.fn();

    render(() => (
      <Textarea label="Test Textarea" name="test" onBlur={mockOnBlur} />
    ));

    const textarea = screen.getByRole("textbox");
    fireEvent.blur(textarea);

    expect(mockOnBlur).toHaveBeenCalled();
  });

  it("displays current value correctly", () => {
    render(() => (
      <Textarea label="Test Textarea" name="test" value="Current value" />
    ));

    const textarea = screen.getByRole("textbox") as HTMLTextAreaElement;
    expect(textarea.value).toBe("Current value");
  });
});

