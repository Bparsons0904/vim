import { Component, JSX } from "solid-js";
import { NavBar } from "@layout/Navbar/Navbar";

interface LayoutProps {
  children: JSX.Element;
}

export const Layout: Component<LayoutProps> = (props) => {
  return (
    <>
      <NavBar />
      <main class="content">{props.children}</main>
    </>
  );
};
