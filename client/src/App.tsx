import "./App.scss";
import { Component } from "solid-js";
import { AuthProvider } from "./context/AuthContext";
import { QueryClient, QueryClientProvider } from "@tanstack/solid-query";
import { WebSocketProvider } from "@context/WebSocketContext";
import { ToastProvider } from "./context/ToastContext";
import { VimMotionProvider } from "./context/VimMotionContext";
import { RouteSectionProps } from "@solidjs/router";
import { NavBar } from "@components/layout/Navbar/Navbar";
import { useAutoCacheInvalidation } from "./services/cacheInvalidation.service";

const App: Component<RouteSectionProps<unknown>> = (props) => {
  const queryClient = new QueryClient();

  return (
    <QueryClientProvider client={queryClient}>
      <ToastProvider>
        <AuthProvider>
          <WebSocketProvider>
            <VimMotionProvider>
              <CacheInvalidationProvider />
              <NavBar />
              <main class="content">{props.children}</main>
            </VimMotionProvider>
          </WebSocketProvider>
        </AuthProvider>
      </ToastProvider>
    </QueryClientProvider>
  );
};

const CacheInvalidationProvider: Component = () => {
  useAutoCacheInvalidation();
  return null;
};

export default App;
