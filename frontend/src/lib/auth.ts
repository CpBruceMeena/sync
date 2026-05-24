const TOKEN_KEY = "access_token";
const REFRESH_KEY = "refresh_token";

export const auth = {
  getAccessToken: (): string | null => {
    if (typeof window === "undefined") return null;
    return localStorage.getItem(TOKEN_KEY);
  },

  getRefreshToken: (): string | null => {
    if (typeof window === "undefined") return null;
    return localStorage.getItem(REFRESH_KEY);
  },

  setTokens: (accessToken: string, refreshToken: string) => {
    localStorage.setItem(TOKEN_KEY, accessToken);
    localStorage.setItem(REFRESH_KEY, refreshToken);
  },

  clearTokens: () => {
    localStorage.removeItem(TOKEN_KEY);
    localStorage.removeItem(REFRESH_KEY);
  },

  isAuthenticated: (): boolean => {
    return !!auth.getAccessToken();
  },
};
