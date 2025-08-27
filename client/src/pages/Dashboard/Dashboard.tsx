import { Component } from "solid-js";
import { useAuth } from "@context/AuthContext";
import styles from "./Dashboard.module.scss";
import { Button } from "@components/common/ui/Button/Button";

const Dashboard: Component = () => {
  const { user } = useAuth();

  const handleAction = (_action: string) => {
    // Handle actions here
    // TODO: Implement specific action handling
  };

  return (
    <div class={styles.dashboard}>
      <div class={styles.container}>
        {/* Header Section */}
        <header class={styles.header}>
          <h1 class={styles.headerTitle}>
            Welcome back, {user?.firstName || "User"}!
          </h1>
          <p class={styles.headerSubtitle}>
            Your dashboard is ready for customization.
          </p>
        </header>

        {/* Main Content Section */}
        <section class={styles.section}>
          <h2 class={styles.sectionTitle}>
            Your Content
          </h2>

          <div class={styles.emptyState}>
            <div class={styles.emptyIcon}>🚀</div>
            <h3>Ready to get started</h3>
            <p>This is your clean dashboard ready for your application features!</p>
          </div>
        </section>

        {/* Actions Section */}
        <section class={styles.section}>
          <h2 class={styles.sectionTitle}>Quick Actions</h2>

          <div class={styles.creationSection}>
            <div
              class={styles.creationCard}
              onClick={() => handleAction("primary")}
            >
              <div class={styles.creationIcon}>⚡</div>
              <h3 class={styles.creationTitle}>Primary Action</h3>
              <p class={styles.creationDescription}>
                Get started with your main feature
              </p>
              <Button
                variant="tertiary"
                onClick={() => handleAction("primary")}
              >
                Get Started
              </Button>
            </div>

            <div
              class={`${styles.creationCard} ${styles.creationCardSecondary}`}
              onClick={() => handleAction("secondary")}
            >
              <div class={styles.creationIcon}>🛠️</div>
              <h3 class={styles.creationTitle}>Secondary Action</h3>
              <p class={styles.creationDescription}>
                Access additional features and tools
              </p>
              <Button
                variant="tertiary"
                onClick={() => handleAction("secondary")}
              >
                Explore
              </Button>
            </div>
          </div>
        </section>
      </div>
    </div>
  );
};

export default Dashboard;
