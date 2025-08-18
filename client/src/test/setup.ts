import { vi } from "vitest";
import "@testing-library/jest-dom/vitest";

// Mock the emoji support library directly instead of Canvas
vi.mock("is-emoji-supported", () => ({
  default: vi.fn(() => true), // Just say all emojis are supported
}));
