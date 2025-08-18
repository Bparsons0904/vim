import {
  Component,
  createSignal,
  onMount,
  onCleanup,
  Show,
  createEffect,
  For,
  on,
} from "solid-js";
import { createStore } from "solid-js/store";
import { Editor } from "@tiptap/core";
import { Document } from "@tiptap/extension-document";
import { Paragraph } from "@tiptap/extension-paragraph";
import { Text } from "@tiptap/extension-text";
import { Bold } from "@tiptap/extension-bold";
import { Italic } from "@tiptap/extension-italic";
import { Underline } from "@tiptap/extension-underline";
import { Strike } from "@tiptap/extension-strike";
import { BulletList } from "@tiptap/extension-bullet-list";
import { OrderedList } from "@tiptap/extension-ordered-list";
import { ListItem } from "@tiptap/extension-list-item";
import { TextAlign } from "@tiptap/extension-text-align";
import { History } from "@tiptap/extension-history";
import { HardBreak } from "@tiptap/extension-hard-break";
import Emoji from "@tiptap/extension-emoji";
import styles from "./TextEditor.module.scss";
import { commonEmojis } from "./emojis";

// --- PROPS INTERFACE ---
export interface TextEditorProps {
  initialValue?: string;
  placeholder?: string;
  onChange?: (html: string) => void;
  disabled?: boolean;
  class?: string;
  minHeight?: string;
}

// --- EMOJI LIST (can be moved or passed as a prop) ---

// --- COMPONENT ---
export const TextEditor: Component<TextEditorProps> = (props) => {
  const [editorElement, setEditorElement] = createSignal<HTMLDivElement>();
  const [editor, setEditor] = createSignal<Editor>();
  const [showEmojiPicker, setShowEmojiPicker] = createSignal(false);

  // Use a single store for all active states
  const [activeStates, setActiveStates] = createStore({
    bold: false,
    italic: false,
    underline: false,
    strike: false,
    bulletList: false,
    orderedList: false,
    textAlign: "left",
  });

  const updateActiveStates = (editor: Editor) => {
    const alignments = ["left", "center", "right"];
    const activeAlignment =
      alignments.find((align) => editor.isActive({ textAlign: align })) ||
      "left";

    setActiveStates({
      bold: editor.isActive("bold"),
      italic: editor.isActive("italic"),
      underline: editor.isActive("underline"),
      strike: editor.isActive("strike"),
      bulletList: editor.isActive("bulletList"),
      orderedList: editor.isActive("orderedList"),
      textAlign: activeAlignment,
    });
  };

  const handleClickOutside = (event: MouseEvent) => {
    const target = event.target as Element;
    // Check if the click is outside the toolbar to close the emoji picker
    if (!target.closest(`.${styles.toolbar}`)) {
      setShowEmojiPicker(false);
    }
  };

  // --- LIFECYCLE HOOKS ---
  onMount(() => {
    const element = editorElement();
    if (!element) return;

    const newEditor = new Editor({
      element,
      extensions: [
        Document,
        Paragraph,
        Text,
        HardBreak,
        Bold,
        Italic,
        Underline,
        Strike,
        ListItem,
        History,
        BulletList.configure({ HTMLAttributes: { class: styles.bulletList } }),
        OrderedList.configure({
          HTMLAttributes: { class: styles.orderedList },
        }),
        TextAlign.configure({
          types: ["heading", "paragraph"],
          alignments: ["left", "center", "right"],
          defaultAlignment: "left",
        }),
        Emoji.configure({ HTMLAttributes: { class: styles.emoji } }),
      ],
      content: props.initialValue || "",
      editable: !props.disabled,
      onUpdate: ({ editor }) => {
        props.onChange?.(editor.getHTML());
        updateActiveStates(editor);
      },
      onSelectionUpdate: ({ editor }) => {
        updateActiveStates(editor);
      },
      editorProps: {
        attributes: {
          class: styles.editorContent,
          "data-placeholder": props.placeholder ?? "Start typing your story...",
        },
      },
    });

    setEditor(newEditor);
    document.addEventListener("click", handleClickOutside);
  });

  onCleanup(() => {
    editor()?.destroy();
    document.removeEventListener("click", handleClickOutside);
  });

  // --- REACTIVE EFFECTS FOR PROP CHANGES ---
  createEffect(() => {
    editor()?.setEditable(!props.disabled);
  });

  // Use `on` to only trigger when initialValue prop itself changes
  createEffect(
    on(
      () => props.initialValue,
      (newValue) => {
        const currentEditor = editor();
        // Prevent overwriting content if it's the same, which can reset cursor
        if (currentEditor && newValue !== currentEditor.getHTML()) {
          currentEditor.commands.setContent(newValue || "", false);
        }
      },
    ),
  );

  // --- EVENT HANDLERS ---
  const setAlignment = (alignment: "left" | "center" | "right") => {
    editor()?.chain().focus().setTextAlign(alignment).run();
  };

  const insertEmoji = (emoji: string) => {
    editor()?.chain().focus().insertContent(emoji).run();
    setShowEmojiPicker(false);
  };

  // --- JSX ---
  return (
    <div class={`${styles.textEditor} ${props.class || ""}`}>
      {/* Toolbar */}
      <div class={styles.toolbar}>
        {/* Text Formatting */}
        <div class={styles.toolbarGroup}>
          <button
            type="button"
            class={`${styles.toolbarButton} ${activeStates.bold ? styles.active : ""}`}
            onClick={() => editor()?.chain().focus().toggleBold().run()}
            aria-label="Bold"
            title="Bold"
          >
            {/* TODO: Replace with an SVG icon, e.g., <BoldIcon /> */}
            <strong>B</strong>
          </button>

          <button
            type="button"
            class={`${styles.toolbarButton} ${activeStates.italic ? styles.active : ""}`}
            onClick={() => editor()?.chain().focus().toggleItalic().run()}
            aria-label="Italic"
            title="Italic"
          >
            <em>I</em>
          </button>

          <button
            type="button"
            class={`${styles.toolbarButton} ${activeStates.underline ? styles.active : ""}`}
            onClick={() => editor()?.chain().focus().toggleUnderline().run()}
            aria-label="Underline"
            title="Underline"
          >
            <u>U</u>
          </button>

          <button
            type="button"
            class={`${styles.toolbarButton} ${activeStates.strike ? styles.active : ""}`}
            onClick={() => editor()?.chain().focus().toggleStrike().run()}
            aria-label="Strikethrough"
            title="Strikethrough"
          >
            <s>S</s>
          </button>
        </div>

        <div class={styles.separator} />

        {/* Lists */}
        <div class={styles.toolbarGroup}>
          <button
            type="button"
            class={`${styles.toolbarButton} ${activeStates.bulletList ? styles.active : ""}`}
            onClick={() => editor()?.chain().focus().toggleBulletList().run()}
            aria-label="Bullet List"
            title="Bullet List"
          >
            •
          </button>

          <button
            type="button"
            class={`${styles.toolbarButton} ${activeStates.orderedList ? styles.active : ""}`}
            onClick={() => editor()?.chain().focus().toggleOrderedList().run()}
            aria-label="Numbered List"
            title="Numbered List"
          >
            1.
          </button>
        </div>

        <div class={styles.separator} />

        {/* Text Alignment */}
        <div class={styles.toolbarGroup}>
          <button
            type="button"
            class={`${styles.toolbarButton} ${activeStates.textAlign === "left" ? styles.active : ""}`}
            onClick={() => setAlignment("left")}
            aria-label="Align Left"
            title="Align Left"
          >
            ←
          </button>

          <button
            type="button"
            class={`${styles.toolbarButton} ${activeStates.textAlign === "center" ? styles.active : ""}`}
            onClick={() => setAlignment("center")}
            aria-label="Align Center"
            title="Align Center"
          >
            ↔
          </button>

          <button
            type="button"
            class={`${styles.toolbarButton} ${activeStates.textAlign === "right" ? styles.active : ""}`}
            onClick={() => setAlignment("right")}
            aria-label="Align Right"
            title="Align Right"
          >
            →
          </button>
        </div>

        <div class={styles.separator} />

        {/* Emoji Picker */}
        <div class={styles.toolbarGroup}>
          <button
            type="button"
            class={`${styles.toolbarButton} ${showEmojiPicker() ? styles.active : ""}`}
            onClick={() => setShowEmojiPicker(!showEmojiPicker())}
            aria-label="Insert Emoji"
            title="Insert Emoji"
          >
            😀
          </button>
        </div>

        {/* Emoji Picker Dropdown */}
        <Show when={showEmojiPicker()}>
          <div class={styles.emojiPicker}>
            <div class={styles.emojiGrid}>
              <For each={commonEmojis}>
                {(emoji) => (
                  <button
                    type="button"
                    class={styles.emojiButton}
                    onClick={() => insertEmoji(emoji)}
                    title={emoji}
                  >
                    {emoji}
                  </button>
                )}
              </For>
            </div>
          </div>
        </Show>
      </div>

      {/* Editor Container */}
      <div class={styles.editorContainer}>
        <div
          ref={setEditorElement}
          class={styles.editorWrapper}
          style={{ "min-height": props.minHeight || "120px" }}
          data-testid="editor-content"
        />
      </div>
    </div>
  );
};

export default TextEditor;
