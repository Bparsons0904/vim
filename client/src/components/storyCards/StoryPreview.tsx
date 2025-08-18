import { Component, createMemo } from "solid-js";
import { Story, StoryInvite } from "../../types/Story";
import { stripHtml } from "../../utils/htmlUtils";
import styles from "./story-card.module.scss";

interface StoryPreviewProps {
  story: Story;
  invite?: StoryInvite;
}

const StoryPreview: Component<StoryPreviewProps> = (props) => {
  const lastContent = createMemo(() => {
    const content = props.story.storyContent;
    return content && content.length > 0 ? content[content.length - 1] : null;
  });

  const previewData = createMemo(() => {
    const content = lastContent();
    if (content) {
      // Strip HTML tags to get plain text for preview
      const plainText = stripHtml(content.content);
      return {
        author: `${content.author.firstName} ${content.author.lastName}`.trim(),
        excerpt:
          plainText.substring(0, 150) +
          (plainText.length > 150 ? "..." : ""),
      };
    }
    return {
      author: "System",
      excerpt: props.story.description || "No content yet. Start writing!",
    };
  });

  return (
    <div class={styles.previewSection}>
      <div class={styles.previewLabel}>
        Last contribution by {previewData().author}:
      </div>
      <div class={styles.previewText}>{previewData().excerpt}</div>
    </div>
  );
};

export default StoryPreview;

