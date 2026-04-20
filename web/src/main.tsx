import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { QueryClientProvider } from "@tanstack/react-query";
import { createBrowserRouter, redirect, RouterProvider } from "react-router-dom";
import Root, { RootErrorBoundary } from "./routes/root";
import Home from "./routes/home";
import About from "./routes/about";
import { createQueryClient } from "./lib/query-client";
import "./locales/i18n"; // side-effect init
import "./index.css";

const qc = createQueryClient();
const router = createBrowserRouter([
  {
    path: "/",
    element: <Root />,
    errorElement: <RootErrorBoundary />,
    children: [
      { index: true, element: <Home /> },
      { path: "about", element: <About /> },
    ],
  },
  { path: "*", loader: () => redirect("/") },
]);

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <QueryClientProvider client={qc}>
      <RouterProvider router={router} />
    </QueryClientProvider>
  </StrictMode>,
);
