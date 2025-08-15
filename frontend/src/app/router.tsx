import { lazy, Suspense } from "react";
import { createBrowserRouter, RouterProvider } from "react-router-dom";
import App from "../App";

const Home = lazy(() => import("../pages/Home"));
const Play = lazy(() => import("../pages/Play"));

const router = createBrowserRouter([
  {
    path: "/",
    element: <App />,
    children: [
      { index: true, element: <Home /> },
      { path: "play", element: <Play /> }
    ]
  }
]);

export default function AppRouter() {
  return (
    <Suspense fallback={<div style={{ padding: 24 }}>Loading…</div>}>
      <RouterProvider router={router} />
    </Suspense>
  );
}
