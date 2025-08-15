import { Outlet } from "react-router-dom";
import { Navbar } from "./components/layout/Navbar";

export default function App() {
  return (
    <div style={{ minHeight: "100vh", background: "#0b0e14", color: "#e6e6e6" }}>
      <Navbar />
      <main style={{ maxWidth: 960, margin: "2rem auto", padding: "0 1rem" }}>
        <Outlet />
      </main>
    </div>
  );
}
