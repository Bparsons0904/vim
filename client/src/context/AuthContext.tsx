import { initializeTokenInterceptor } from "@services/api/api.service";
import { 
  useCurrentUser, 
  useLoginUser, 
  useRegisterUser, 
  useLogoutUser,
  LoginCredentials,
  RegisterData 
} from "@services/api/endpoints/users.api";
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
import { User } from "src/types/User";

type AuthContextValue = {
  isAuthenticated: Accessor<boolean | null>;
  user: User | null;
  authToken: Accessor<string | null>;
  login: (credentials: LoginCredentials) => Promise<void>;
  register: (userData: RegisterData) => Promise<void>;
  logout: () => void;
};

const AuthContext = createContext<AuthContextValue>({} as AuthContextValue);

export function AuthProvider(props: { children: JSX.Element }) {
  const navigate = useNavigate();
  const [user, setUser] = createStore(null);
  const [isAuthenticated, setIsAuthenticated] = createSignal(null);
  const [authToken, setAuthToken] = createSignal<string | null>(null);

  initializeTokenInterceptor(setAuthToken);

  const getUserResponse = useCurrentUser();

  createEffect(() => {
    // Handle the auth state based on results
    if (getUserResponse.isSuccess && getUserResponse.data?.user) {
      setUser(getUserResponse.data.user);
      setIsAuthenticated(true);
    } else if (
      getUserResponse.isError ||
      (getUserResponse.isSuccess && !getUserResponse.data?.user)
    ) {
      setIsAuthenticated(false);
      setUser(null);
    }
  });

  const loginMutation = useLoginUser();
  const registerMutation = useRegisterUser();
  const logoutMutation = useLogoutUser();

  const login = async (credentials: LoginCredentials) => {
    try {
      const user = await loginMutation.mutateAsync(credentials);
      if (!user) return;
      setUser(user);
      setIsAuthenticated(!!user);
      navigate("/");
    } catch (error) {
      console.error('Login failed:', error);
    }
  };

  const register = async (userData: RegisterData) => {
    try {
      const user = await registerMutation.mutateAsync(userData);
      if (!user) return;
      setUser(user);
      setIsAuthenticated(!!user);
      navigate("/");
    } catch (error) {
      console.error('Registration failed:', error);
    }
  };

  const logout = async () => {
    try {
      await logoutMutation.mutateAsync({});
      setUser(null);
      setIsAuthenticated(false);
      setAuthToken(null);
      navigate("/login");
    } catch (error) {
      console.error('Logout failed:', error);
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
        // setAuthToken,
      }}
    >
      <Show when={isAuthenticated() !== null}>{props.children}</Show>
    </AuthContext.Provider>
  );
}

export function useAuth() {
  return useContext(AuthContext);
}
