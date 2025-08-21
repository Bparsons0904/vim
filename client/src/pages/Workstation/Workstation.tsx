import { Component, createSignal, createEffect } from "solid-js";
import styles from "./Workstation.module.scss";
import { useVimMotions } from "@context/VimMotionContext";

export const Workstation: Component = () => {
  const { currentMotion, clearMotion } = useVimMotions();
  
  // Grid configuration
  const GRID_ROWS = 3;
  const GRID_COLS = 3;
  
  // Track position as row/column coordinates
  const [currentRow, setCurrentRow] = createSignal(0);
  const [currentCol, setCurrentCol] = createSignal(0);

  const moveSelection = (direction: string, count: number) => {
    let newRow = currentRow();
    let newCol = currentCol();

    console.log(`Moving ${direction} ${count} times from (${newRow}, ${newCol})`);

    for (let i = 0; i < count; i++) {
      const prevRow = newRow;
      const prevCol = newCol;
      
      switch (direction) {
        case "left":
          if (newCol > 0) newCol--;
          break;
        case "right":
          if (newCol < GRID_COLS - 1) newCol++;
          break;
        case "up":
          if (newRow > 0) newRow--;
          break;
        case "down":
          if (newRow < GRID_ROWS - 1) newRow++;
          break;
      }
      
      console.log(`Step ${i + 1}: (${prevRow}, ${prevCol}) -> (${newRow}, ${newCol})`);
    }

    console.log(`Final position: (${newRow}, ${newCol})`);
    setCurrentRow(newRow);
    setCurrentCol(newCol);
  };

  createEffect(() => {
    const motion = currentMotion();
    console.log("Motion:", motion);
    if (motion?.type === "movement") {
      moveSelection(motion.direction, motion.count);
      clearMotion(); // Clear the motion after processing
    }
  });

  return (
    <div class={styles.workstation}>
      <h1 class={styles.title}>Workstation</h1>
      <div class={styles.grid}>
        {Array.from({ length: GRID_ROWS }, (_, row) =>
          Array.from({ length: GRID_COLS }, (_, col) => {
            const isSelected = currentRow() === row && currentCol() === col;
            
            return (
              <div
                class={`${styles.gridCell} ${isSelected ? styles.selected : ""}`}
                data-position={`${row},${col}`}
              >
                Cell {row},{col}
              </div>
            );
          })
        ).flat()}
      </div>
      <div class={styles.debug}>
        <div>Current Position: Row {currentRow()}, Col {currentCol()}</div>
        <div>Current Motion: {currentMotion() ? JSON.stringify(currentMotion()) : "None"}</div>
      </div>
    </div>
  );
};

export default Workstation;

