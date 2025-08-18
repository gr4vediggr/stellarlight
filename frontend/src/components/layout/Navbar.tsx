import { Link, NavLink } from "react-router-dom";

export function Navbar() {
  const linkStyle: React.CSSProperties = {
    padding: "0.5rem 0.75rem",
    borderRadius: 8,
    textDecoration: "none",
    color: "#e6e6e6"
  };

  return (
    <header style={{
      borderBottom: "1px solid #1b2330",
      background: "#11161f"
    }}>
      <nav style={{
        width: "100%",
        padding: "0.75rem 1rem",
        display: "flex", alignItems: "center", gap: "1rem"
      }}>
        <Link to="/" style={{ fontWeight: 700, letterSpacing: 0.5 }}>Stellarlight</Link>
        <NavLink to="/" style={linkStyle}>Home</NavLink>
        <NavLink to="/play" style={linkStyle}>Play</NavLink>
        <NavLink to="/new-game" style={linkStyle}>New Game</NavLink>
        <NavLink to="/join-game" style={linkStyle}>Join Game</NavLink>

      </nav>
    </header>
  );
}
