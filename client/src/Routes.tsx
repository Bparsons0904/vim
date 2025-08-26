import { Route } from "@solidjs/router";
import { Component, lazy } from "solid-js";
import { useAuth } from "@context/AuthContext";
import { useNavigate } from "@solidjs/router";
import { createEffect } from "solid-js";

const LoginPage = lazy(() => import("@pages/Auth/Login"));
const RegisterPage = lazy(() => import("@pages/Auth/Register"));
const ProfilePage = lazy(() => import("@pages/Profile/Profile"));
const DashboardPage = lazy(() => import("@pages/Dashboard/Dashboard"));
const LandingPage = lazy(() => import("@pages/Landing/Landing"));
const WorkstationComponent = lazy(() => import("@pages/Workstation/Workstation"));
const LoadTestPage = lazy(() => import("@pages/LoadTest/LoadTest"));

// Create a 7x7 workstation wrapper
const WorkstationPage: Component = () => {
  return <WorkstationComponent gridRows={7} gridCols={7} />;
};

const ProtectedRoute = (Component: Component) => {
  return () => {
    const { isAuthenticated } = useAuth();
    const navigate = useNavigate();

    createEffect(() => {
      if (isAuthenticated() === false) {
        navigate("/", { replace: true });
      }
    });

    return isAuthenticated() === true ? <Component /> : null;
  };
};

export const Routes: Component = () => {
  return (
    <>
      <Route path="/" component={LandingPage} />
      <Route path="/login" component={LoginPage} />
      <Route path="/register" component={RegisterPage} />
      <Route path="/dashboard" component={ProtectedRoute(DashboardPage)} />
      <Route path="/profile" component={ProtectedRoute(ProfilePage)} />
      <Route path="/workstation" component={ProtectedRoute(WorkstationPage)} />
      <Route path="/loadtest" component={ProtectedRoute(LoadTestPage)} />
    </>
  );
};
