import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@solidjs/testing-library";
import { QueryClient, QueryClientProvider } from "@tanstack/solid-query";
import { JSX } from "solid-js";
import { CreateStoryForm } from "./CreateStoryForm";

// Mock the API service
vi.mock("@services/api/api.service", () => ({
  getApi: vi.fn().mockResolvedValue({
    data: {
      genres: [
        { id: 1, name: "Fantasy", code: "fantasy" },
        { id: 2, name: "Science Fiction", code: "sci-fi" },
        { id: 3, name: "Mystery", code: "mystery" },
        { id: 4, name: "Romance", code: "romance" },
        { id: 5, name: "Horror", code: "horror" },
        { id: 6, name: "Adventure", code: "adventure" },
      ],
      orderTypes: [
        {
          id: 1,
          name: "Round Robin",
          code: "round-robin",
          description: "Authors take turns",
        },
        {
          id: 2,
          name: "Random",
          code: "random",
          description: "Authors selected randomly",
        },
      ],
    },
  }),
}));

describe("CreateStoryForm", () => {
  const mockOnClose = vi.fn();
  const mockOnSubmit = vi.fn();

  const defaultProps = {
    isOpen: true,
    onClose: mockOnClose,
    onSubmit: mockOnSubmit,
  };

  // Helper function to render component with QueryClient
  const renderWithQueryClient = (component: () => JSX.Element) => {
    const queryClient = new QueryClient({
      defaultOptions: {
        queries: {
          retry: false, // Disable retries for tests
        },
      },
    });

    return render(() => (
      <QueryClientProvider client={queryClient}>
        {component()}
      </QueryClientProvider>
    ));
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    // No cleanup needed, it's handled automatically
  });

  it("renders when isOpen is true", () => {
    renderWithQueryClient(() => <CreateStoryForm {...defaultProps} />);

    expect(screen.getByRole("dialog")).toBeInTheDocument();
    expect(screen.getByText("Create New Story")).toBeInTheDocument();
  });

  it("does not render when isOpen is false", () => {
    renderWithQueryClient(() => (
      <CreateStoryForm {...defaultProps} isOpen={false} />
    ));

    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
  });

  it("renders all form fields", () => {
    renderWithQueryClient(() => <CreateStoryForm {...defaultProps} />);

    // Check for all form fields
    expect(screen.getByLabelText(/Story Title/)).toBeInTheDocument();
    expect(screen.getByLabelText(/Description/)).toBeInTheDocument();
    expect(screen.getByLabelText(/Genre/)).toBeInTheDocument();
    expect(
      screen.getByText(/Allow other participants to invite new members/),
    ).toBeInTheDocument();
  });

  it("renders all genre options", async () => {
    renderWithQueryClient(() => <CreateStoryForm {...defaultProps} />);

    const genreSelect = screen.getByLabelText(/Genre/) as HTMLSelectElement;

    // Wait for the options to load
    await waitFor(() => {
      const options = genreSelect.querySelectorAll("option");
      expect(options).toHaveLength(7); // placeholder + 6 genres
    });

    const options = genreSelect.querySelectorAll("option");
    expect(options[0]).toHaveTextContent("Select a genre...");
    expect(options[1]).toHaveTextContent("Fantasy");
    expect(options[2]).toHaveTextContent("Science Fiction");
  });

  it("shows character counter for description", () => {
    renderWithQueryClient(() => <CreateStoryForm {...defaultProps} />);

    expect(screen.getByText("0/500")).toBeInTheDocument();
  });

  it("updates character counter when typing in description", () => {
    renderWithQueryClient(() => <CreateStoryForm {...defaultProps} />);

    const description = screen.getByLabelText(/Description/);
    fireEvent.input(description, { target: { value: "Hello world" } });

    expect(screen.getByText("11/500")).toBeInTheDocument();
  });

  it("checkbox is checked by default", () => {
    renderWithQueryClient(() => <CreateStoryForm {...defaultProps} />);

    const checkbox = screen.getByRole("checkbox");
    expect(checkbox).toBeInTheDocument();
    expect(checkbox).toHaveProperty("checked", true);
  });

  it("submit button is disabled when form is invalid", () => {
    renderWithQueryClient(() => <CreateStoryForm {...defaultProps} />);

    const submitButton = screen.getByRole("button", { name: "Create Story" });
    expect(submitButton).toBeDisabled();
  });

  it("calls onClose when cancel button is clicked", () => {
    renderWithQueryClient(() => <CreateStoryForm {...defaultProps} />);

    const cancelButton = screen.getByRole("button", { name: "Cancel" });
    fireEvent.click(cancelButton);

    expect(mockOnClose).toHaveBeenCalled();
  });

  it("calls onClose when close button is clicked", () => {
    renderWithQueryClient(() => <CreateStoryForm {...defaultProps} />);

    const closeButton = screen.getByRole("button", { name: "Close modal" });
    fireEvent.click(closeButton);

    expect(mockOnClose).toHaveBeenCalled();
  });

  it("renders form buttons", () => {
    renderWithQueryClient(() => <CreateStoryForm {...defaultProps} />);

    expect(screen.getByRole("button", { name: "Cancel" })).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: "Create Story" }),
    ).toBeInTheDocument();
  });

  it("has correct modal title", () => {
    renderWithQueryClient(() => <CreateStoryForm {...defaultProps} />);

    expect(screen.getByText("Create New Story")).toBeInTheDocument();
  });

  it("renders with medium modal size", () => {
    renderWithQueryClient(() => <CreateStoryForm {...defaultProps} />);

    const modal = screen.getByRole("dialog");
    expect(modal.firstElementChild).toHaveClass("_modalMedium_bd140b");
  });
});

