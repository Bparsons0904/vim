import { Component, createSignal } from "solid-js";
import { A } from "@solidjs/router";
import { Button } from "@components/common/ui/Button/Button";
import { useAuth } from "@context/AuthContext";
import styles from "./Home.module.scss";

const Home: Component = () => {
  const { isAuthenticated } = useAuth();
  const [activeSegment, setActiveSegment] = createSignal(0);

  const storySegments = [
    {
      author: "Sarah",
      avatar: "S",
      text: "The old lighthouse keeper had been warning ships away from the rocky coast for forty years, but tonight something was different. As the storm clouds gathered, he noticed a strange blue light pulsing from beneath the waves...",
    },
    {
      author: "Jake",
      avatar: "J",
      text: "Marcus grabbed his grandfather's telescope and peered into the churning water. The blue light wasn't just glowing—it was moving in patterns, almost like it was trying to communicate. Suddenly, the lighthouse beam flickered and died...",
    },
    {
      author: "Alex",
      avatar: "A",
      text: "In the darkness, Marcus heard footsteps on the spiral staircase behind him. But that was impossible—he was alone in the lighthouse. The blue light below pulsed brighter, and he realized with growing dread that it was responding to the approaching footsteps...",
    },
  ];

  const socialCards = [
    {
      title: "Creative Collaboration",
      description:
        "Connect with writers and storytellers. Share ideas, get feedback, and collaborate on amazing projects.",
      placeholder: "Writers collaborating together",
    },
    {
      title: "Community Driven",
      description:
        "Join a vibrant community of creators. Participate in writing challenges and collaborative storytelling.",
      placeholder: "Community members sharing stories",
    },
    {
      title: "Personal Expression",
      description:
        "Express your creativity through interactive storytelling. Build your portfolio and share your voice.",
      placeholder: "Individual creating content",
    },
    {
      title: "Global Connection",
      description:
        "Connect with creators worldwide. Share stories across cultures and build lasting creative relationships.",
      placeholder: "Global community of creators",
    },
  ];

  const handleSegmentClick = (index: number) => {
    setActiveSegment(index);
    if (index === storySegments.length) {
      // "Your turn" segment clicked
      setTimeout(() => {
        alert(
          "Ready to join Billy Wu? Sign up to start your creative journey!",
        );
      }, 300);
    }
  };

  return (
    <div class={styles.homePage}>
      {/* Hero Section */}
      <section class={styles.hero}>
        <div class={styles.container}>
          <h1 class={styles.heroTitle}>Welcome to Billy Wu</h1>
          <p class={styles.heroSubtitle}>
            Experience interactive storytelling and collaborative creativity in
            a modern platform designed for writers, friends, and communities.
          </p>
          <div class={styles.heroCta}>
            <A href={isAuthenticated() ? "/" : "/login"} class={styles.btnLink}>
              <Button variant="gradient" size="lg">
                {isAuthenticated() ? "Continue Your Story" : "Start Your Story"}
              </Button>
            </A>
            <Button
              variant="ghost"
              size="lg"
              onClick={() => {
                document
                  .getElementById("demo")
                  ?.scrollIntoView({ behavior: "smooth" });
              }}
            >
              See How It Works
            </Button>
          </div>
          <div class={styles.heroImage}>
            [Friends collaborating around a table - placeholder image]
          </div>
        </div>
      </section>

      {/* Story Demo Section */}
      <section class={styles.storyDemo} id="demo">
        <div class={styles.container}>
          <h2 class={styles.sectionTitle}>Watch a Story Unfold</h2>
          <div class={styles.storyChain}>
            {storySegments.map((segment, index) => (
              <div
                class={`${styles.storySegment} ${activeSegment() === index ? styles.active : ""}`}
                onClick={() => handleSegmentClick(index)}
              >
                <div class={styles.avatar}>{segment.avatar}</div>
                <div class={styles.storyContent}>
                  <div class={styles.authorName}>{segment.author}</div>
                  <div class={styles.storyText}>{segment.text}</div>
                </div>
              </div>
            ))}

            <div
              class={`${styles.storySegment} ${styles.yourTurn} ${activeSegment() === storySegments.length ? styles.active : ""}`}
              onClick={() => handleSegmentClick(storySegments.length)}
            >
              <div class={styles.avatar}>?</div>
              <div class={styles.storyContent}>
                <div class={styles.authorName}>Your Turn!</div>
                <div class={styles.storyText}>
                  What happens next? Click to add your part to the story...
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* Social Fun Section */}
      <section class={styles.socialFun}>
        <div class={styles.container}>
          <h2 class={styles.sectionTitle}>Built for Everyone</h2>
          <div class={styles.socialGrid}>
            {socialCards.map((card) => (
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
          <h2 class={styles.footerTitle}>Ready to Join Billy Wu?</h2>
          <p class={styles.footerSubtitle}>
            Join a creative community where stories come to life and connections
            are made
          </p>
          <A href={isAuthenticated() ? "/" : "/login"} class={styles.btnLink}>
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
