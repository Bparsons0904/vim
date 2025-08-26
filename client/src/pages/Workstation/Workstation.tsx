import { Component, createSignal, createEffect, onMount, onCleanup } from "solid-js";
import styles from "./Workstation.module.scss";
import { useVimMotions } from "@context/VimMotionContext";

interface WorkstationProps {
  gridRows?: number;
  gridCols?: number;
}

type AttackType = 'drain' | 'overcharge' | 'stable';

interface CellState {
  attackType: AttackType;
  energyLevel: number; // -100 to 100, 0 is center/stable
}

export const Workstation: Component<WorkstationProps> = (props) => {
  const { currentMotion, clearMotion } = useVimMotions();
  
  // Grid configuration with defaults
  const GRID_ROWS = props.gridRows || 3;
  const GRID_COLS = props.gridCols || 3;
  
  // Track position as row/column coordinates
  const [currentRow, setCurrentRow] = createSignal(0);
  const [currentCol, setCurrentCol] = createSignal(0);
  
  // Attack simulation state
  const [attackedCell, setAttackedCell] = createSignal<{row: number, col: number} | null>(null);
  const [attackType, setAttackType] = createSignal<AttackType>('stable');
  const [energyProgress, setEnergyProgress] = createSignal(0); // 0 to 100 - how far the bar has filled
  let attackInterval: number | undefined;
  
  // User stats
  const [userHealth, setUserHealth] = createSignal(100); // 0-100
  const [userEnergy, setUserEnergy] = createSignal(100); // 0-100
  const [userLevel, setUserLevel] = createSignal(1);
  const [successfulDefenses, setSuccessfulDefenses] = createSignal(0);
  
  // Idle energy regeneration
  let lastActionTime = Date.now();
  let idleRegenInterval: number | undefined;
  
  // Refs for DOM elements
  let gridContainerRef: HTMLDivElement | undefined;
  let gridRef: HTMLDivElement | undefined;
  
  // Viewport configuration - more aggressive scrolling
  const BUFFER_ZONE = 1; // Number of cells from edge to trigger scrolling
  
  // Start a progressive attack that fills the bar over time
  const startAttack = () => {
    // Don't start new attack if one is already in progress
    if (attackedCell()) return;
    
    const randomRow = Math.floor(Math.random() * GRID_ROWS);
    const randomCol = Math.floor(Math.random() * GRID_COLS);
    const randomAttackType: AttackType = Math.random() > 0.5 ? 'drain' : 'overcharge';
    
    setAttackedCell({row: randomRow, col: randomCol});
    setAttackType(randomAttackType);
    setEnergyProgress(0);
    
    console.log(`Attack started: Cell (${randomRow}, ${randomCol}) - ${randomAttackType}`);
    
    // Gradually fill the bar over 30 seconds (will reach 100% if not stopped)
    const fillRate = 100 / 300; // 100% over 300 intervals (30 seconds at 100ms intervals)
    attackInterval = setInterval(() => {
      setEnergyProgress(prev => {
        const newProgress = prev + fillRate;
        if (newProgress >= 100) {
          // Attack reached maximum - cause damage
          stopAttack(false);
          return 100;
        }
        return newProgress;
      });
    }, 100); // Update every 100ms for smooth animation
  };
  
  // Stop the current attack (when user presses correct key)
  const stopAttack = (successful: boolean = false) => {
    if (attackInterval) {
      clearInterval(attackInterval);
      attackInterval = undefined;
    }
    
    if (successful) {
      console.log('Attack successfully stopped!');
      // Reward successful defense with energy recovery
      setUserEnergy(prev => Math.min(100, prev + 10));
      
      // Track successful defenses and level progression
      setSuccessfulDefenses(prev => {
        const newCount = prev + 1;
        // Level up every 5 successful defenses
        if (newCount % 5 === 0) {
          setUserLevel(prevLevel => prevLevel + 1);
          console.log(`Level up! Now level ${userLevel() + 1}`);
        }
        return newCount;
      });
      
      // Start next attack in 5 seconds
      setTimeout(startAttack, 5000);
    } else {
      // Attack reached 100% - damage the user
      console.log('Attack completed! User takes damage!');
      setUserHealth(prev => Math.max(0, prev - 10));
      
      // Start next attack in 5 seconds
      setTimeout(startAttack, 5000);
    }
    
    setAttackedCell(null);
    setAttackType('stable');
    setEnergyProgress(0);
  };
  
  // Handle user input for defending against attacks
  const handleDefenseAction = (action: 'insert' | 'absorb') => {
    const currentAttack = attackType();
    if (!attackedCell() || currentAttack === 'stable') return;
    
    // Check if correct action for current attack type
    const correctAction = (currentAttack === 'drain' && action === 'insert') || 
                         (currentAttack === 'overcharge' && action === 'absorb');
    
    if (correctAction) {
      // No upfront cost - just get the +10 energy reward for success
      lastActionTime = Date.now(); // Reset idle timer
      stopAttack(true);
    } else {
      // Wrong action - small energy penalty
      setUserEnergy(prev => Math.max(0, prev - 5));
      lastActionTime = Date.now(); // Reset idle timer
      console.log('Wrong action! Energy penalty.');
    }
  };
  
  // Start idle energy regeneration
  const startIdleRegeneration = () => {
    idleRegenInterval = setInterval(() => {
      const timeSinceLastAction = Date.now() - lastActionTime;
      
      // Give 1 energy point every 5 seconds of inactivity
      if (timeSinceLastAction >= 5000) {
        setUserEnergy(prev => {
          const newEnergy = Math.min(100, prev + 1);
          if (newEnergy !== prev) {
            console.log('Idle energy regeneration: +1 energy');
          }
          return newEnergy;
        });
        lastActionTime = Date.now(); // Reset timer after giving energy
      }
    }, 1000); // Check every second
  };
  
  // Start attack simulation when component mounts
  onMount(() => {
    // Start idle energy regeneration
    startIdleRegeneration();
    
    // Start first attack after 2 seconds
    setTimeout(startAttack, 2000);
  });
  
  // Cleanup intervals on unmount
  onCleanup(() => {
    if (idleRegenInterval) {
      clearInterval(idleRegenInterval);
    }
    if (attackInterval) {
      clearInterval(attackInterval);
    }
  });

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
    
    // Movement costs 1 energy per command (regardless of count)
    setUserEnergy(prev => Math.max(0, prev - 1));
    
    setCurrentRow(newRow);
    setCurrentCol(newCol);
    
    // Reset idle timer
    lastActionTime = Date.now();
    
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
    } else if (motion?.type === "action") {
      // Handle 'i' and 'a' key presses for defense
      if (motion.action === "insert") {
        handleDefenseAction('insert');
      } else if (motion.action === "append") {
        handleDefenseAction('absorb');
      }
      clearMotion();
    }
  });

  return (
    <div class={styles.workstation}>
      <h1 class={styles.title}>Workstation</h1>
      
      {/* User Stats */}
      <div class={styles.userStats}>
        <div class={styles.statBar}>
          <label class={styles.statLabel}>Health</label>
          <div class={styles.healthBar}>
            <div class={styles.healthBarTrack}>
              <div 
                class={styles.healthBarFill}
                style={{ "width": `${userHealth()}%` }}
              ></div>
            </div>
            <span class={styles.statValue}>{userHealth()}/100</span>
          </div>
        </div>
        
        <div class={styles.statBar}>
          <label class={styles.statLabel}>Energy</label>
          <div class={styles.energyBarStat}>
            <div class={styles.energyBarTrack}>
              <div 
                class={styles.energyBarStatFill}
                style={{ "width": `${userEnergy()}%` }}
              ></div>
            </div>
            <span class={styles.statValue}>{userEnergy()}/100</span>
          </div>
        </div>
        
        <div class={styles.statBar}>
          <label class={styles.statLabel}>Level</label>
          <div class={styles.levelDisplay}>
            <span class={styles.levelValue}>{userLevel()}</span>
            <span class={styles.levelProgress}>({successfulDefenses() % 5}/5 to next)</span>
          </div>
        </div>
      </div>

      <div class={styles.gridWrapper}>
        {/* Column numbers header */}
        <div class={styles.columnHeaders}>
          <div class={styles.cornerSpace}></div>
          {Array.from({ length: GRID_COLS }, (_, col) => (
            <div class={styles.columnHeader}>{col}</div>
          ))}
        </div>
        
        {/* Grid with row numbers */}
        <div class={styles.gridWithRows}>
          <div class={styles.rowHeaders}>
            {Array.from({ length: GRID_ROWS }, (_, row) => (
              <div class={styles.rowHeader}>{row}</div>
            ))}
          </div>
          
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
            const isAttacked = attackedCell()?.row === row && attackedCell()?.col === col;
            const cellAttackType = isAttacked ? attackType() : 'stable';
            const cellEnergyProgress = isAttacked ? energyProgress() : 0;
            
            return (
              <div
                class={`${styles.gridCell} ${isSelected ? styles.selected : ""} ${isAttacked ? styles.attacked : ""}`}
                data-position={`${row},${col}`}
              >
                <div class={styles.cellContent}>
                  <div class={styles.energyBar}>
                    <div class={styles.energyBarTrack}>
                      <div class={styles.energyBarCenter}></div>
                      <div 
                        class={`${styles.energyBarFill} ${
                          cellAttackType === 'drain' ? styles.drain : 
                          cellAttackType === 'overcharge' ? styles.overcharge : 
                          styles.stable
                        }`}
                        style={{
                          "width": `${cellEnergyProgress}%`
                        }}
                      ></div>
                    </div>
                  </div>
                  {isAttacked && (
                    <div class={`${styles.attackIndicator} ${styles[cellAttackType]}`}>
                      {cellAttackType === 'drain' ? '⚡ DRAIN' : '🔥 OVERCHARGE'}
                    </div>
                  )}
                </div>
              </div>
            );
          })
        ).flat()}
            </div>
          </div>
        </div>
      </div>
      
      <div class={styles.debug}>
        <div>Current Position: Row {currentRow()}, Col {currentCol()}</div>
        <div>Current Motion: {currentMotion() ? JSON.stringify(currentMotion()) : "None"}</div>
        <div>Attack Status: {attackedCell() ? `Cell (${attackedCell()?.row}, ${attackedCell()?.col}) - ${attackType()} (${Math.round(energyProgress())}% filled)` : "No active attacks"}</div>
        <div>Action Needed: {attackedCell() && attackType() === 'drain' ? "Press 'i' to insert energy (+10 energy reward)" : attackedCell() && attackType() === 'overcharge' ? "Press 'a' to absorb energy (+10 energy reward)" : "System stable"}</div>
        <div>Game Rules: Move command = -1 energy, Successful defense = +10 energy, Idle 5sec = +1 energy, Failed defense = -10 health</div>
        <div>User Stats: Health {userHealth()}/100, Energy {userEnergy()}/100, Level {userLevel()} ({successfulDefenses()} defenses)</div>
      </div>
    </div>
  );
};

export default Workstation;

