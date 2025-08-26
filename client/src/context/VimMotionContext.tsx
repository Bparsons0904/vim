import {
  createContext,
  useContext,
  ParentComponent,
  createSignal,
  onMount,
  onCleanup,
} from "solid-js";

export interface MovementMotion {
  type: "movement";
  direction: "up" | "down" | "left" | "right";
  count: number;
}

export interface ActionMotion {
  type: "action";
  action: "insert" | "append";
}

export type VimMotion = MovementMotion | ActionMotion;

interface VimMotionContextValue {
  currentMotion: () => VimMotion | null;
  clearMotion: () => void;
}

const VimMotionContext = createContext<VimMotionContextValue>();

export const VimMotionProvider: ParentComponent = (props) => {
  const [currentMotion, setCurrentMotion] = createSignal<VimMotion | null>(
    null,
  );
  const [keySequence, setKeySequence] = createSignal<string>("");
  const [sequenceTimeout, setSequenceTimeout] = createSignal<number | null>(
    null,
  );

  const resetSequence = () => {
    setKeySequence("");
    const timeout = sequenceTimeout();
    if (timeout) {
      clearTimeout(timeout);
      setSequenceTimeout(null);
    }
  };

  const parseKeySequence = (sequence: string): VimMotion | null => {
    // Check for action keys first (single keys)
    if (sequence === "i") {
      return { type: "action", action: "insert" };
    }
    if (sequence === "a") {
      return { type: "action", action: "append" };
    }
    
    // Check for movement keys (with optional count)
    const match = sequence.match(/^(\d*)([hjkl])$/);
    if (!match) return null;

    const [, countStr, key] = match;
    const count = countStr ? parseInt(countStr, 10) : 1;

    const directionMap: Record<string, MovementMotion["direction"]> = {
      h: "left",
      j: "down",
      k: "up",
      l: "right",
    };

    return {
      type: "movement",
      direction: directionMap[key],
      count,
    };
  };

  const handleKeyDown = (event: KeyboardEvent) => {
    const key = event.key.toLowerCase();

    // Only handle vim motion and action keys
    if (
      ![
        "h",
        "j",
        "k",
        "l",
        "i",
        "a",
        "1",
        "2",
        "3",
        "4",
        "5",
        "6",
        "7",
        "8",
        "9",
      ].includes(key)
    ) {
      return;
    }

    // Prevent default behavior for vim keys
    event.preventDefault();

    const currentSequence = keySequence();
    const newSequence = currentSequence + key;
    setKeySequence(newSequence);

    // Try to parse the current sequence
    const motion = parseKeySequence(newSequence);

    if (motion) {
      setCurrentMotion(motion);
      resetSequence();
    } else {
      // Set timeout to reset incomplete sequence
      const timeout = sequenceTimeout();
      if (timeout) clearTimeout(timeout);

      const newTimeout = window.setTimeout(() => {
        resetSequence();
      }, 1000);
      setSequenceTimeout(newTimeout);
    }
  };

  onMount(() => {
    document.addEventListener("keydown", handleKeyDown);
  });

  onCleanup(() => {
    document.removeEventListener("keydown", handleKeyDown);
    const timeout = sequenceTimeout();
    if (timeout) clearTimeout(timeout);
  });

  const clearMotion = () => {
    setCurrentMotion(null);
  };

  return (
    <VimMotionContext.Provider value={{ currentMotion, clearMotion }}>
      {props.children}
    </VimMotionContext.Provider>
  );
};

export const useVimMotions = () => {
  const context = useContext(VimMotionContext);
  if (!context) {
    throw new Error("useVimMotions must be used within a VimMotionProvider");
  }
  return context;
};

