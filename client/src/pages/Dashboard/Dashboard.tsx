import { Component, For, Show, createSignal, createEffect } from "solid-js";
import { useAuth } from "@context/AuthContext";
import { useToast } from "@context/ToastContext";
import { Story, StoryInvite } from "../../types/Story";
import {
  CreateStoryForm,
  CreateStoryData,
} from "@components/storyForms/CreateStoryForm/CreateStoryForm";
import { CreateStoryRequest } from "@services/api/endpoints/stories.api";
import styles from "./Dashboard.module.scss";
import StoryCard from "@components/storyCards/StoryCard";
import {
  useUserStories,
  useCreateStory,
  useAcceptInvite,
  useDeclineInvite,
} from "@services/api/endpoints/stories.api";
import { Button } from "@components/common/ui/Button/Button";
import { useNavigate } from "@solidjs/router";

const Dashboard: Component = () => {
  const { user } = useAuth();
  const toast = useToast();
  const [stories, setStories] = createSignal<Story[]>([]);
  const [invitations, setInvitations] = createSignal<StoryInvite[]>([]);
  const [showCreateStoryForm, setShowCreateStoryForm] = createSignal(false);

  const storiesQuery = useUserStories();
  const createStoryMutation = useCreateStory();
  const acceptInviteMutation = useAcceptInvite();
  const declineInviteMutation = useDeclineInvite();

  createEffect(() => {
    if (storiesQuery.data) {
      setStories(storiesQuery.data.stories);
      setInvitations(storiesQuery.data.invites);
    }
  });

  // Simple derived values - no need for memoization
  const activeStories = () => stories();
  const pendingInvitations = () => invitations();
  const yourTurnCount = () => activeStories().length;
  const invitationCount = () => pendingInvitations().length;

  const navigate = useNavigate();

  const handleStoryAction = async (storyId: string, action: string) => {
    console.log(`Action: ${action} on story: ${storyId}`);
    try {
      switch (action) {
        case "continue":
        case "view":
          navigate(`/story/${storyId}`);
          break;
        case "accept":
          await acceptInviteMutation.mutateAsync(storyId);
          toast.showSuccess("Invitation accepted successfully!");
          break;
        case "decline":
          await declineInviteMutation.mutateAsync(storyId);
          toast.showSuccess("Invitation declined");
          break;
      }
    } catch (error) {
      console.error(`Failed to ${action} invitation:`, error);
      const errorMessage =
        error instanceof Error
          ? error.message
          : `Failed to ${action} invitation. Please try again.`;
      toast.showError(errorMessage);
    }
  };

  const handleCreateAction = (type: "story" | "group") => {
    console.log(`Create new ${type}`);
    if (type === "story") {
      setShowCreateStoryForm(true);
    }
    // Handle other creation actions
  };

  const handleCreateStory = async (data: CreateStoryData) => {
    try {
      console.log("Creating story:", data);

      // Map CreateStoryData to CreateStoryRequest
      const requestData: CreateStoryRequest = {
        title: data.title,
        description: data.description,
        genreId: data.genreId,
        storyOrderTypeId: data.storyOrderTypeId,
        allowOthersToInvite: data.allowOthersToInvite,
      };

      const result = await createStoryMutation.mutateAsync(requestData);
      console.log("Story created successfully!", result);
      
      if (result?.story?.id) {
        toast.showSuccess("Story created successfully!");
        navigate(`/story/${result.story.id}`);
      } else {
        // Handle invalid result gracefully without throwing
        console.error("Story creation returned invalid result:", result);
        toast.showError("Story creation failed: Invalid response from server. Please try again.");
        return; // Return early instead of throwing
      }
    } catch (error) {
      console.error("Failed to create story:", error);
      const errorMessage =
        error instanceof Error
          ? error.message
          : "Failed to create story. Please try again.";
      toast.showError(errorMessage);
    }
  };

  return (
    <div class={styles.dashboard}>
      <div class={styles.container}>
        {/* Header Section */}
        <header class={styles.header}>
          <h1 class={styles.headerTitle}>
            Welcome back, {user?.firstName || "Writer"}!
          </h1>
          <p class={styles.headerSubtitle}>
            You have {yourTurnCount()} stories waiting for your contribution
            <Show when={invitationCount() > 0}>
              {" "}
              and {invitationCount()} new invitation
              {invitationCount() !== 1 ? "s" : ""}
            </Show>
          </p>
        </header>

        {/* Active Stories Section */}
        <section class={styles.section}>
          <h2 class={styles.sectionTitle}>
            Your Active Stories
            <span class={styles.badge}>{activeStories().length}</span>
          </h2>

          <Show
            when={activeStories().length > 0}
            fallback={
              <div class={styles.emptyState}>
                <div class={styles.emptyIcon}>📚</div>
                <h3>No active stories yet</h3>
                <p>Start a new story or accept an invitation to get writing!</p>
              </div>
            }
          >
            <div class={styles.cardsGrid}>
              <For each={activeStories()}>
                {(story) => (
                  <StoryCard story={story} onAction={handleStoryAction} />
                )}
              </For>
            </div>
          </Show>
        </section>

        {/* Story Invitations Section */}
        <Show when={pendingInvitations().length > 0}>
          <section class={styles.section}>
            <h2 class={styles.sectionTitle}>
              Story Invitations
              <span class={styles.badge}>{invitationCount()}</span>
            </h2>

            <div class={styles.cardsGrid}>
              <For each={pendingInvitations()}>
                {(invitation) => (
                  <StoryCard onAction={handleStoryAction} invite={invitation} />
                )}
              </For>
            </div>
          </section>
        </Show>

        {/* Create Something New Section */}
        <section class={styles.section}>
          <h2 class={styles.sectionTitle}>Create Something New</h2>

          <div class={styles.creationSection}>
            <div
              class={styles.creationCard}
              onClick={() => handleCreateAction("story")}
            >
              <div class={styles.creationIcon}>📖</div>
              <h3 class={styles.creationTitle}>Start New Story</h3>
              <p class={styles.creationDescription}>
                Begin a new story within one of your existing groups
              </p>
              <Button
                variant="tertiary"
                onClick={() => handleCreateAction("story")}
              >
                Get Started
              </Button>
            </div>

            <div
              class={`${styles.creationCard} ${styles.creationCardSecondary}`}
              onClick={() => handleCreateAction("group")}
            >
              <div class={styles.creationIcon}>👥</div>
              <h3 class={styles.creationTitle}>Create Story Group</h3>
              <p class={styles.creationDescription}>
                Invite friends or family to start writing stories together
              </p>
              <Button
                variant="tertiary"
                onClick={() => handleCreateAction("group")}
              >
                Create Group
              </Button>
            </div>
          </div>
        </section>
      </div>

      {/* Create Story Modal */}
      <CreateStoryForm
        isOpen={showCreateStoryForm()}
        onClose={() => setShowCreateStoryForm(false)}
        onSubmit={handleCreateStory}
        isSubmitting={createStoryMutation.isPending}
      />
    </div>
  );
};

export default Dashboard;
