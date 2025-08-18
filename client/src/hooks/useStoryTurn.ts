import { createMemo } from "solid-js";
import { useAuth } from "@context/AuthContext";
import type { Story, StoryInvite, User } from "../types";

export interface StoryTurnInfo {
  currentAuthor: User | undefined;
  isUserTurn: boolean;
  isInvited: boolean;
}

export const useStoryTurn = (
  story: Story | undefined,
  invite?: StoryInvite,
): StoryTurnInfo => {
  const { user } = useAuth();

  const storyTurnInfo = createMemo(() => {
    // Guard against undefined story
    if (!story) {
      return {
        currentAuthor: undefined,
        isUserTurn: false,
        isInvited: !!invite,
      };
    }

    // Use the nextUp field directly from the story model if available
    let currentAuthor = story.nextUp;
    let isUserTurn = user?.id === story.nextUpId;

    // Fallback to old calculation if nextUp is not available
    if (!story.nextUpId && story.storyAuthors && story.storyAuthors.length > 0) {
      const authorOrder = (story.storyContent?.length || 0) % story.storyAuthors.length;
      currentAuthor = story.storyAuthors.find(
        (sa) => sa.order === authorOrder,
      )?.author;
      isUserTurn = user?.id === currentAuthor?.id;
    }

    return {
      currentAuthor,
      isUserTurn,
      isInvited: !!invite,
    };
  });

  return storyTurnInfo();
};
