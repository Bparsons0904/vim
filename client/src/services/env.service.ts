export interface Environment {
  apiUrl: string;
  wsUrl: string;
  environment: string;
  isProduction: boolean;
}

export const env: Environment = {
  apiUrl: import.meta.env.VITE_API_URL as string,
  wsUrl: import.meta.env.VITE_WS_URL as string,
  environment: import.meta.env.VITE_ENV as string,
  isProduction: import.meta.env.VITE_ENV === "production",
};
