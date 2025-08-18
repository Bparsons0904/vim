import { Component, Show } from "solid-js";
import { useAuth } from "@context/AuthContext";
import Home from "@pages/Home/Home";
import Dashboard from "@pages/Dashboard/Dashboard";

const Landing: Component = () => {
  const { isAuthenticated } = useAuth();

  return (
    <Show when={isAuthenticated()} fallback={<Home />}>
      <Dashboard />
    </Show>
  );
};

export default Landing;