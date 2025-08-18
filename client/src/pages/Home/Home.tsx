import { Component } from "solid-js";
import { A } from "@solidjs/router";
import { Button } from "@components/common/ui/Button/Button";
import { useAuth } from "@context/AuthContext";
import styles from "./Home.module.scss";

const Home: Component = () => {
  const { isAuthenticated } = useAuth();

  const featureCards = [
    {
      title: "Modern Development",
      description:
        "Built with the latest technologies and best practices for optimal performance and developer experience.",
      placeholder: "Modern development tools",
    },
    {
      title: "Scalable Architecture",
      description:
        "Designed with scalability in mind, supporting growth from prototype to production.",
      placeholder: "Scalable system architecture",
    },
    {
      title: "User-Centered Design",
      description:
        "Focused on providing an excellent user experience with intuitive interfaces and smooth interactions.",
      placeholder: "User-centered design",
    },
    {
      title: "Community Driven",
      description:
        "Open to collaboration and built to serve the needs of its community.",
      placeholder: "Community collaboration",
    },
  ];

  return (
    <div class={styles.homePage}>
      {/* Hero Section */}
      <section class={styles.hero}>
        <div class={styles.container}>
          <h1 class={styles.heroTitle}>Welcome to Our Platform</h1>
          <p class={styles.heroSubtitle}>
            A modern, scalable platform built with the latest technologies
            and designed for exceptional user experiences.
          </p>
          <div class={styles.heroCta}>
            <A href={isAuthenticated() ? "/dashboard" : "/login"} class={styles.btnLink}>
              <Button variant="gradient" size="lg">
                {isAuthenticated() ? "Go to Dashboard" : "Get Started"}
              </Button>
            </A>
          </div>
          <div class={styles.heroImage}>
            [Platform overview - placeholder image]
          </div>
        </div>
      </section>

      {/* Features Section */}
      <section class={styles.socialFun}>
        <div class={styles.container}>
          <h2 class={styles.sectionTitle}>Built for Excellence</h2>
          <div class={styles.socialGrid}>
            {featureCards.map((card) => (
              <div class={styles.socialCard}>
                <div class={styles.socialCardImage}>
                  [Image: {card.placeholder}]
                </div>
                <h3 class={styles.socialCardTitle}>{card.title}</h3>
                <p class={styles.socialCardDescription}>{card.description}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Footer CTA */}
      <section class={styles.footerCta}>
        <div class={styles.container}>
          <h2 class={styles.footerTitle}>Ready to Get Started?</h2>
          <p class={styles.footerSubtitle}>
            Join our platform and experience the difference modern development makes.
          </p>
          <A href={isAuthenticated() ? "/dashboard" : "/login"} class={styles.btnLink}>
            <Button variant="gradient" size="lg">
              {isAuthenticated()
                ? "Go to Dashboard"
                : "Get Started - It's Free!"}
            </Button>
          </A>
        </div>
      </section>
    </div>
  );
};

export default Home;
