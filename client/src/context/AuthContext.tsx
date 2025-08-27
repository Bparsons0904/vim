import { initializeTokenInterceptor } from "@services/api/api.service";
import { useNavigate } from "@solidjs/router";
import {
  createContext,
  useContext,
  createSignal,
  JSX,
  Accessor,
  createEffect,
  Show,
} from "solid-js";
import { createStore } from "solid-js/store";
import { User, LoginRequest, RegisterRequest } from "src/types/User";
import { apiClient } from "@services/api/client";

type AuthContextValue = {
  isAuthenticated: Accessor<boolean | null>;
  user: User | null;
  authToken: Accessor<string | null>;
  login: (credentials: LoginRequest) => Promise<void>;
  register: (userData: RegisterRequest) => Promise<void>;
  logout: () => void;
};

const AuthContext = createContext<AuthContextValue>({} as AuthContextValue);

export function AuthProvider(props: { children: JSX.Element }) {
  const navigate = useNavigate();
  const [user, setUser] = createStore<User | null>(null);
  const [isAuthenticated, setIsAuthenticated] = createSignal<boolean | null>(null);
  const [authToken, setAuthToken] = createSignal<string | null>(null);

  initializeTokenInterceptor(setAuthToken);

  const meQuery = apiClient.auth.me();

  createEffect(() => {
    if (meQuery.isSuccess && meQuery.data?.user) {
      setUser(meQuery.data.user);
      setIsAuthenticated(true);
    } else if (
      meQuery.isError ||
      (meQuery.isSuccess && !meQuery.data?.user)
    ) {
      setIsAuthenticated(false);
      setUser(null);
    }
  });

  const login = async (credentials: LoginRequest) => {
    try {
      const loginMutation = apiClient.auth.login()();
      const result = await loginMutation.mutateAsync(credentials);
      if (!result.user) return;
      setUser(result.user);
      setAuthToken(result.token);
      setIsAuthenticated(!!result.user);
      navigate("/");
    } catch {
      // Error handling is done in the mutation hook
    }
  };

  const register = async (userData: RegisterRequest) => {
    try {
      const registerMutation = apiClient.users.create()();
      const user = await registerMutation.mutateAsync(userData);
      if (!user) return;
      setUser(user);
      setIsAuthenticated(!!user);
      navigate("/");
    } catch {
      // Error handling is done in the mutation hook
    }
  };

  const logout = async () => {
    try {
      const logoutMutation = apiClient.auth.logout()();
      await logoutMutation.mutateAsync();
      setUser(null);
      setIsAuthenticated(false);
      setAuthToken(null);
      navigate("/login");
    } catch {
      // Still clear local state even if server logout fails
      setUser(null);
      setIsAuthenticated(false);
      setAuthToken(null);
      navigate("/login");
    }
  };

  return (
    <AuthContext.Provider
      value={{
        isAuthenticated,
        user,
        login,
        register,
        logout,
        authToken,
      }}
    >
      <Show when={isAuthenticated() !== null}>{props.children}</Show>
    </AuthContext.Provider>
  );
}

export function useAuth() {
  return useContext(AuthContext);
}

