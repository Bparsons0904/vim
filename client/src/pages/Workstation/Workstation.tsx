import { Component, createSignal, createEffect, onMount } from "solid-js";
import styles from "./Workstation.module.scss";
import { useVimMotions } from "@context/VimMotionContext";

interface WorkstationProps {
  gridRows?: number;
  gridCols?: number;
}

export const Workstation: Component<WorkstationProps> = (props) => {
  const { currentMotion, clearMotion } = useVimMotions();
  
  // Grid configuration with defaults
  const GRID_ROWS = props.gridRows || 3;
  const GRID_COLS = props.gridCols || 3;
  
  // Track position as row/column coordinates
  const [currentRow, setCurrentRow] = createSignal(0);
  const [currentCol, setCurrentCol] = createSignal(0);
  
  // Refs for DOM elements
  let gridContainerRef: HTMLDivElement | undefined;
  let gridRef: HTMLDivElement | undefined;
  
  // Viewport configuration - more aggressive scrolling
  const BUFFER_ZONE = 1; // Number of cells from edge to trigger scrolling

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
    
    // Immediately scroll - no delay for more aggressive response
    scrollToKeepInView(newRow, newCol);
  };
  
  const scrollToKeepInView = (row: number, col: number) => {
    if (!gridContainerRef || !gridRef) return;
    
    const containerRect = gridContainerRef.getBoundingClientRect();
    const gridRect = gridRef.getBoundingClientRect();
    
    // Calculate cell dimensions
    const cellWidth = gridRect.width / GRID_COLS;
    const cellHeight = gridRect.height / GRID_ROWS;
    
    // Calculate selected cell position relative to grid
    const cellLeft = col * cellWidth;
    const cellTop = row * cellHeight;
    const cellRight = cellLeft + cellWidth;
    const cellBottom = cellTop + cellHeight;
    
    // Calculate buffer zone dimensions
    const bufferWidth = BUFFER_ZONE * cellWidth;
    const bufferHeight = BUFFER_ZONE * cellHeight;
    
    // Get current scroll position
    const scrollLeft = gridContainerRef.scrollLeft;
    const scrollTop = gridContainerRef.scrollTop;
    
    // Calculate visible area within container
    const visibleLeft = scrollLeft;
    const visibleTop = scrollTop;
    const visibleRight = scrollLeft + containerRect.width;
    const visibleBottom = scrollTop + containerRect.height;
    
    // Calculate buffer zones within visible area
    const bufferLeft = visibleLeft + bufferWidth;
    const bufferTop = visibleTop + bufferHeight;
    const bufferRight = visibleRight - bufferWidth;
    const bufferBottom = visibleBottom - bufferHeight;
    
    let newScrollLeft = scrollLeft;
    let newScrollTop = scrollTop;
    
    // Check horizontal scrolling
    if (cellLeft < bufferLeft) {
      // Scroll left - position cell at buffer zone from left edge
      newScrollLeft = Math.max(0, cellLeft - bufferWidth);
    } else if (cellRight > bufferRight) {
      // Scroll right - position cell at buffer zone from right edge
      newScrollLeft = Math.min(
        gridRect.width - containerRect.width,
        cellRight - containerRect.width + bufferWidth
      );
    }
    
    // Check vertical scrolling
    if (cellTop < bufferTop) {
      // Scroll up - position cell at buffer zone from top edge
      newScrollTop = Math.max(0, cellTop - bufferHeight);
    } else if (cellBottom > bufferBottom) {
      // Scroll down - position cell at buffer zone from bottom edge
      newScrollTop = Math.min(
        gridRect.height - containerRect.height,
        cellBottom - containerRect.height + bufferHeight
      );
    }
    
    // Perform immediate scroll if needed - more aggressive
    if (newScrollLeft !== scrollLeft || newScrollTop !== scrollTop) {
      gridContainerRef.scrollTo({
        left: newScrollLeft,
        top: newScrollTop,
        behavior: 'auto' // Changed from 'smooth' to 'auto' for instant response
      });
      
      console.log(`Scrolling to: (${newScrollLeft}, ${newScrollTop}) for cell (${row}, ${col})`);
    }
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
      <div 
        ref={gridContainerRef}
        class={styles.gridContainer}
      >
        <div 
          ref={gridRef}
          class={styles.grid} 
          style={{
            "grid-template-columns": `repeat(${GRID_COLS}, 1fr)`,
            "grid-template-rows": `repeat(${GRID_ROWS}, 1fr)`
          }}
        >
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
      </div>
      <div class={styles.debug}>
        <div>Current Position: Row {currentRow()}, Col {currentCol()}</div>
        <div>Current Motion: {currentMotion() ? JSON.stringify(currentMotion()) : "None"}</div>
      </div>
    </div>
  );
};

export default Workstation;

