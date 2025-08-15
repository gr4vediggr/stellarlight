import type  { ReactNode } from "react";

export default function Providers({ children }: { children: ReactNode }) {
  // space for future contexts (theme, i18n, query, etc.)
  return <>{children}</>;
}
