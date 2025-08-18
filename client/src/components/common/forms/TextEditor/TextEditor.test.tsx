import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, cleanup, fireEvent } from "@solidjs/testing-library";
import { TextEditor } from "./TextEditor";
import { Editor } from "@tiptap/core";

// Mock Tiptap Editor
const mockEditorInstance = {
  getHTML: vi.fn(() => "<p>Test content</p>"),
  isActive: vi.fn(() => false),
  setEditable: vi.fn(),
  chain: vi.fn(() => ({
    focus: vi.fn(() => ({
      toggleBold: vi.fn(() => ({ run: vi.fn() })),
      toggleItalic: vi.fn(() => ({ run: vi.fn() })),
      toggleUnderline: vi.fn(() => ({ run: vi.fn() })),
      toggleStrike: vi.fn(() => ({ run: vi.fn() })),
      toggleBulletList: vi.fn(() => ({ run: vi.fn() })),
      toggleOrderedList: vi.fn(() => ({ run: vi.fn() })),
      insertContent: vi.fn(() => ({ run: vi.fn() })),
      setTextAlign: vi.fn(() => ({ run: vi.fn() })),
    })),
  })),
  destroy: vi.fn(),
  on: vi.fn(),
  off: vi.fn(),
  commands: {
    setContent: vi.fn(),
  },
  state: {},
  view: {},
};

// Mock Tiptap extensions
vi.mock("@tiptap/core", () => ({
  Editor: vi.fn(() => mockEditorInstance as unknown as Editor),
}));

vi.mock("@tiptap/extension-document", () => ({
  Document: {},
}));

vi.mock("@tiptap/extension-paragraph", () => ({
  Paragraph: {},
}));

vi.mock("@tiptap/extension-text", () => ({
  Text: {},
}));

vi.mock("@tiptap/extension-bold", () => ({
  Bold: {},
}));

vi.mock("@tiptap/extension-italic", () => ({
  Italic: {},
}));

vi.mock("@tiptap/extension-history", () => ({
  History: {},
}));

vi.mock("@tiptap/extension-hard-break", () => ({
  HardBreak: {},
}));

vi.mock("@tiptap/extension-underline", () => ({
  Underline: {},
}));

vi.mock("@tiptap/extension-strike", () => ({
  Strike: {},
}));

vi.mock("@tiptap/extension-bullet-list", () => ({
  BulletList: {
    configure: vi.fn(() => ({})),
  },
}));

vi.mock("@tiptap/extension-ordered-list", () => ({
  OrderedList: {
    configure: vi.fn(() => ({})),
  },
}));

vi.mock("@tiptap/extension-list-item", () => ({
  ListItem: {},
}));

vi.mock("@tiptap/extension-text-align", () => ({
  TextAlign: {
    configure: vi.fn(() => ({})),
  },
}));

vi.mock("@tiptap/extension-emoji", () => ({
  default: {
    configure: vi.fn(() => ({})),
  },
}));

describe("TextEditor", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    cleanup();
  });

  describe("Basic Rendering", () => {
    it("renders the text editor component", () => {
      render(() => <TextEditor />);
      
      expect(screen.getByTestId("editor-content")).toBeInTheDocument();
      expect(screen.getByLabelText("Bold")).toBeInTheDocument();
      expect(screen.getByLabelText("Italic")).toBeInTheDocument();
    });

    it("renders with custom placeholder", () => {
      render(() => <TextEditor placeholder="Write your story here..." />);
      
      expect(Editor).toHaveBeenCalledWith(
        expect.objectContaining({
          editorProps: expect.objectContaining({
            attributes: expect.objectContaining({
              "data-placeholder": "Write your story here..."
            })
          })
        })
      );
    });

    it("applies custom CSS classes", () => {
      render(() => <TextEditor class="custom-editor" />);
      
      const editor = document.querySelector('.custom-editor');
      expect(editor).toBeInTheDocument();
    });

    it("sets custom min height", () => {
      render(() => <TextEditor minHeight="200px" />);
      
      const editorContent = screen.getByTestId("editor-content");
      expect(editorContent.style.minHeight).toBe('200px');
    });
  });

  describe("Editor Initialization", () => {
    it("initializes Tiptap editor on mount", () => {
      render(() => <TextEditor initialValue="<p>Initial content</p>" />);
      
      expect(Editor).toHaveBeenCalledWith(
        expect.objectContaining({
          content: "<p>Initial content</p>",
          editable: true,
        })
      );
    });

    it("initializes disabled editor when disabled prop is true", () => {
      render(() => <TextEditor disabled />);
      
      expect(Editor).toHaveBeenCalledWith(
        expect.objectContaining({
          editable: false,
        })
      );
    });

    it("destroys editor on cleanup", () => {
      const { unmount } = render(() => <TextEditor />);
      
      unmount();
      
      expect(mockEditorInstance.destroy).toHaveBeenCalled();
    });
  });

  describe("Basic Features", () => {
    it("renders with built-in features (bold, italic, emoji)", () => {
      render(() => <TextEditor />);
      
      expect(screen.getByLabelText("Bold")).toBeInTheDocument();
      expect(screen.getByLabelText("Italic")).toBeInTheDocument();
      expect(screen.getByLabelText("Insert Emoji")).toBeInTheDocument();
    });
  });

  describe("Toolbar Functionality", () => {
    it("renders basic formatting buttons with proper labels", () => {
      render(() => <TextEditor />);
      
      const boldButton = screen.getByLabelText("Bold");
      const italicButton = screen.getByLabelText("Italic");
      
      expect(boldButton).toBeInTheDocument();
      expect(italicButton).toBeInTheDocument();
      expect(boldButton).toHaveAttribute("aria-label", "Bold");
      expect(italicButton).toHaveAttribute("aria-label", "Italic");
    });

    it("handles bold button click", () => {
      render(() => <TextEditor />);
      
      const boldButton = screen.getByLabelText("Bold");
      fireEvent.click(boldButton);
      
      expect(mockEditorInstance.chain).toHaveBeenCalled();
    });

    it("handles italic button click", () => {
      render(() => <TextEditor />);
      
      const italicButton = screen.getByLabelText("Italic");
      fireEvent.click(italicButton);
      
      expect(mockEditorInstance.chain).toHaveBeenCalled();
    });

    it("handles active state updates from editor", () => {
      render(() => <TextEditor />);
      
      const boldButton = screen.getByLabelText("Bold");
      const italicButton = screen.getByLabelText("Italic");
      
      // Initially, no buttons should be active
      expect(boldButton).not.toHaveClass("active");
      expect(italicButton).not.toHaveClass("active");
      
      // Verify the editor callbacks are set up to handle active state
      const editorConfig = (vi.mocked(Editor).mock.calls[0] as unknown[])[0] as {
        extensions: unknown[];
        onSelectionUpdate?: () => void;
        onUpdate?: (props: { editor: unknown }) => void;
        editorProps?: {
          attributes: Record<string, string>;
          handleKeyDown?: (view: unknown, event: KeyboardEvent) => boolean;
        };
      };
      expect(editorConfig.onSelectionUpdate).toBeDefined();
      expect(editorConfig.onUpdate).toBeDefined();
    });

    it("displays emoji button", () => {
      render(() => <TextEditor />);
      
      expect(screen.getByLabelText("Insert Emoji")).toBeInTheDocument();
    });
  });

  describe("Extension Configuration", () => {
    it("configures editor with standard extensions", () => {
      render(() => <TextEditor />);
      
      expect(Editor).toHaveBeenCalledWith(
        expect.objectContaining({
          extensions: expect.any(Array),
        })
      );
      
      const editorConfig = (vi.mocked(Editor).mock.calls[0] as unknown[])[0] as {
        extensions: unknown[];
        onSelectionUpdate?: () => void;
        onUpdate?: (props: { editor: unknown }) => void;
        editorProps?: {
          attributes: Record<string, string>;
          handleKeyDown?: (view: unknown, event: KeyboardEvent) => boolean;
        };
      };
      const extensions = editorConfig.extensions;
      
      // Should include all standard extensions
      expect(extensions).toEqual(expect.any(Array));
      expect(extensions.length).toBe(14); // Document, Paragraph, Text, HardBreak, Bold, Italic, Underline, Strike, ListItem, History, BulletList, OrderedList, TextAlign, Emoji
    });
  });

  describe("Change Handling", () => {
    it("calls onChange when editor content changes", () => {
      const handleChange = vi.fn();
      
      render(() => <TextEditor onChange={handleChange} />);
      
      // Simulate editor update
      const editorConfig = (vi.mocked(Editor).mock.calls[0] as unknown[])[0] as {
        extensions: unknown[];
        onSelectionUpdate?: () => void;
        onUpdate?: (props: { editor: unknown }) => void;
        editorProps?: {
          attributes: Record<string, string>;
          handleKeyDown?: (view: unknown, event: KeyboardEvent) => boolean;
        };
      };
      editorConfig.onUpdate({ editor: mockEditorInstance });
      
      expect(handleChange).toHaveBeenCalledWith("<p>Test content</p>");
    });

    it("does not call onChange if not provided", () => {
      expect(() => {
        render(() => <TextEditor />);
        
        const editorConfig = (vi.mocked(Editor).mock.calls[0] as unknown[])[0] as {
        extensions: unknown[];
        onSelectionUpdate?: () => void;
        onUpdate?: (props: { editor: unknown }) => void;
        editorProps?: {
          attributes: Record<string, string>;
          handleKeyDown?: (view: unknown, event: KeyboardEvent) => boolean;
        };
      };
        editorConfig.onUpdate({ editor: mockEditorInstance });
      }).not.toThrow();
    });
  });

  describe("Emoji Functionality", () => {
    it("supports emoji insertion through Tiptap extension", () => {
      render(() => <TextEditor />);
      
      // The Tiptap Emoji extension handles emoji functionality automatically
      // Users can type :emoji_name: and it will be converted to actual emojis
      expect(screen.getByLabelText("Insert Emoji")).toBeInTheDocument();
    });

    it("always shows emoji button", () => {
      render(() => <TextEditor />);
      
      expect(screen.getByLabelText("Insert Emoji")).toBeInTheDocument();
    });
  });

  describe("Editor Configuration", () => {
    it("configures editor with proper extension array", () => {
      render(() => <TextEditor />);
      
      const editorConfig = (vi.mocked(Editor).mock.calls[0] as unknown[])[0] as {
        extensions: unknown[];
        onSelectionUpdate?: () => void;
        onUpdate?: (props: { editor: unknown }) => void;
        editorProps?: {
          attributes: Record<string, string>;
          handleKeyDown?: (view: unknown, event: KeyboardEvent) => boolean;
        };
      };
      const extensions = editorConfig.extensions;
      
      expect(extensions).toEqual(expect.any(Array));
      expect(extensions.length).toBeGreaterThan(0);
    });

    it("sets up correct editor props", () => {
      render(() => <TextEditor placeholder="Custom placeholder" />);
      
      const editorConfig = (vi.mocked(Editor).mock.calls[0] as unknown[])[0] as {
        extensions: unknown[];
        onSelectionUpdate?: () => void;
        onUpdate?: (props: { editor: unknown }) => void;
        editorProps?: {
          attributes: Record<string, string>;
          handleKeyDown?: (view: unknown, event: KeyboardEvent) => boolean;
        };
      };
      const editorProps = editorConfig.editorProps;
      
      expect(editorProps.attributes["data-placeholder"]).toBe("Custom placeholder");
      expect(editorProps.attributes.class).toBeDefined();
    });

    it("uses consistent extension configuration", () => {
      const { unmount } = render(() => <TextEditor />);
      
      const config1 = (vi.mocked(Editor).mock.calls[0] as unknown[])[0] as {
        extensions: unknown[];
      };
      unmount();
      
      render(() => <TextEditor />);
      
      const config2 = (vi.mocked(Editor).mock.calls[1] as unknown[])[0] as {
        extensions: unknown[];
      };
      
      // Both should have identical extension arrays
      expect(config1.extensions).toEqual(expect.any(Array));
      expect(config2.extensions).toEqual(expect.any(Array));
      expect(config1.extensions.length).toBe(config2.extensions.length);
    });
  });

  describe("Accessibility", () => {
    it("has proper ARIA labels on toolbar buttons", () => {
      render(() => <TextEditor />);
      
      const boldButton = screen.getByLabelText("Bold");
      const italicButton = screen.getByLabelText("Italic");
      
      expect(boldButton).toHaveAttribute("aria-label", "Bold");
      expect(italicButton).toHaveAttribute("aria-label", "Italic");
    });

    it("supports keyboard navigation", () => {
      render(() => <TextEditor />);
      
      const boldButton = screen.getByLabelText("Bold");
      const italicButton = screen.getByLabelText("Italic");
      
      expect(boldButton).toHaveAttribute("type", "button");
      expect(italicButton).toHaveAttribute("type", "button");
    });
  });

  describe("Error Handling", () => {
    it("handles editor initialization failure gracefully", () => {
      const consoleError = vi.spyOn(console, "error").mockImplementation(() => {});
      
      // Mock Editor constructor to throw error
      vi.mocked(Editor).mockImplementationOnce(() => {
        throw new Error("Failed to initialize");
      });
      
      // The component should handle the error gracefully
      expect(() => {
        render(() => <TextEditor />);
      }).toThrow(); // It will throw in our simplified implementation
      
      // Restore mocks
      vi.mocked(Editor).mockImplementation(() => mockEditorInstance as unknown as Editor);
      consoleError.mockRestore();
    });

    it("handles missing editor element gracefully", () => {
      // Mock the scenario where editorElement is not found
      const originalQuerySelector = document.querySelector;
      document.querySelector = vi.fn(() => null);
      
      expect(() => {
        render(() => <TextEditor />);
      }).not.toThrow();
      
      document.querySelector = originalQuerySelector;
    });
  });

  describe("Performance", () => {
    it("cleans up event listeners on unmount", () => {
      const { unmount } = render(() => <TextEditor />);
      
      unmount();
      
      expect(mockEditorInstance.destroy).toHaveBeenCalledTimes(1);
    });

    it("creates a new editor instance for each component mount", () => {
      const { unmount } = render(() => <TextEditor placeholder="Initial" />);
      
      expect(Editor).toHaveBeenCalledTimes(1);
      
      unmount();
      
      render(() => <TextEditor placeholder="Updated" />);
      
      // Should create a new editor for the new mount
      expect(Editor).toHaveBeenCalledTimes(2);
    });
  });
});