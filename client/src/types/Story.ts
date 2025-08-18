import { User } from "./User";
import { Genre } from "./Genre";
import { StoryOrderType } from "./StoryOrderType";

export interface Story {
  id: string;
  title: string;
  description: string;
  ownerId: string;
  owner: User;
  shortLink: string;
  genreId: number;
  genre: Genre;
  allowInvites: boolean;
  storyContent: StoryContent[];
  storyAuthors: StoryAuthor[];
  storyInvites: StoryInvite[];
  nextUpId?: string;
  nextUp?: User;
  storyOrderTypeId: number;
  storyOrderType: StoryOrderType;
  createdAt: string;
  updatedAt: string;
  deletedAt?: string;
}

export interface StoryContent {
  id: string;
  content: string;
  authorId: string;
  author: User;
  storyId: string;
  story: Story;
  createdAt: string;
  updatedAt: string;
  deletedAt?: string;
}

export interface StoryAuthor {
  id: string;
  storyId: string;
  authorId: string;
  story: Story;
  author: User;
  order: number;
  createdAt: string;
  updatedAt: string;
  deletedAt?: string;
}

export interface StoryInvite {
  id: string;
  storyId: string;
  story: Story;
  inviterId: string;
  inviter: User;
  email: string;
  status: "pending" | "accepted" | "expired";
  expiresAt: string;
  userId?: string;
  user?: User;
  createdAt: string;
  updatedAt: string;
  deletedAt?: string;
}

