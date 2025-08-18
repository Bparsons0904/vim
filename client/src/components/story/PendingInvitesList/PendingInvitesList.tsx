import { Component, createMemo, For } from "solid-js";
import { Card } from "../../common/ui/Card/Card";
import { Avatar } from "../../common/ui/Avatar/Avatar";
import { Button } from "../../common/ui/Button/Button";
import { useAuth } from "../../../context/AuthContext";
import { useRevokeInvite } from "../../../services/api/endpoints/stories.api";
import { useToast } from "../../../context/ToastContext";
import type { Story } from "../../../types/Story";
import type { User } from "../../../types/User";
import styles from "./PendingInvitesList.module.scss";

export interface PendingInvitesListProps {
  story: Story;
  class?: string;
}

export const PendingInvitesList: Component<PendingInvitesListProps> = (
  props,
) => {
  const { user } = useAuth();
  const revokeInviteMutation = useRevokeInvite();
  const toast = useToast();

  const getUserDisplayName = (user: User) => {
    return `${user.firstName} ${user.lastName}`;
  };

  const isUserStoryOwner = createMemo(() => {
    return user && props.story.ownerId === user.id;
  });

  const pendingInvites = createMemo(() => {
    return props.story.storyInvites.filter(
      (invite) => invite.status === "pending" && invite.user,
    );
  });

  const handleRevokeInvite = async (inviteId: string, userName: string) => {
    try {
      await revokeInviteMutation.mutateAsync(inviteId);
      toast.showSuccess(`Invitation to ${userName} has been revoked`);
    } catch (error) {
      console.error("Failed to revoke invite:", error);
      const errorMessage =
        error instanceof Error
          ? error.message
          : "Failed to revoke invitation. Please try again.";
      toast.showError(errorMessage);
    }
  };

  // Don't render anything if there are no pending invites
  if (pendingInvites().length === 0) {
    return null;
  }

  return (
    <Card class={`${styles.sidebarSection} ${props.class || ""}`}>
      <h3 class={styles.sidebarTitle}>📤 Pending Invites</h3>
      <div class={styles.invitesList}>
        <For each={pendingInvites()}>
          {(invite, index) => (
            <div class={styles.inviteItem}>
              <div class={styles.inviteUser}>
                <Avatar
                  name={getUserDisplayName(invite.user!)}
                  variant={((index() % 5) + 1) as 1 | 2 | 3 | 4 | 5}
                  size="md"
                  class={styles.userAvatar}
                />
                <div class={styles.userInfo}>
                  <div class={styles.userName}>
                    {getUserDisplayName(invite.user!)}
                  </div>
                  <div class={styles.userStatus}>Invited</div>
                </div>
              </div>
              {isUserStoryOwner() && (
                <Button
                  variant="danger"
                  size="sm"
                  onClick={() =>
                    handleRevokeInvite(
                      invite.id,
                      getUserDisplayName(invite.user!),
                    )
                  }
                  disabled={revokeInviteMutation.isPending}
                  class={styles.revokeButton}
                  // title={`Revoke invitation for ${getUserDisplayName(invite.user!)}`}
                >
                  ×
                </Button>
              )}
            </div>
          )}
        </For>
      </div>
    </Card>
  );
};

export default PendingInvitesList;

