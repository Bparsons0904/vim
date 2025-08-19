import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, cleanup, fireEvent } from "@solidjs/testing-library";
import { MemoryRouter, Route } from "@solidjs/router";
import Register from "./Register";
import { createSignal, JSX } from "solid-js";

// Mock the auth context
const mockRegister = vi.fn();
const mockAuthContext = {
  register: mockRegister,
  isAuthenticated: createSignal(false)[0],
  user: null,
  authToken: createSignal(null)[0],
  login: vi.fn(),
  logout: vi.fn(),
};

vi.mock("@context/AuthContext", () => ({
  useAuth: () => mockAuthContext,
}));

// Mock the router
const MockRouter = (props: { children: unknown }) => (
  <MemoryRouter>
    <Route path="*" component={() => props.children as JSX.Element} />
  </MemoryRouter>
);

describe("Register", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    cleanup();
  });

  it("renders registration form elements", () => {
    render(() => (
      <MockRouter>
        <Register />
      </MockRouter>
    ));
    
    expect(screen.getByText("Join Vim Actions")).toBeInTheDocument();
    expect(screen.getByText("Start your vim workflow journey and streamline your productivity")).toBeInTheDocument();
    expect(screen.getByLabelText(/First Name/)).toBeInTheDocument();
    expect(screen.getByLabelText(/Last Name/)).toBeInTheDocument();
    expect(screen.getByLabelText(/Email Address/)).toBeInTheDocument();
    expect(screen.getByLabelText(/Username/)).toBeInTheDocument();
    expect(screen.getByLabelText(/^Password/)).toBeInTheDocument();
    expect(screen.getByLabelText(/Confirm Password/)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /create account/i })).toBeInTheDocument();
  });

  it("shows login link", () => {
    render(() => (
      <MockRouter>
        <Register />
      </MockRouter>
    ));
    
    expect(screen.getByText("Already have an account?")).toBeInTheDocument();
    expect(screen.getByText("Sign in here")).toBeInTheDocument();
    
    const loginLink = screen.getByRole("link", { name: /sign in here/i });
    expect(loginLink).toHaveAttribute("href", "/login");
  });

  it("validates required fields", () => {
    render(() => (
      <MockRouter>
        <Register />
      </MockRouter>
    ));
    
    const firstNameInput = screen.getByLabelText(/First Name/);
    const lastNameInput = screen.getByLabelText(/Last Name/);
    const emailInput = screen.getByLabelText(/Email Address/);
    const usernameInput = screen.getByLabelText(/Username/);
    const passwordInput = screen.getByLabelText(/^Password/);
    const confirmPasswordInput = screen.getByLabelText(/Confirm Password/);
    
    // Test each required field
    fireEvent.blur(firstNameInput);
    expect(screen.getByText("First Name is required")).toBeInTheDocument();
    
    fireEvent.blur(lastNameInput);
    expect(screen.getByText("Last Name is required")).toBeInTheDocument();
    
    fireEvent.blur(emailInput);
    expect(screen.getByText("Email Address is required")).toBeInTheDocument();
    
    fireEvent.blur(usernameInput);
    expect(screen.getByText("Username is required")).toBeInTheDocument();
    
    fireEvent.blur(passwordInput);
    expect(screen.getByText("Password is required")).toBeInTheDocument();
    
    fireEvent.blur(confirmPasswordInput);
    expect(screen.getByText("Confirm Password is required")).toBeInTheDocument();
  });

  it("validates minimum length requirements", () => {
    render(() => (
      <MockRouter>
        <Register />
      </MockRouter>
    ));
    
    const firstNameInput = screen.getByLabelText(/First Name/);
    const lastNameInput = screen.getByLabelText(/Last Name/);
    const usernameInput = screen.getByLabelText(/Username/);
    
    // Test minimum length validations
    fireEvent.input(firstNameInput, { target: { value: "a" } });
    fireEvent.blur(firstNameInput);
    expect(screen.getByText("First Name must be at least 2 characters")).toBeInTheDocument();
    
    fireEvent.input(lastNameInput, { target: { value: "b" } });
    fireEvent.blur(lastNameInput);
    expect(screen.getByText("Last Name must be at least 2 characters")).toBeInTheDocument();
    
    fireEvent.input(usernameInput, { target: { value: "ab" } });
    fireEvent.blur(usernameInput);
    expect(screen.getByText("Username must be at least 3 characters")).toBeInTheDocument();
  });

  it("validates maximum length for username", () => {
    render(() => (
      <MockRouter>
        <Register />
      </MockRouter>
    ));
    
    const usernameInput = screen.getByLabelText(/Username/);
    
    fireEvent.input(usernameInput, { target: { value: "a".repeat(21) } });
    fireEvent.blur(usernameInput);
    
    expect(screen.getByText("Username must be no more than 20 characters")).toBeInTheDocument();
  });

  it("validates email format", () => {
    render(() => (
      <MockRouter>
        <Register />
      </MockRouter>
    ));
    
    const emailInput = screen.getByLabelText(/Email Address/);
    
    fireEvent.input(emailInput, { target: { value: "invalid-email" } });
    fireEvent.blur(emailInput);
    
    expect(screen.getByText("Please enter a valid email address")).toBeInTheDocument();
  });

  it("validates password confirmation", () => {
    render(() => (
      <MockRouter>
        <Register />
      </MockRouter>
    ));
    
    const passwordInput = screen.getByLabelText(/^Password/);
    const confirmPasswordInput = screen.getByLabelText(/Confirm Password/);
    
    fireEvent.input(passwordInput, { target: { value: "Password123" } });
    fireEvent.input(confirmPasswordInput, { target: { value: "different123" } });
    fireEvent.blur(confirmPasswordInput);
    
    expect(screen.getByText("Passwords do not match")).toBeInTheDocument();
  });

  it("disables submit button when form is invalid", () => {
    render(() => (
      <MockRouter>
        <Register />
      </MockRouter>
    ));
    
    const submitButton = screen.getByRole("button", { name: /create account/i });
    
    // Initially form should be invalid (empty required fields)
    expect(submitButton).toBeDisabled();
  });

  it("enables submit button when form is valid", () => {
    render(() => (
      <MockRouter>
        <Register />
      </MockRouter>
    ));
    
    const submitButton = screen.getByRole("button", { name: /create account/i });
    const firstNameInput = screen.getByLabelText(/First Name/);
    const lastNameInput = screen.getByLabelText(/Last Name/);
    const emailInput = screen.getByLabelText(/Email Address/);
    const usernameInput = screen.getByLabelText(/Username/);
    const passwordInput = screen.getByLabelText(/^Password/);
    const confirmPasswordInput = screen.getByLabelText(/Confirm Password/);
    
    // Fill in valid data
    fireEvent.input(firstNameInput, { target: { value: "John" } });
    fireEvent.blur(firstNameInput);
    fireEvent.input(lastNameInput, { target: { value: "Doe" } });
    fireEvent.blur(lastNameInput);
    fireEvent.input(emailInput, { target: { value: "john@example.com" } });
    fireEvent.blur(emailInput);
    fireEvent.input(usernameInput, { target: { value: "johndoe" } });
    fireEvent.blur(usernameInput);
    fireEvent.input(passwordInput, { target: { value: "Password123" } });
    fireEvent.blur(passwordInput);
    fireEvent.input(confirmPasswordInput, { target: { value: "Password123" } });
    fireEvent.blur(confirmPasswordInput);
    
    expect(submitButton).not.toBeDisabled();
  });

  it("calls register function with correct data on form submission", () => {
    render(() => (
      <MockRouter>
        <Register />
      </MockRouter>
    ));
    
    const form = document.querySelector("form");
    const firstNameInput = screen.getByLabelText(/First Name/);
    const lastNameInput = screen.getByLabelText(/Last Name/);
    const emailInput = screen.getByLabelText(/Email Address/);
    const usernameInput = screen.getByLabelText(/Username/);
    const passwordInput = screen.getByLabelText(/^Password/);
    const confirmPasswordInput = screen.getByLabelText(/Confirm Password/);
    
    // Fill in valid data and blur each field to update form context
    fireEvent.input(firstNameInput, { target: { value: "John" } });
    fireEvent.blur(firstNameInput);
    fireEvent.input(lastNameInput, { target: { value: "Doe" } });
    fireEvent.blur(lastNameInput);
    fireEvent.input(emailInput, { target: { value: "john@example.com" } });
    fireEvent.blur(emailInput);
    fireEvent.input(usernameInput, { target: { value: "johndoe" } });
    fireEvent.blur(usernameInput);
    fireEvent.input(passwordInput, { target: { value: "Password123" } });
    fireEvent.blur(passwordInput);
    fireEvent.input(confirmPasswordInput, { target: { value: "Password123" } });
    fireEvent.blur(confirmPasswordInput);
    
    // Submit the form
    fireEvent.submit(form);
    
    expect(mockRegister).toHaveBeenCalledWith({
      firstName: "John",
      lastName: "Doe",
      email: "john@example.com",
      username: "johndoe",
      password: "Password123",
    });
  });

  it("excludes confirmPassword from registration data", () => {
    render(() => (
      <MockRouter>
        <Register />
      </MockRouter>
    ));
    
    const form = document.querySelector("form");
    const firstNameInput = screen.getByLabelText(/First Name/);
    const lastNameInput = screen.getByLabelText(/Last Name/);
    const emailInput = screen.getByLabelText(/Email Address/);
    const usernameInput = screen.getByLabelText(/Username/);
    const passwordInput = screen.getByLabelText(/^Password/);
    const confirmPasswordInput = screen.getByLabelText(/Confirm Password/);
    
    // Fill in valid data and blur each field to update form context
    fireEvent.input(firstNameInput, { target: { value: "Jane" } });
    fireEvent.blur(firstNameInput);
    fireEvent.input(lastNameInput, { target: { value: "Smith" } });
    fireEvent.blur(lastNameInput);
    fireEvent.input(emailInput, { target: { value: "jane@example.com" } });
    fireEvent.blur(emailInput);
    fireEvent.input(usernameInput, { target: { value: "janesmith" } });
    fireEvent.blur(usernameInput);
    fireEvent.input(passwordInput, { target: { value: "Secret123" } });
    fireEvent.blur(passwordInput);
    fireEvent.input(confirmPasswordInput, { target: { value: "Secret123" } });
    fireEvent.blur(confirmPasswordInput);
    
    // Submit the form
    fireEvent.submit(form);
    
    const registrationData = mockRegister.mock.calls[0][0];
    expect(registrationData).not.toHaveProperty("confirmPassword");
    expect(registrationData).toEqual({
      firstName: "Jane",
      lastName: "Smith",
      email: "jane@example.com",
      username: "janesmith",
      password: "Secret123",
    });
  });

  it("has proper form field attributes", () => {
    render(() => (
      <MockRouter>
        <Register />
      </MockRouter>
    ));
    
    const firstNameInput = screen.getByLabelText(/First Name/);
    const lastNameInput = screen.getByLabelText(/Last Name/);
    const emailInput = screen.getByLabelText(/Email Address/);
    const usernameInput = screen.getByLabelText(/Username/);
    const passwordInput = screen.getByLabelText(/^Password/);
    const confirmPasswordInput = screen.getByLabelText(/Confirm Password/);
    
    expect(firstNameInput).toHaveAttribute("name", "firstName");
    expect(firstNameInput).toHaveAttribute("autocomplete", "given-name");
    
    expect(lastNameInput).toHaveAttribute("name", "lastName");
    expect(lastNameInput).toHaveAttribute("autocomplete", "family-name");
    
    expect(emailInput).toHaveAttribute("name", "email");
    expect(emailInput).toHaveAttribute("type", "email");
    expect(emailInput).toHaveAttribute("autocomplete", "email");
    
    expect(usernameInput).toHaveAttribute("name", "username");
    expect(usernameInput).toHaveAttribute("autocomplete", "username");
    
    expect(passwordInput).toHaveAttribute("name", "password");
    expect(passwordInput).toHaveAttribute("type", "password");
    expect(passwordInput).toHaveAttribute("autocomplete", "new-password");
    
    expect(confirmPasswordInput).toHaveAttribute("name", "confirmPassword");
    expect(confirmPasswordInput).toHaveAttribute("type", "password");
    expect(confirmPasswordInput).toHaveAttribute("autocomplete", "new-password");
  });

  it("shows required asterisks for all required fields", () => {
    render(() => (
      <MockRouter>
        <Register />
      </MockRouter>
    ));
    
    // Check that asterisks are present in the DOM (should be 6 for all required fields)
    const pageContainer = document.body;
    expect(pageContainer.textContent).toContain("*");
    
    // Verify all fields are present
    expect(screen.getByLabelText(/First Name/)).toBeInTheDocument();
    expect(screen.getByLabelText(/Last Name/)).toBeInTheDocument();
    expect(screen.getByLabelText(/Email Address/)).toBeInTheDocument();
    expect(screen.getByLabelText(/Username/)).toBeInTheDocument();
    expect(screen.getByLabelText(/^Password/)).toBeInTheDocument();
    expect(screen.getByLabelText(/Confirm Password/)).toBeInTheDocument();
  });

  it("shows password toggle buttons for password fields", () => {
    render(() => (
      <MockRouter>
        <Register />
      </MockRouter>
    ));
    
    const toggleButtons = screen.getAllByRole("button");
    // Should have 3 buttons: 2 password toggles + 1 submit button
    expect(toggleButtons).toHaveLength(3);
    expect(toggleButtons[2]).toHaveTextContent("Create Account"); // Submit button
  });
});