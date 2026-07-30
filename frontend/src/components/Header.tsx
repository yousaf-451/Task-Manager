import { useEffect, useRef, useState } from "react";
import type { View } from "./Sidebar";
import type { User } from "../types/task";

interface HeaderProps {
  view: View;
  onAddTask: () => void;
  onOpenSidebar: () => void;
  user: User;
  onLogout: () => void;
  onDeleteAccount: () => void;
}

const COPY: Record<View, { eyebrow: string; title: string; subtitle: string }> = {
  dashboard: {
    eyebrow: "Granet Technologies",
    title: "Dashboard",
    subtitle: "A snapshot of everything on your plate right now.",
  },
  tasks: {
    eyebrow: "Granet Technologies",
    title: "Tasks",
    subtitle: "Track what's pending, in progress, and done — search, filter, and sort as your list grows.",
  },
};

export function Header({ view, onAddTask, onOpenSidebar, user, onLogout, onDeleteAccount }: HeaderProps) {
  const copy = COPY[view];
  const [menuOpen, setMenuOpen] = useState(false);
  const menuRef = useRef<HTMLDivElement>(null);

  // Close the account menu on an outside click or Escape, same pattern as
  // the Modal component's dismiss behavior.
  useEffect(() => {
    if (!menuOpen) return;

    function handlePointerDown(e: MouseEvent) {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
        setMenuOpen(false);
      }
    }
    function handleKeyDown(e: KeyboardEvent) {
      if (e.key === "Escape") setMenuOpen(false);
    }

    document.addEventListener("mousedown", handlePointerDown);
    document.addEventListener("keydown", handleKeyDown);
    return () => {
      document.removeEventListener("mousedown", handlePointerDown);
      document.removeEventListener("keydown", handleKeyDown);
    };
  }, [menuOpen]);

  return (
    <header className="app-header">
      <div className="app-header__left">
        <button
          className="app-header__menu-btn"
          onClick={onOpenSidebar}
          aria-label="Open navigation menu"
          type="button"
        >
          <svg width="20" height="20" viewBox="0 0 20 20" fill="none" aria-hidden="true">
            <path d="M3 5.5h14M3 10h14M3 14.5h14" stroke="currentColor" strokeWidth="1.6" strokeLinecap="round" />
          </svg>
        </button>
        <div>
          <p className="app-header__eyebrow">{copy.eyebrow}</p>
          <h1 className="app-header__title">{copy.title}</h1>
          <p className="app-header__subtitle">{copy.subtitle}</p>
        </div>
      </div>
      <div className="app-header__right">
        <div className="account-menu" ref={menuRef}>
          <button
            type="button"
            className="btn btn-ghost account-menu__trigger"
            onClick={() => setMenuOpen((v) => !v)}
            aria-haspopup="menu"
            aria-expanded={menuOpen}
          >
            <span className="account-menu__avatar">{user.name.charAt(0).toUpperCase()}</span>
            {user.name}
          </button>
          {menuOpen && (
            <div className="account-menu__panel" role="menu">
              <p className="account-menu__email">{user.email}</p>
              <button
                type="button"
                className="account-menu__item"
                role="menuitem"
                onClick={() => {
                  setMenuOpen(false);
                  onLogout();
                }}
              >
                Log out
              </button>
              <button
                type="button"
                className="account-menu__item account-menu__item--danger"
                role="menuitem"
                onClick={() => {
                  setMenuOpen(false);
                  onDeleteAccount();
                }}
              >
                Delete account
              </button>
            </div>
          )}
        </div>
        <button className="btn btn-accent" onClick={onAddTask}>
          + Add task <span className="app-header__shortcut">N</span>
        </button>
      </div>
    </header>
  );
}
