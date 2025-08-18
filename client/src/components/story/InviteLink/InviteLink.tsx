import { Component } from "solid-js";
import { Card } from "../../common/ui/Card/Card";
import { Button } from "../../common/ui/Button/Button";
import type { Story } from "../../../types/Story";
import styles from "./InviteLink.module.scss";

export interface InviteLinkProps {
  story: Story;
  onCopyLink?: (link: string) => void;
  class?: string;
}

export const InviteLink: Component<InviteLinkProps> = (props) => {
  const getInviteLink = () => {
    return `${window.location.origin}/story/${props.story.shortLink}`;
  };

  const handleCopyLink = () => {
    const inviteLink = getInviteLink();
    navigator.clipboard.writeText(inviteLink);
    props.onCopyLink?.(inviteLink);
  };

  return (
    <Card class={`${styles.sidebarSection} ${props.class || ""}`}>
      <h3 class={styles.sidebarTitle}>🔗 Invite Link</h3>
      <div class={styles.inviteLink}>
        {getInviteLink()}
      </div>
      <Button 
        variant="secondary" 
        size="md" 
        onClick={handleCopyLink}
        class={styles.copyButton}
      >
        📋 Copy Link
      </Button>
    </Card>
  );
};

export default InviteLink;