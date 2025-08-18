import { Component, For } from "solid-js";
import { Card } from "../../common/ui/Card/Card";
import { Avatar } from "../../common/ui/Avatar/Avatar";
import type { User } from "../../../types/User";
import styles from "./UsersList.module.scss";

export interface UsersListItem {
  id: string;
  user: User;
  status: string;
  isHighlighted?: boolean;
}

export interface UsersListProps {
  title: string;
  icon: string;
  items: UsersListItem[];
  class?: string;
}

export const UsersList: Component<UsersListProps> = (props) => {
  const getUserDisplayName = (user: User) => {
    return `${user.firstName} ${user.lastName}`;
  };

  return (
    <Card class={`${styles.sidebarSection} ${props.class || ""}`}>
      <h3 class={styles.sidebarTitle}>{props.icon} {props.title}</h3>
      <div class={styles.usersList}>
        <For each={props.items}>
          {(item, index) => (
            <div
              class={`${styles.user} ${item.isHighlighted ? styles.highlighted : ""}`}
            >
              <Avatar
                name={getUserDisplayName(item.user)}
                variant={((index() % 5) + 1) as 1 | 2 | 3 | 4 | 5}
                size="md"
                class={styles.userAvatar}
              />
              <div class={styles.userInfo}>
                <div class={styles.userName}>
                  {getUserDisplayName(item.user)}
                </div>
                <div class={styles.userStatus}>
                  {item.status}
                </div>
              </div>
            </div>
          )}
        </For>
      </div>
    </Card>
  );
};

export default UsersList;