import { Component } from "solid-js";
import styles from "./Avatar.module.scss";

export interface AvatarProps {
  name: string;
  size?: "sm" | "md" | "lg";
  variant?: 1 | 2 | 3 | 4 | 5;
  class?: string;
}

export const Avatar: Component<AvatarProps> = (props) => {
  const getInitial = () => {
    return props.name.charAt(0).toUpperCase();
  };

  const sizeClass = () => {
    switch (props.size) {
      case "sm": return styles.small;
      case "lg": return styles.large;
      default: return styles.medium;
    }
  };

  const variantClass = () => {
    return styles[`avatar${props.variant || 1}`];
  };

  return (
    <div 
      class={`${styles.avatar} ${sizeClass()} ${variantClass()} ${props.class || ""}`}
      title={props.name}
    >
      {getInitial()}
    </div>
  );
};

export default Avatar;