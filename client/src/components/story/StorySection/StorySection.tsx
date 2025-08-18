import { Component, createSignal, Show } from "solid-js";
import { Card } from "../../common/ui/Card/Card";
import { Avatar } from "../../common/ui/Avatar/Avatar";
import { Button } from "../../common/ui/Button/Button";
import { TextEditor } from "../../common/forms/TextEditor/TextEditor";
import { useRelativeTime } from "../../../hooks/useRelativeTime";
import { useAuth } from "../../../context/AuthContext";
import {
  useUpdateSection,
  useDeleteSection,
} from "../../../services/api/endpoints/section.api";
import { isHtmlEmpty } from "../../../utils/htmlUtils";
import type { StoryContent } from "../../../types/Story";
import styles from "./StorySection.module.scss";

export interface StorySectionProps {
  content: StoryContent;
  index: number;
  isLatest?: boolean;
  canEdit?: boolean; 
  storyId: string;
  class?: string;
}

export const StorySection: Component<StorySectionProps> = (props) => {
  const relativeTime = useRelativeTime(props.content.createdAt);
  const { user } = useAuth();
  const [isEditing, setIsEditing] = createSignal(false);
  const [editContent, setEditContent] = createSignal(props.content.content);

  const updateSection = useUpdateSection(props.storyId, props.content.id);
  const deleteSection = useDeleteSection(props.storyId, props.content.id);

  // TODO: Replace with proper user object
  const getUserDisplayName = (user: {
    firstName: string;
    lastName: string;
  }) => {
    return `${user.firstName} ${user.lastName}`;
  };

  // TODO: Why are we passing in canEdit and checking here? 
  const canUserEdit = () => {
    return user && props.content.authorId === user.id && props.canEdit === true;
  };

  const handleEdit = () => {
    setIsEditing(true);
    setEditContent(props.content.content);
  };

  const handleSaveEdit = () => {
    // Remove HTML tags for validation
    if (isHtmlEmpty(editContent())) return;
    updateSection.mutate(
      { content: editContent() },
      {
        onSuccess: () => {
          setIsEditing(false);
        },
      },
    );
  };

  const handleCancelEdit = () => {
    setIsEditing(false);
    setEditContent(props.content.content);
  };

  const handleDelete = () => {
    if (
      confirm(
        "Are you sure you want to delete this section? This action cannot be undone.",
      )
    ) {
      deleteSection.mutate();
    }
  };

  return (
    <Card
      variant={props.isLatest ? "primary" : "default"}
      class={`${styles.storySection} ${props.class || ""}`}
    >
      <Avatar
        name={getUserDisplayName(props.content.author)}
        variant={((props.index % 5) + 1) as 1 | 2 | 3 | 4 | 5}
        size="md"
        class={styles.sectionAvatar}
      />
      <div class={styles.sectionContent}>
        <div class={styles.sectionHeader}>
          <div class={styles.authorName}>
            {getUserDisplayName(props.content.author)}
          </div>
          <Show when={canUserEdit() && !isEditing()}>
            <div class={styles.sectionActions}>
              <Button
                variant="tertiary"
                size="sm"
                onClick={handleEdit}
                disabled={updateSection.isPending}
              >
                ✏️ Edit
              </Button>
              <Button
                variant="danger"
                size="sm"
                onClick={handleDelete}
                disabled={deleteSection.isPending}
              >
                🗑️ Delete
              </Button>
            </div>
          </Show>
        </div>

        <Show when={!isEditing()}>
          <div class={styles.sectionText} innerHTML={props.content.content}></div>
        </Show>

        <Show when={isEditing()}>
          <div class={styles.editForm}>
            <TextEditor
              class={styles.editTextarea}
              initialValue={editContent()}
              onChange={setEditContent}
              placeholder="Edit your section..."
              minHeight="100px"
            />
            <div class={styles.editActions}>
              <Button
                variant="primary"
                size="sm"
                onClick={handleSaveEdit}
                disabled={
                  updateSection.isPending || isHtmlEmpty(editContent())
                }
              >
                {updateSection.isPending ? "Saving..." : "Save"}
              </Button>
              <Button
                variant="secondary"
                size="sm"
                onClick={handleCancelEdit}
                disabled={updateSection.isPending}
              >
                Cancel
              </Button>
            </div>
          </div>
        </Show>

        <div class={styles.sectionMeta}>{relativeTime()}</div>
      </div>
    </Card>
  );
};

export default StorySection;
