import { Component } from "solid-js";
import { Card } from "../../common/ui/Card/Card";
import { useStoryTurn } from "../../../hooks/useStoryTurn";
import type { Story, StoryInvite } from "../../../types/Story";
import styles from "./TurnIndicator.module.scss";

export interface TurnIndicatorProps {
  story: Story;
  invite?: StoryInvite;
  class?: string;
}

export const TurnIndicator: Component<TurnIndicatorProps> = (props) => {
  const turnInfo = () => useStoryTurn(props.story, props.invite);

  const getMessage = () => {
    const info = turnInfo();
    
    if (info.isUserTurn) {
      return "🎯 It's your turn! Continue the story...";
    }
    
    if (info.currentAuthor) {
      return `⏳ Waiting for ${info.currentAuthor.firstName}...`;
    }
    
    return "⏳ Loading...";
  };

  const getVariant = () => {
    return turnInfo().isUserTurn ? "success" : "warning";
  };

  return (
    <Card variant={getVariant()} class={`${styles.turnIndicator} ${props.class || ""}`}>
      <div class={styles.turnMessage}>
        {getMessage()}
      </div>
    </Card>
  );
};

export default TurnIndicator;