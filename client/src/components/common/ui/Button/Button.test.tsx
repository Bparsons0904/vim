import { describe, it, expect, vi, afterEach } from "vitest";
import { render, screen, cleanup } from "@solidjs/testing-library";
import { Button } from "./Button";

describe("Button", () => {
  afterEach(() => {
    cleanup();
  });

  it("renders with default props", () => {
    render(() => <Button>Click me</Button>);
    
    const button = screen.getByRole("button");
    expect(button).toBeInTheDocument();
    expect(button).toHaveTextContent("Click me");
    expect(button).toHaveAttribute("type", "button");
    expect(button).not.toBeDisabled();
  });

  it("renders with primary variant", () => {
    render(() => <Button variant="primary">Primary</Button>);
    const button = screen.getByRole("button");
    expect(button.className).toMatch(/primary/);
  });

  it("renders with secondary variant", () => {
    render(() => <Button variant="secondary">Secondary</Button>);
    const button = screen.getByRole("button");
    expect(button.className).toMatch(/secondary/);
  });

  it("renders with danger variant", () => {
    render(() => <Button variant="danger">Danger</Button>);
    const button = screen.getByRole("button");
    expect(button.className).toMatch(/danger/);
  });

  it("renders with gradient variant", () => {
    render(() => <Button variant="gradient">Gradient</Button>);
    const button = screen.getByRole("button");
    expect(button.className).toMatch(/gradient/);
  });

  it("renders with ghost variant", () => {
    render(() => <Button variant="ghost">Ghost</Button>);
    const button = screen.getByRole("button");
    expect(button.className).toMatch(/ghost/);
  });

  it("renders with tertiary variant", () => {
    render(() => <Button variant="tertiary">Tertiary</Button>);
    const button = screen.getByRole("button");
    expect(button.className).toMatch(/tertiary/);
  });

  it("renders with small size", () => {
    render(() => <Button size="sm">Small</Button>);
    const button = screen.getByRole("button");
    expect(button.className).toMatch(/sm/);
  });

  it("renders with large size", () => {
    render(() => <Button size="lg">Large</Button>);
    const button = screen.getByRole("button");
    expect(button.className).toMatch(/lg/);
  });

  it("handles disabled state", () => {
    render(() => <Button disabled>Disabled</Button>);
    
    const button = screen.getByRole("button");
    expect(button).toBeDisabled();
  });

  it("handles click events", () => {
    const handleClick = vi.fn();
    render(() => <Button onClick={handleClick}>Click me</Button>);
    
    const button = screen.getByRole("button");
    button.click();
    
    expect(handleClick).toHaveBeenCalledTimes(1);
  });

  it("renders with submit type", () => {
    render(() => <Button type="submit">Submit</Button>);
    
    const button = screen.getByRole("button");
    expect(button).toHaveAttribute("type", "submit");
  });

  it("renders with reset type", () => {
    render(() => <Button type="reset">Reset</Button>);
    
    const button = screen.getByRole("button");
    expect(button).toHaveAttribute("type", "reset");
  });

  it("does not trigger click when disabled", () => {
    const handleClick = vi.fn();
    render(() => <Button disabled onClick={handleClick}>Disabled</Button>);
    
    const button = screen.getByRole("button");
    button.click();
    
    expect(handleClick).not.toHaveBeenCalled();
  });

  it("renders with medium size by default", () => {
    render(() => <Button>Default Size</Button>);
    const button = screen.getByRole("button");
    expect(button.className).toMatch(/md/);
  });

  it("renders with primary variant by default", () => {
    render(() => <Button>Default Variant</Button>);
    const button = screen.getByRole("button");
    expect(button.className).toMatch(/primary/);
  });
});