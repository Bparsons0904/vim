import { Component, Match, Switch } from "solid-js";
import { A } from "@solidjs/router";
import styles from "./NavBar.module.scss";
import { useAuth } from "@context/AuthContext";

export const NavBar: Component = () => {
  const { isAuthenticated, logout } = useAuth();

  return (
    <>
      <nav class={styles.navbar}>
        <div class={styles.navbarContainer}>
          <div class={styles.navbarLogo}>
            <A href="/" class={styles.navbarTitle}>
              VIM Wars
            </A>
          </div>
          <div class={styles.navbarMenu}>
            <ul class={styles.navbarItems}>
              <li class={styles.navbarItem}>
                <A
                  href="/"
                  class={styles.navbarLink}
                  activeClass={styles.active}
                  end
                >
                  Home
                </A>
              </li>
              <Switch>
                <Match when={isAuthenticated()}>
                  <li class={styles.navbarItem}>
                    <A
                      href="/workstation"
                      class={styles.navbarLink}
                      activeClass={styles.active}
                    >
                      Workstation
                    </A>
                  </li>
                  <li class={styles.navbarItem}>
                    <A
                      href="/loadtest"
                      class={styles.navbarLink}
                      activeClass={styles.active}
                    >
                      Load Test
                    </A>
                  </li>
                  <li class={styles.navbarItem}>
                    <A
                      href="/profile"
                      class={styles.navbarLink}
                      activeClass={styles.active}
                    >
                      Profile
                    </A>
                  </li>
                  {/* <li class={styles.navbarItem}> */}
                  {/*   <A */}
                  {/*     href="/story/1" */}
                  {/*     class={styles.navbarLink} */}
                  {/*     activeClass={styles.active} */}
                  {/*   > */}
                  {/*     Story */}
                  {/*   </A> */}
                  {/* </li> */}
                </Match>
              </Switch>
              <li class={styles.navbarItem}>
                <Switch>
                  <Match when={!isAuthenticated()}>
                    <A
                      href="/login"
                      class={styles.navbarLink}
                      activeClass={styles.active}
                    >
                      Login
                    </A>
                  </Match>
                  <Match when={isAuthenticated()}>
                    <A href="/" class={styles.navbarLink} onClick={logout}>
                      Logout
                    </A>
                  </Match>
                </Switch>
              </li>
            </ul>
          </div>
        </div>
      </nav>
    </>
  );
};
