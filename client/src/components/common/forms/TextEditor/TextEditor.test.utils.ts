/**
 * Test utilities for TextEditor component
 * Provides helpers for mocking Tiptap Editor and testing rich text functionality
 */

import { vi, expect } from "vitest";
import { stripHtml } from "../../../../utils/htmlUtils";

export interface MockEditorInstance {
  getHTML: ReturnType<typeof vi.fn>;
  getText: ReturnType<typeof vi.fn>;
  isActive: ReturnType<typeof vi.fn>;
  chain: ReturnType<typeof vi.fn>;
  destroy: ReturnType<typeof vi.fn>;
  on: ReturnType<typeof vi.fn>;
  off: ReturnType<typeof vi.fn>;
  commands: object;
  state: object;
  view: object;
  setOptions: ReturnType<typeof vi.fn>;
  focus: ReturnType<typeof vi.fn>;
  blur: ReturnType<typeof vi.fn>;
  setEditable: ReturnType<typeof vi.fn>;
}

/**
 * Creates a mock Tiptap Editor instance for testing
 */
export const createMockEditor = (
  overrides: Partial<MockEditorInstance> = {},
): MockEditorInstance => {
  const mockChain = vi.fn(() => ({
    focus: vi.fn(() => ({
      toggleBold: vi.fn(() => ({ run: vi.fn() })),
      toggleItalic: vi.fn(() => ({ run: vi.fn() })),
      insertContent: vi.fn(() => ({ run: vi.fn() })),
      clearContent: vi.fn(() => ({ run: vi.fn() })),
      setContent: vi.fn(() => ({ run: vi.fn() })),
    })),
  }));

  return {
    getHTML: vi.fn(() => "<p></p>"),
    getText: vi.fn(() => ""),
    isActive: vi.fn(() => false),
    chain: mockChain,
    destroy: vi.fn(),
    on: vi.fn(),
    off: vi.fn(),
    commands: {
      setContent: vi.fn(),
    },
    state: {},
    view: {},
    setOptions: vi.fn(),
    focus: vi.fn(),
    blur: vi.fn(),
    setEditable: vi.fn(),
    ...overrides,
  };
};

/**
 * Mock Editor constructor that returns a mock instance
 */
export const createMockEditorConstructor = (
  mockInstance?: MockEditorInstance,
) => {
  const instance = mockInstance || createMockEditor();
  return vi.fn(() => instance as unknown as MockEditorInstance);
};

/**
 * Helper to simulate editor content changes
 */
export const simulateEditorUpdate = (
  mockEditor: MockEditorInstance,
  newContent: string,
  updateCallback?: (params: { editor: MockEditorInstance }) => void,
) => {
  mockEditor.getHTML.mockReturnValue(newContent);
  mockEditor.getText.mockReturnValue(stripHtml(newContent));

  if (updateCallback) {
    updateCallback({ editor: mockEditor });
  }
};

/**
 * Helper to simulate active formatting states
 */
export const simulateActiveStates = (
  mockEditor: MockEditorInstance,
  activeFormats: string[],
) => {
  mockEditor.isActive.mockImplementation((format: string) =>
    activeFormats.includes(format),
  );
};

/**
 * Helper to simulate keyboard events in the editor
 */
export const simulateKeyboardEvent = (
  editorConfig: {
    editorProps?: {
      handleKeyDown?: (view: unknown, event: KeyboardEvent) => boolean;
    };
  },
  key: string,
  options: Partial<KeyboardEvent> = {},
): boolean => {
  const event = {
    key,
    preventDefault: vi.fn(),
    stopPropagation: vi.fn(),
    ...options,
  } as KeyboardEvent;

  return editorConfig.editorProps?.handleKeyDown?.(null, event) ?? false;
};

/**
 * Helper to get editor configuration from Editor mock calls
 */
export const getEditorConfig = (
  EditorMock: { mock: { calls: unknown[][] } },
  callIndex = 0,
) => {
  return EditorMock.mock.calls[callIndex]?.[0] || {};
};

/**
 * Test data for emoji picker
 */
export const TEST_EMOJIS = [
  "😊",
  "😂",
  "🎉",
  "❤️",
  "👍",
  "🔥",
  "💯",
  "✨",
  "🚀",
  "💡",
];

/**
 * HTML content samples for testing
 */
export const TEST_HTML_CONTENT = {
  empty: "<p></p>",
  plainText: "<p>Hello world</p>",
  boldText: "<p><strong>Bold text</strong></p>",
  italicText: "<p><em>Italic text</em></p>",
  mixedFormatting: "<p><strong>Bold</strong> and <em>italic</em> text</p>",
  multiParagraph: "<p>First paragraph</p><p>Second paragraph</p>",
  withEmoji: "<p>Hello 😊 world</p>",
  longContent: "<p>" + "A".repeat(500) + "</p>",
};

/**
 * Accessibility test helpers
 */
export const accessibilityHelpers = {
  /**
   * Check if element has proper ARIA attributes
   */
  checkAriaAttributes: (
    element: HTMLElement,
    expectedAttributes: Record<string, string>,
  ) => {
    Object.entries(expectedAttributes).forEach(([attr, value]) => {
      expect(element).toHaveAttribute(attr, value);
    });
  },

  /**
   * Check if element is keyboard accessible
   */
  checkKeyboardAccessible: (element: HTMLElement) => {
    expect(element).toHaveAttribute("type", "button");
    expect(element.tabIndex).not.toBe(-1);
  },
};

/**
 * Performance test helpers
 */
export const performanceHelpers = {
  /**
   * Measure render time (mock implementation for testing)
   */
  measureRenderTime: (renderFn: () => void): number => {
    const start = performance.now();
    renderFn();
    return performance.now() - start;
  },

  /**
   * Check for memory leaks in editor cleanup
   */
  checkMemoryCleanup: (mockEditor: MockEditorInstance) => {
    expect(mockEditor.destroy).toHaveBeenCalled();
    expect(mockEditor.off).toHaveBeenCalled();
  },
};

/**
 * Integration test helpers
 */
export const integrationHelpers = {
  /**
   * Simulate user typing in the editor
   */
  simulateTyping: async (
    mockEditor: MockEditorInstance,
    text: string,
    updateCallback?: (params: { editor: MockEditorInstance }) => void,
  ) => {
    const finalContent = `<p>${text}</p>`;
    simulateEditorUpdate(mockEditor, finalContent, updateCallback);
  },

  /**
   * Simulate user formatting text
   */
  simulateFormatting: async (
    mockEditor: MockEditorInstance,
    format: "bold" | "italic",
    isActive: boolean = true,
  ) => {
    simulateActiveStates(mockEditor, isActive ? [format] : []);

    // Simulate the formatting command
    const chainMock = mockEditor.chain();
    if (format === "bold") {
      chainMock.focus().toggleBold();
    } else {
      chainMock.focus().toggleItalic();
    }
  },
};

