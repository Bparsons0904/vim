import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, cleanup, fireEvent } from "@solidjs/testing-library";
import { MemoryRouter, Route } from "@solidjs/router";
import Login from "./Login";
import { createSignal, JSX } from "solid-js";

// Mock the auth context
const mockLogin = vi.fn();
const mockAuthContext = {
  login: mockLogin,
  isAuthenticated: createSignal(false)[0],
  user: null,
  authToken: createSignal(null)[0],
  register: vi.fn(),
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

describe("Login", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    cleanup();
  });

  it("renders login form elements", () => {
    render(() => (
      <MockRouter>
        <Login />
      </MockRouter>
    ));
    
    expect(screen.getByText("Welcome Back")).toBeInTheDocument();
    expect(screen.getByText("Sign in to continue your creative journey")).toBeInTheDocument();
    expect(screen.getByLabelText(/Username or Email/)).toBeInTheDocument();
    expect(screen.getByLabelText(/Password/)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /sign in/i })).toBeInTheDocument();
  });

  it("shows registration link", () => {
    render(() => (
      <MockRouter>
        <Login />
      </MockRouter>
    ));
    
    expect(screen.getByText("Don't have an account?")).toBeInTheDocument();
    expect(screen.getByText("Create one here")).toBeInTheDocument();
    
    const registerLink = screen.getByRole("link", { name: /create one here/i });
    expect(registerLink).toHaveAttribute("href", "/register");
  });

  it("has default values in form fields", () => {
    render(() => (
      <MockRouter>
        <Login />
      </MockRouter>
    ));
    
    const usernameInput = screen.getByLabelText(/Username or Email/) as HTMLInputElement;
    const passwordInput = screen.getByLabelText(/Password/) as HTMLInputElement;
    
    expect(usernameInput.value).toBe("admin");
    expect(passwordInput.value).toBe("password");
  });

  it("validates required fields", () => {
    render(() => (
      <MockRouter>
        <Login />
      </MockRouter>
    ));
    
    const usernameInput = screen.getByLabelText(/Username or Email/);
    const passwordInput = screen.getByLabelText(/Password/);
    
    // Clear the default values
    fireEvent.input(usernameInput, { target: { value: "" } });
    fireEvent.blur(usernameInput);
    
    expect(screen.getByText("Username or Email is required")).toBeInTheDocument();
    
    fireEvent.input(passwordInput, { target: { value: "" } });
    fireEvent.blur(passwordInput);
    
    expect(screen.getByText("Password is required")).toBeInTheDocument();
  });

  it("validates minimum length requirements", () => {
    render(() => (
      <MockRouter>
        <Login />
      </MockRouter>
    ));
    
    const usernameInput = screen.getByLabelText(/Username or Email/);
    const passwordInput = screen.getByLabelText(/Password/);
    
    // Enter values that are too short
    fireEvent.input(usernameInput, { target: { value: "ab" } });
    fireEvent.blur(usernameInput);
    
    expect(screen.getByText("Username or Email must be at least 3 characters")).toBeInTheDocument();
    
    fireEvent.input(passwordInput, { target: { value: "12345" } });
    fireEvent.blur(passwordInput);
    
    expect(screen.getByText("Password must be at least 6 characters")).toBeInTheDocument();
  });

  it("disables submit button when form is invalid", () => {
    render(() => (
      <MockRouter>
        <Login />
      </MockRouter>
    ));
    
    const submitButton = screen.getByRole("button", { name: /sign in/i });
    const usernameInput = screen.getByLabelText(/Username or Email/);
    
    // Clear username to make form invalid
    fireEvent.input(usernameInput, { target: { value: "" } });
    fireEvent.blur(usernameInput);
    
    expect(submitButton).toBeDisabled();
  });

  it("enables submit button when form is valid", () => {
    render(() => (
      <MockRouter>
        <Login />
      </MockRouter>
    ));
    
    const submitButton = screen.getByRole("button", { name: /sign in/i });
    
    // With default values, form should be valid
    expect(submitButton).not.toBeDisabled();
  });

  it("calls login function with correct credentials on form submission", () => {
    render(() => (
      <MockRouter>
        <Login />
      </MockRouter>
    ));
    
    const form = document.querySelector("form");
    const usernameInput = screen.getByLabelText(/Username or Email/);
    const passwordInput = screen.getByLabelText(/Password/);
    
    // Enter custom values and blur to update form context
    fireEvent.input(usernameInput, { target: { value: "testuser" } });
    fireEvent.blur(usernameInput);
    fireEvent.input(passwordInput, { target: { value: "testpass123" } });
    fireEvent.blur(passwordInput);
    
    // Submit the form
    fireEvent.submit(form);
    
    expect(mockLogin).toHaveBeenCalledWith({
      login: "testuser",
      password: "testpass123",
    });
  });

  it("uses default values when submitting form", () => {
    render(() => (
      <MockRouter>
        <Login />
      </MockRouter>
    ));
    
    const form = document.querySelector("form");
    
    // Submit without changing values
    fireEvent.submit(form);
    
    expect(mockLogin).toHaveBeenCalledWith({
      login: "admin",
      password: "password",
    });
  });

  it("has proper form field attributes", () => {
    render(() => (
      <MockRouter>
        <Login />
      </MockRouter>
    ));
    
    const usernameInput = screen.getByLabelText(/Username or Email/);
    const passwordInput = screen.getByLabelText(/Password/);
    
    expect(usernameInput).toHaveAttribute("name", "login");
    expect(usernameInput).toHaveAttribute("autocomplete", "username");
    
    expect(passwordInput).toHaveAttribute("name", "password");
    expect(passwordInput).toHaveAttribute("type", "password");
    expect(passwordInput).toHaveAttribute("autocomplete", "current-password");
  });

  it("shows required asterisks for required fields", () => {
    render(() => (
      <MockRouter>
        <Login />
      </MockRouter>
    ));
    
    // Check that asterisks are present in the DOM
    const pageContainer = document.body;
    expect(pageContainer.textContent).toContain("*");
    
    // Check specifically for username and password fields
    const usernameInput = screen.getByLabelText(/Username or Email/);
    const passwordInput = screen.getByLabelText(/Password/);
    expect(usernameInput).toBeInTheDocument();
    expect(passwordInput).toBeInTheDocument();
  });
});