import { Component, For, Show, createMemo } from "solid-js";
import { ParticipantAvatarsProps } from "./types";
import styles from "./story-card.module.scss";

const ParticipantAvatars: Component<ParticipantAvatarsProps> = (props) => {
  const maxVisible = () => props.maxVisible || 4;
  const size = () => props.size || 32;

  const participants = createMemo(
    () =>
      props.story.storyAuthors?.map((storyAuthor) => ({
        id: storyAuthor.author.id,
        name: `${storyAuthor.author.firstName} ${storyAuthor.author.lastName}`.trim(),
        initials:
          `${storyAuthor.author.firstName?.[0] || ""}${storyAuthor.author.lastName?.[0] || ""}`.toUpperCase(),
      })) || [],
  );

  const visibleParticipants = () => participants().slice(0, maxVisible());
  const additionalCount = () => Math.max(0, participants().length - maxVisible());

  const avatarStyle = () => ({
    width: `${size()}px`,
    height: `${size()}px`,
    "font-size": `${size() * 0.375}px`,
  });

  return (
    <div class={styles.participants}>
      <For each={visibleParticipants()}>
        {(participant) => (
          <div
            class={styles.avatar}
            style={avatarStyle()}
            title={participant.name}
          >
            {participant.initials}
          </div>
        )}
      </For>
      <Show when={additionalCount() > 0}>
        <div
          class={`${styles.avatar} ${styles.avatarMore}`}
          style={avatarStyle()}
          title={`+${additionalCount()} more participants`}
        >
          +{additionalCount()}
        </div>
      </Show>
    </div>
  );
};

export default ParticipantAvatars;

