import { QueryCache, QueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import i18n from "@/locales/i18n";
import { ApiError } from "./api";

export function createQueryClient(): QueryClient {
  return new QueryClient({
    queryCache: new QueryCache({
      onError: (err) => {
        if (err instanceof ApiError && err.code === "rate_limited") {
          toast.error(i18n.t("error.rate_limited"));
        } else if (err instanceof ApiError && err.code === "network") {
          toast.error(i18n.t("error.network"));
        }
        // invalid_grid errors are handled inline by GridInput — don't toast those.
      },
    }),
    defaultOptions: {
      queries: {
        staleTime: Infinity,
        gcTime: 24 * 60 * 60 * 1000, // 24h
        retry: (attempt, err) => {
          if (err instanceof ApiError && (err.code === "invalid_grid" || err.code === "rate_limited")) {
            return false;
          }
          return attempt < 2;
        },
        retryDelay: (attempt) => Math.min(1000 * 2 ** attempt, 10_000),
        refetchOnWindowFocus: false,
      },
    },
  });
}
