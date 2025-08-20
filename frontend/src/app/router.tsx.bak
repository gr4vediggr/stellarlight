import { lazy, Suspense } from "react";
import { createBrowserRouter, RouterProvider } from "react-router-dom";
import App from "../App";
import GalaxyCreatorV2 from "../pages/old/GalaxyCreatorV2";
import NewGame from "../pages/NewGame";

const Home = lazy(() => import("../pages/Home"));
const Play = lazy(() => import("../pages/old/Play"));
const GalaxyCreator = lazy(() => import("../pages/old/GalaxyCreator"));

const router = createBrowserRouter([
  {
    path: "/",
    element: <App />,
    children: [
      { index: true, element: <Home /> },
      { path: "play", element: <Play /> },
      { path: "galaxy-creator", element: <GalaxyCreator /> },
      { path: "galaxy-creator-v2", element: <GalaxyCreatorV2 /> },
      { path: "new-game", element: <NewGame /> },

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
