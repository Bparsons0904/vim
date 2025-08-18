import { createSignal, onCleanup, onMount } from "solid-js";

export const useRelativeTime = (date: Date | string) => {
  const [relativeTime, setRelativeTime] = createSignal<string>("Just now");

  const formatRelativeTime = (targetDate: Date): string => {
    const now = new Date();
    const diff = now.getTime() - targetDate.getTime();
    const minutes = Math.floor(diff / 60000);
    const hours = Math.floor(minutes / 60);
    const days = Math.floor(hours / 24);

    if (minutes < 1) {
      return "Just now";
    } else if (minutes < 60) {
      return `${minutes}m ago`;
    } else if (hours < 24) {
      return `${hours}h ago`;
    } else if (days < 7) {
      return `${days}d ago`;
    } else {
      return targetDate.toLocaleDateString();
    }
  };

  const updateTime = () => {
    const targetDate = typeof date === "string" ? new Date(date) : date;
    setRelativeTime(formatRelativeTime(targetDate));
  };

  onMount(() => {
    updateTime(); // Initial update
    
    // Set up interval for updates every 30 seconds
    const interval = setInterval(updateTime, 30000);
    
    onCleanup(() => {
      clearInterval(interval);
    });
  });

  return relativeTime;
};