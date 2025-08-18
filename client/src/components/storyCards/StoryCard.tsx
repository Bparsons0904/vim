import { Component, Match, Switch } from "solid-js";
import { StoryCardProps } from "./types";
import TurnStatusBadge from "./TurnStatusBadge";
import ParticipantAvatars from "./ParticipantAvatars";
import StoryPreview from "./StoryPreview";
import styles from "./story-card.module.scss";
import { Button } from "@components/common/ui/Button/Button";
import { useStoryTurn } from "@hooks/useStoryTurn";
import { useRelativeTime } from "@hooks/useRelativeTime";
import clsx from "clsx";

// interface ActionButton {
//   label: string;
//   action: string;
//   variant: "primary" | "secondary" | "success";
// }

// type StoryCardType = "your-turn" | "waiting" | "invitation";

const StoryCard: Component<StoryCardProps> = (props) => {
  if (props.invite) {
    props.story = props.invite.story;
  }
  // For now, default to 'your-turn' status since we need to implement proper logic
  // TODO: Implement proper status logic based on story authors and current user
  // const cardType = (): StoryCardType => "your-turn";

  // const cardClass = () => {
  //   const baseClass = styles.storyCard;
  //   const typeClass = (() => {
  //     switch (cardType()) {
  //       case "your-turn":
  //         return styles.yourTurn;
  //       case "waiting":
  //         return styles.waiting;
  //       case "invitation":
  //         return styles.invitation;
  //       default:
  //         return "";
  //     }
  //   })();
  //
  //   return `${baseClass} ${typeClass} ${props.class || ""}`.trim();
  // };

  // const actions = (): ActionButton[] => {
  //   switch (cardType()) {
  //     case "your-turn":
  //     case "waiting":
  //       return [
  //         { label: "Continue Story", action: "continue", variant: "primary" },
  //       ];
  //     case "invitation":
  //       return [
  //         { label: "Accept", action: "accept", variant: "success" },
  //         { label: "Decline", action: "decline", variant: "secondary" },
  //       ];
  //     default:
  //       return [];
  //   }
  // };

  // Calculate the timestamp of the most recent story content
  const lastActivityTime =
    props.story.storyContent?.length === 0
      ? props.story.updatedAt
      : props.story.storyContent
          ?.slice()
          .sort(
            (a, b) =>
              new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime(),
          )[0].createdAt;

  const relativeTime = useRelativeTime(lastActivityTime);

  // const getButtonClass = (variant: string) => {
  //   switch (variant) {
  //     case "primary":
  //       return `${styles.btn} ${styles.primary}`;
  //     case "secondary":
  //       return `${styles.btn} ${styles.secondary}`;
  //     case "success":
  //       return `${styles.btn} ${styles.success}`;
  //     default:
  //       return `${styles.btn} ${styles.secondary}`;
  //   }
  // };

  const handleActionClick = (action: string, event: Event) => {
    event.stopPropagation();
    // For invitation actions, pass the invite ID; otherwise pass the story ID
    const actionId = props.invite && (action === "accept" || action === "decline") 
      ? props.invite.id 
      : props.story.id;
    props.onAction(actionId, action);
  };

  const handleCardClick = () => {
    props.onAction(props.story.id, "view");
  };

  const { currentAuthor, isUserTurn, isInvited } = useStoryTurn(
    props.story,
    props.invite,
  );

  return (
    <div
      class={clsx(
        styles.storyCard,
        isUserTurn ? styles.yourTurn : styles.waiting,
        isInvited ? styles.invitation : "",
      )}
      onClick={handleCardClick}
    >
      {/* Header Section */}
      <header class={styles.cardHeader}>
        <div class={styles.headerLeft}>
          <div class={styles.storyTitle}>{props.story.title}</div>
          <div class={styles.storyGroup}>
            {props.story.genre?.name || "General"}
          </div>
        </div>
        <div class={styles.headerRight}>
          <TurnStatusBadge story={props.story} invite={props.invite} />
        </div>
      </header>

      {/* Content Area - grows to fill space */}
      <div class={styles.contentArea}>
        {/* Story Preview Section */}
        <StoryPreview story={props.story} />
      </div>

      {/* Bottom Section - stays at bottom */}
      <div class={styles.bottomSection}>
        {/* Participants Section */}
        <ParticipantAvatars story={props.story} />

        {/* Footer Section */}
        <footer class={styles.cardFooter}>
          <div class={styles.lastActivity}>Last update: {relativeTime()}</div>
          <div class={styles.footerActions}>
            {/* <For each={actions()}> */}
            {/*   {(action) => ( */}
            {/*     <button */}
            {/*       class={getButtonClass(action.variant)} */}
            {/*       onClick={(e) => handleActionClick(action.action, e)} */}
            {/*     > */}
            {/*       {action.label} */}
            {/*     </button> */}
            {/*   )} */}
            {/* </For> */}
            <Switch>
              <Match when={!!props.invite}>
                <Button
                  variant="secondary"
                  onClick={(e) => handleActionClick("decline", e)}
                >
                  Decline
                </Button>
                <Button
                  variant="primary"
                  onClick={(e) => handleActionClick("accept", e)}
                >
                  Accept
                </Button>
              </Match>
              <Match when={isUserTurn}>
                <Button
                  variant="primary"
                  onClick={(e) => handleActionClick("continue", e)}
                >
                  Continue Story
                </Button>
              </Match>
              <Match when={!isUserTurn}>
                <Button
                  variant="warning"
                  onClick={(e) => handleActionClick("nudge", e)}
                >
                  Nudge {currentAuthor?.firstName}
                </Button>
              </Match>
            </Switch>
          </div>
        </footer>
      </div>
    </div>
  );
};

export default StoryCard;
