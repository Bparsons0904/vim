import { describe, it, expect, vi } from "vitest";
import { render, screen, fireEvent } from "@solidjs/testing-library";
import { ConfirmPopup } from "./ConfirmPopup";

describe("ConfirmPopup", () => {
  it("renders with default props", () => {
    const mockOnClose = vi.fn();
    const mockOnConfirm = vi.fn();

    render(() => (
      <ConfirmPopup
        isOpen={true}
        onClose={mockOnClose}
        onConfirm={mockOnConfirm}
      />
    ));

    expect(screen.getByText("Confirm Action")).toBeInTheDocument();
    expect(screen.getByText("Are you sure you want to proceed with this action?")).toBeInTheDocument();
    expect(screen.getByText("Cancel")).toBeInTheDocument();
    expect(screen.getByText("Confirm")).toBeInTheDocument();
  });

  it("renders with custom props", () => {
    const mockOnClose = vi.fn();
    const mockOnConfirm = vi.fn();

    render(() => (
      <ConfirmPopup
        isOpen={true}
        onClose={mockOnClose}
        onConfirm={mockOnConfirm}
        title="Delete Item"
        description="This action cannot be undone."
        confirmText="Delete"
        cancelText="Keep"
        variant="danger"
      />
    ));

    expect(screen.getByText("Delete Item")).toBeInTheDocument();
    expect(screen.getByText("This action cannot be undone.")).toBeInTheDocument();
    expect(screen.getByText("Keep")).toBeInTheDocument();
    expect(screen.getByText("Delete")).toBeInTheDocument();
  });

  it("calls onConfirm when confirm button is clicked", () => {
    const mockOnClose = vi.fn();
    const mockOnConfirm = vi.fn();

    render(() => (
      <ConfirmPopup
        isOpen={true}
        onClose={mockOnClose}
        onConfirm={mockOnConfirm}
      />
    ));

    fireEvent.click(screen.getByText("Confirm"));
    expect(mockOnConfirm).toHaveBeenCalledTimes(1);
  });

  it("calls onClose when cancel button is clicked", () => {
    const mockOnClose = vi.fn();
    const mockOnConfirm = vi.fn();

    render(() => (
      <ConfirmPopup
        isOpen={true}
        onClose={mockOnClose}
        onConfirm={mockOnConfirm}
      />
    ));

    fireEvent.click(screen.getByText("Cancel"));
    expect(mockOnClose).toHaveBeenCalledTimes(1);
  });

  it("does not render when isOpen is false", () => {
    const mockOnClose = vi.fn();
    const mockOnConfirm = vi.fn();

    render(() => (
      <ConfirmPopup
        isOpen={false}
        onClose={mockOnClose}
        onConfirm={mockOnConfirm}
      />
    ));

    expect(screen.queryByText("Confirm Action")).not.toBeInTheDocument();
  });
});