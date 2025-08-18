import {
  Component,
  createMemo,
  createSignal,
  For,
  onCleanup,
  Show,
} from "solid-js";
import { useParams } from "@solidjs/router";
import { Button } from "../../components/common/ui/Button/Button";
import { TurnIndicator } from "../../components/story/TurnIndicator/TurnIndicator";
import { StorySection } from "../../components/story/StorySection/StorySection";
import { WriteArea } from "../../components/story/WriteArea/WriteArea";
import { ParticipantsList } from "../../components/story/ParticipantsList/ParticipantsList";
import { PendingInvitesList } from "../../components/story/PendingInvitesList/PendingInvitesList";
import { StoryStats } from "../../components/story/StoryStats/StoryStats";
import { InviteLink } from "../../components/story/InviteLink/InviteLink";
import { InviteMembersModal } from "../../components/story/InviteMembersModal/InviteMembersModal";
import { useStory } from "../../services/api/endpoints/stories.api";
import { useAuth } from "../../context/AuthContext";
import { useWebSocket } from "../../context/WebSocketContext";
import { useStoryTurn } from "../../hooks/useStoryTurn";
import styles from "./StoryPage.module.scss";
import { useCreateSection } from "@services/api/endpoints/section.api";

export const StoryPage: Component = () => {
  console.log("Rendering StoryPage");
  const params = useParams<{ id: string }>();
  const { user } = useAuth();
  const { sendMessage } = useWebSocket();
  const storyQuery = useStory(params.id);
  const [showInviteModal, setShowInviteModal] = createSignal(false);

  const story = createMemo(() => storyQuery.data?.story);
  const turnInfo = () => useStoryTurn(story());

  let typingTimeout: NodeJS.Timeout;
  let isTyping = false;

  const handleTyping = () => {
    if (!isTyping) {
      isTyping = true;
      sendMessage(
        JSON.stringify({ type: "start_typing", data: { storyId: params.id } }),
      );
    }
    clearTimeout(typingTimeout);
    typingTimeout = setTimeout(() => {
      isTyping = false;
      sendMessage(
        JSON.stringify({ type: "stop_typing", data: { storyId: params.id } }),
      );
    }, 2000);
  };

  onCleanup(() => {
    clearTimeout(typingTimeout);
    if (isTyping) {
      sendMessage(
        JSON.stringify({ type: "stop_typing", data: { storyId: params.id } }),
      );
    }
  });

  // Claude: Don't we have a hook for this?
  const isUserStoryOwner = createMemo(() => {
    const currentStory = story();
    const currentUser = user;
    return (
      currentStory && currentUser && currentStory.ownerId === currentUser.id
    );
  });

  const handleShareStory = () => {
    console.log("Share story clicked");
  };

  const handleInviteMembers = () => {
    console.log("Invite members clicked");
    setShowInviteModal(true);
  };

  const handleSkipTurn = () => {
    console.log("Skip turn clicked");
  };

  const handleArchiveStory = () => {
    console.log("Archive story clicked");
  };

  const handleDeleteStory = () => {
    console.log("Delete story clicked");
  };

  const handleSaveDraft = (text: string) => {
    console.log("Save draft clicked:", text);
  };

  const createSection = useCreateSection(params.id);
  const handleAddSection = (text: string) => {
    console.log("Add section clicked:", text);
    createSection.mutate({ content: text });
  };

  const handleCopyLink = (link: string) => {
    console.log("Link copied:", link);
  };

  return (
    <div class={styles.storyPage}>
      <Show when={storyQuery.isLoading}>
        <div>Loading story...</div>
      </Show>

      <Show when={storyQuery.isError}>
        <div>Error loading story: {storyQuery.error?.message}</div>
      </Show>

      <Show when={story()}>
        {/* Header Section */}
        <header class={styles.header}>
          <div class={styles.storyInfo}>
            <h1 class={styles.storyTitle}>{story()!.title}</h1>
            <div class={styles.storyMeta}>
              {story()!.genre.name} • Created by {story()!.owner.displayName} •{" "}
              {story()!.storyAuthors.length} participants •{" "}
              {story()!.storyContent.length} sections
            </div>
            <p class={styles.storyDescription}>{story()!.description}</p>
          </div>

          <Show when={isUserStoryOwner()}>
            <div class={styles.storyControls}>
              <Button variant="primary" size="md" onClick={handleShareStory}>
                📤 Share Story
              </Button>
              <Button
                variant="secondary"
                size="md"
                onClick={handleInviteMembers}
              >
                ➕ Invite Members
              </Button>
              <Button variant="tertiary" size="md" onClick={handleSkipTurn}>
                ⏭️ Skip Turn
              </Button>
              <Button variant="tertiary" size="md" onClick={handleArchiveStory}>
                📁 Archive
              </Button>
              <Button variant="danger" size="md" onClick={handleDeleteStory}>
                🗑️ Delete
              </Button>
            </div>
          </Show>
        </header>

        <main class={styles.mainContent}>
          {/* Story Content Area */}
          <div class={styles.storyArea}>
            {/* Turn Indicator */}
            <TurnIndicator story={story()!} />

            {/* Story Chain */}
            <div class={styles.storyChain}>
              <For each={story()!.storyContent}>
                {(content, index) => {
                  const isLastSection =
                    index() === story()!.storyContent.length - 1;
                  const canEdit = isLastSection; // Can only edit the latest section

                  return (
                    <StorySection
                      content={content}
                      index={index()}
                      isLatest={isLastSection}
                      canEdit={canEdit}
                      storyId={params.id}
                    />
                  );
                }}
              </For>
            </div>

            {/* Writing Area */}
            <Show when={turnInfo()?.isUserTurn}>
              <WriteArea
                storyId={params.id}
                onSaveDraft={handleSaveDraft}
                onAddSection={handleAddSection}
                onChange={handleTyping}
              />
            </Show>
          </div>

          {/* Sidebar */}
          <aside class={styles.sidebar}>
            <ParticipantsList story={story()!} />
            <PendingInvitesList story={story()!} />
            <StoryStats story={story()!} />
            <InviteLink story={story()!} onCopyLink={handleCopyLink} />
          </aside>
        </main>

        {/* Invite Members Modal */}
        <Show when={story()}>
          <InviteMembersModal
            isOpen={showInviteModal()}
            onClose={() => setShowInviteModal(false)}
            storyId={params.id}
            storyTitle={story()!.title}
            story={story()!}
          />
        </Show>
      </Show>
    </div>
  );
};

export default StoryPage;
