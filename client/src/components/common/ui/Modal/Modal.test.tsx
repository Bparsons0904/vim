import { describe, it, expect, vi, afterEach } from "vitest";
import { render, screen, cleanup, fireEvent } from "@solidjs/testing-library";
import { Modal, ModalSize } from "./Modal";

describe("Modal", () => {
  afterEach(() => {
    cleanup();
  });

  it("renders when isOpen is true", () => {
    const mockOnClose = vi.fn();
    
    render(() => (
      <Modal isOpen={true} onClose={mockOnClose}>
        <div>Modal Content</div>
      </Modal>
    ));
    
    expect(screen.getByRole("dialog")).toBeInTheDocument();
    expect(screen.getByText("Modal Content")).toBeInTheDocument();
  });

  it("does not render when isOpen is false", () => {
    const mockOnClose = vi.fn();
    
    render(() => (
      <Modal isOpen={false} onClose={mockOnClose}>
        <div>Modal Content</div>
      </Modal>
    ));
    
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    expect(screen.queryByText("Modal Content")).not.toBeInTheDocument();
  });

  it("renders title when provided", () => {
    const mockOnClose = vi.fn();
    
    render(() => (
      <Modal isOpen={true} onClose={mockOnClose} title="Test Modal">
        <div>Modal Content</div>
      </Modal>
    ));
    
    expect(screen.getByText("Test Modal")).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Test Modal" })).toBeInTheDocument();
  });

  it("renders close button by default", () => {
    const mockOnClose = vi.fn();
    
    render(() => (
      <Modal isOpen={true} onClose={mockOnClose}>
        <div>Modal Content</div>
      </Modal>
    ));
    
    const closeButton = screen.getByRole("button", { name: "Close modal" });
    expect(closeButton).toBeInTheDocument();
  });

  it("calls onClose when close button is clicked", () => {
    const mockOnClose = vi.fn();
    
    render(() => (
      <Modal isOpen={true} onClose={mockOnClose}>
        <div>Modal Content</div>
      </Modal>
    ));
    
    const closeButton = screen.getByRole("button", { name: "Close modal" });
    fireEvent.click(closeButton);
    
    expect(mockOnClose).toHaveBeenCalledTimes(1);
  });

  it("calls onClose when backdrop is clicked", () => {
    const mockOnClose = vi.fn();
    
    render(() => (
      <Modal isOpen={true} onClose={mockOnClose}>
        <div>Modal Content</div>
      </Modal>
    ));
    
    const backdrop = screen.getByRole("dialog");
    fireEvent.click(backdrop);
    
    expect(mockOnClose).toHaveBeenCalledTimes(1);
  });

  it("does not call onClose when modal content is clicked", () => {
    const mockOnClose = vi.fn();
    
    render(() => (
      <Modal isOpen={true} onClose={mockOnClose}>
        <div>Modal Content</div>
      </Modal>
    ));
    
    const content = screen.getByText("Modal Content");
    fireEvent.click(content);
    
    expect(mockOnClose).not.toHaveBeenCalled();
  });

  it("hides close button when showCloseButton is false", () => {
    const mockOnClose = vi.fn();
    
    render(() => (
      <Modal isOpen={true} onClose={mockOnClose} showCloseButton={false}>
        <div>Modal Content</div>
      </Modal>
    ));
    
    expect(screen.queryByRole("button", { name: "Close modal" })).not.toBeInTheDocument();
  });

  it("renders with different modal sizes", () => {
    const mockOnClose = vi.fn();
    
    const { unmount } = render(() => (
      <Modal isOpen={true} onClose={mockOnClose} size={ModalSize.Small}>
        <div>Small Modal</div>
      </Modal>
    ));
    
    expect(screen.getByText("Small Modal")).toBeInTheDocument();
    
    unmount();
    
    render(() => (
      <Modal isOpen={true} onClose={mockOnClose} size={ModalSize.Large}>
        <div>Large Modal</div>
      </Modal>
    ));
    
    expect(screen.getByText("Large Modal")).toBeInTheDocument();
  });

  it("does not call onClose when closeOnBackdropClick is false", () => {
    const mockOnClose = vi.fn();
    
    render(() => (
      <Modal isOpen={true} onClose={mockOnClose} closeOnBackdropClick={false}>
        <div>Modal Content</div>
      </Modal>
    ));
    
    const backdrop = screen.getByRole("dialog");
    fireEvent.click(backdrop);
    
    expect(mockOnClose).not.toHaveBeenCalled();
  });

  it("has proper accessibility attributes", () => {
    const mockOnClose = vi.fn();
    
    render(() => (
      <Modal isOpen={true} onClose={mockOnClose} title="Accessible Modal">
        <div>Modal Content</div>
      </Modal>
    ));
    
    const dialog = screen.getByRole("dialog");
    expect(dialog).toHaveAttribute("aria-modal", "true");
    expect(dialog).toHaveAttribute("aria-labelledby", "modal-title");
    
    const title = screen.getByText("Accessible Modal");
    expect(title).toHaveAttribute("id", "modal-title");
  });
});