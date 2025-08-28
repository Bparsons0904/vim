import { useMutation } from "@tanstack/solid-query";
import { postApi } from "./api.service";
import { User } from "src/types/User";

interface LoginCredentials {
  login: string;
  password: string;
}

export const useLogin = () => {
  return useMutation(() => ({
    mutationFn: (credentials: LoginCredentials) =>
      postApi<User, LoginCredentials>("users/login", credentials),
  }));
};

export const useRegister = () => {
  return useMutation(() => ({
    mutationFn: (credentials: LoginCredentials) =>
      postApi<User, LoginCredentials>("users/register", credentials),
  }));
};

export const useLogout = () => {
  return useMutation(() => ({
    mutationFn: () => postApi("users/logout", {}),
  }));
};
