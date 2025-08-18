import { describe, it, expect, vi, afterEach } from "vitest";
import { render, screen, cleanup, fireEvent } from "@solidjs/testing-library";
import { Select, SelectOption } from "./Select";

const mockOptions: SelectOption[] = [
  { value: "", label: "Select an option..." },
  { value: "option1", label: "Option 1" },
  { value: "option2", label: "Option 2" },
  { value: "option3", label: "Option 3", disabled: true },
];

describe("Select", () => {
  afterEach(() => {
    cleanup();
  });

  it("calls onChange when selection changes", () => {
    const mockOnChange = vi.fn();

    render(() => (
      <Select
        label="Test Select"
        name="test"
        options={mockOptions}
        onChange={mockOnChange}
      />
    ));

    const select = screen.getByRole("combobox");
    fireEvent.change(select, { target: { value: "option1" } });

    expect(mockOnChange).toHaveBeenCalledWith("option1");
  });

  it("calls onBlur when focus is lost", () => {
    const mockOnBlur = vi.fn();

    render(() => (
      <Select
        label="Test Select"
        name="test"
        options={mockOptions}
        onBlur={mockOnBlur}
      />
    ));

    const select = screen.getByRole("combobox");
    fireEvent.blur(select);

    expect(mockOnBlur).toHaveBeenCalled();
  });

  it("displays current value correctly", () => {
    render(() => (
      <Select
        label="Test Select"
        name="test"
        options={mockOptions}
        value="option2"
      />
    ));

    const select = screen.getByRole("combobox") as HTMLSelectElement;
    expect(select.value).toBe("option2");
  });
});

