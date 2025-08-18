import { Component, Match, Switch } from "solid-js";
import { TurnStatusBadgeProps } from "./types";
import styles from "./story-card.module.scss";
import { useStoryTurn } from "@hooks/useStoryTurn";

const TurnStatusBadge: Component<TurnStatusBadgeProps> = (props) => {
  const { currentAuthor, isUserTurn, isInvited } = useStoryTurn(
    props.story,
    props.invite,
  );

  return (
    <Switch>
      <Match when={isInvited}>
        <div class={`${styles.turnStatus} ${styles.invitation}`}>
          ✉️ Invited by {props.invite?.inviter?.firstName || "someone"}
        </div>
      </Match>
      <Match when={isUserTurn}>
        <div class={`${styles.turnStatus} ${styles.yourTurn}`}>
          ⏰ Your turn to write
        </div>
      </Match>
      <Match when={true}>
        <div class={`${styles.turnStatus} ${styles.waiting}`}>
          ⏳ Waiting for {currentAuthor.firstName}
        </div>
      </Match>
    </Switch>
  );
};

export default TurnStatusBadge;
