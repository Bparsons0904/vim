import { Component, createSignal } from "solid-js";
import { Card } from "../../common/ui/Card/Card";
import { Button } from "../../common/ui/Button/Button";
import { TextEditor } from "../../common/forms/TextEditor/TextEditor";
import { getTextLength, isHtmlEmpty } from "../../../utils/htmlUtils";
import styles from "./WriteArea.module.scss";

export interface WriteAreaProps {
  storyId: string;
  onSaveDraft?: (html: string) => void;
  onAddSection?: (html: string) => void;
  onChange?: (html: string) => void;
  placeholder?: string;
  maxLength?: number;
  class?: string;
}

export const WriteArea: Component<WriteAreaProps> = (props) => {
  const [storyHtml, setStoryHtml] = createSignal("");
  const maxLength = () => props.maxLength || 1000;

  const handleChange = (html: string) => {
    setStoryHtml(html);
    props.onChange?.(html);
  };

  const handleSaveDraft = () => {
    props.onSaveDraft?.(storyHtml());
  };

  const handleAddSection = () => {
    props.onAddSection?.(storyHtml());
  };

  const getCharacterCount = () => {
    // Calculate character count from HTML, excluding tags
    const textLength = getTextLength(storyHtml());
    return `${textLength} / ${maxLength()} characters`;
  };

  return (
    <Card variant="primary" class={`${styles.writeArea} ${props.class || ""}`}>
      <div class={styles.writePrompt}>
        ✍️ Your turn to continue the story...
      </div>
      <TextEditor
        class={styles.textEditor}
        placeholder={props.placeholder || "Marcus turned around slowly, his heart pounding..."}
        onChange={handleChange}
        minHeight="120px"
      />
      <div class={styles.writeActions}>
        <div class={styles.charCount}>
          {getCharacterCount()}
        </div>
        <div class={styles.writeButtons}>
          <Button
            variant="tertiary"
            size="md"
            onClick={handleSaveDraft}
          >
            💾 Save Draft
          </Button>
          <Button
            variant="primary"
            size="md"
            onClick={handleAddSection}
            disabled={isHtmlEmpty(storyHtml())}
          >
            📝 Add Section
          </Button>
        </div>
      </div>
    </Card>
  );
};

export default WriteArea;
