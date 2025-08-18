import { useMutation } from "@tanstack/solid-query";
import { postApi } from "./api.service";
import { User } from "src/types/User";
// import { useNavigate } from "@solidjs/router";

interface LoginCredentials {
  login: string;
  password: string;
}

export const useLogin = () => {
  // const navigate = useNavigate();
  return useMutation(() => ({
    mutationFn: (credentials: LoginCredentials) =>
      postApi<User, LoginCredentials>("users/login", credentials),
    // onSuccess: () => {
    //   navigate("/");
    // },
  }));
};

export const useRegister = () => {
  // const navigate = useNavigate();
  return useMutation(() => ({
    mutationFn: (credentials: LoginCredentials) =>
      postApi<User, LoginCredentials>("users/register", credentials),
    // onSuccess: () => {
    //   navigate("/");
    // },
  }));
};

export const useLogout = () => {
  // const navigate = useNavigate();
  return useMutation(() => ({
    mutationFn: () => postApi("users/logout", {}),
  }));
};
