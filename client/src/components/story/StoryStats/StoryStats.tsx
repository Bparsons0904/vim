import { Component, createMemo } from "solid-js";
import { Card } from "../../common/ui/Card/Card";
import type { Story } from "../../../types/Story";
import styles from "./StoryStats.module.scss";

export interface StoryStatsProps {
  story: Story;
  class?: string;
}

export const StoryStats: Component<StoryStatsProps> = (props) => {
  const sectionCount = createMemo(() => props.story.storyContent.length);
  const writerCount = createMemo(() => props.story.storyAuthors.length);

  const daysActive = createMemo(() => {
    const diffTime = Math.abs(
      new Date().getTime() - new Date(props.story.createdAt).getTime(),
    );
    return Math.floor(diffTime / (1000 * 60 * 60 * 24));
  });

  const wordCount = createMemo(() => {
    return props.story.storyContent.reduce((total, section) => {
      const textContent = section.content.replace(/<[^>]*>/g, "").trim();
      if (!textContent) {
        return total;
      }
      const words = textContent.split(/\s+/).filter(Boolean);
      return total + words.length;
    }, 0);
  });

  return (
    <Card class={`${styles.sidebarSection} ${props.class || ""}`}>
      <h3 class={styles.sidebarTitle}>📊 Story Stats</h3>
      <div class={styles.storyStats}>
        <div class={styles.stat}>
          <div class={styles.statValue}>{sectionCount()}</div>
          <div class={styles.statLabel}>Sections</div>
        </div>
        <div class={styles.stat}>
          <div class={styles.statValue}>{wordCount()}</div>
          <div class={styles.statLabel}>Words</div>
        </div>
        <div class={styles.stat}>
          <div class={styles.statValue}>{writerCount()}</div>
          <div class={styles.statLabel}>Writers</div>
        </div>
        <div class={styles.stat}>
          <div class={styles.statValue}>{daysActive()}d</div>
          <div class={styles.statLabel}>Active</div>
        </div>
      </div>
    </Card>
  );
};

export default StoryStats;

