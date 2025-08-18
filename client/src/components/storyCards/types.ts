import { Story, StoryInvite } from "../../types/Story";

export type { Story, StoryInvite };

export interface Participant {
  id: string;
  name: string;
  initials: string;
  avatarColor?: string;
}

export type StoryStatus =
  | "your-turn"
  | "waiting-for-user"
  | "waiting-for-others"
  | "invitation";

export interface LastContribution {
  author: string;
  excerpt: string;
}

export interface StoryCardProps {
  story?: Story;
  invite?: StoryInvite;
  onAction: (storyId: string, action: string) => void;
  class?: string;
}

export interface TurnStatusBadgeProps {
  story: Story;
  invite?: StoryInvite;
}

export interface ParticipantAvatarsProps {
  story: Story;
  maxVisible?: number;
  size?: number;
}
