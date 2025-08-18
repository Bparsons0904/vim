import { Component, For, createSignal, createEffect } from "solid-js";
import { Card } from "../../common/ui/Card/Card";
import { Avatar } from "../../common/ui/Avatar/Avatar";
import { useWebSocket } from "../../../context/WebSocketContext";
import type { Story  } from "../../../types/Story";
import styles from "./ParticipantsList.module.scss";

export interface ParticipantsListProps {
  story: Story;
  class?: string;
}

export const ParticipantsList: Component<ParticipantsListProps> = (props) => {
  const { lastMessage } = useWebSocket();
  const [typingUsers, setTypingUsers] = createSignal<string[]>([]);

  createEffect(() => {
    const message = lastMessage();
    if (!message) return;

    const parsedMessage = JSON.parse(message);
    if (parsedMessage.data?.storyId !== props.story.id) return;

    if (parsedMessage.type === "start_typing") {
      setTypingUsers((prev) => [...new Set([...prev, parsedMessage.userId])]);
    } else if (parsedMessage.type === "stop_typing") {
      setTypingUsers((prev) =>
        prev.filter((id) => id !== parsedMessage.userId),
      );
    }
  });

  return (
    <Card class={`${styles.sidebarSection} ${props.class || ""}`}>
      <h3 class={styles.sidebarTitle}>👥 Participants</h3>
      <div class={styles.participantsList}>
        <For each={props.story.storyAuthors}>
          {(storyAuthor) => {
            const isCurrentTurn = storyAuthor.authorId === props.story.nextUpId;
            const isCreator = storyAuthor.authorId === props.story.ownerId;

            return (
              <div
                class={`${styles.participant} ${isCurrentTurn ? styles.currentTurn : ""}`}
              >
                <Avatar
                  name={storyAuthor.author.displayName}
                  size="md"
                  class={styles.participantAvatar}
                />
                <div class={styles.participantInfo}>
                  <div class={styles.participantName}>
                    {storyAuthor.author.displayName}
                  </div>
                  <div class={styles.participantStatus}>
                    {typingUsers().includes(storyAuthor.authorId) ? (
                      <span class={styles.typingIndicator}>typing...</span>
                    ) : isCreator ? (
                      "Story Creator"
                    ) : isCurrentTurn ? (
                      "Current Turn"
                    ) : (
                      "Active"
                    )}
                  </div>
                </div>
              </div>
            );
          }}
        </For>
      </div>
    </Card>
  );
};

export default ParticipantsList;
