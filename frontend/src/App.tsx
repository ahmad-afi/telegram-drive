import { useEffect, useState } from "react";
import Login from "./Login";
import Drive from "./Drive";
import "./App.css";

export default function App() {
  const [authed, setAuthed] = useState(!!localStorage.getItem("tg_session"));
  const [dark, setDark] = useState(() => {
    const saved = localStorage.getItem("tg_dark");
    return saved ? saved === "1" : matchMedia("(prefers-color-scheme: dark)").matches;
  });

  useEffect(() => {
    document.documentElement.classList.toggle("dark", dark);
    localStorage.setItem("tg_dark", dark ? "1" : "0");
  }, [dark]);

  function logout() {
    localStorage.removeItem("tg_session");
    setAuthed(false);
  }

  return (
    <>
      <button className="theme-toggle" onClick={() => setDark(!dark)} aria-label="Toggle theme">
        {dark ? "☀" : "☾"}
      </button>
      {authed ? (
        <Drive onLogout={logout} />
      ) : (
        <Login onAuth={() => setAuthed(true)} />
      )}
    </>
  );
}
