import type { View } from "./Sidebar";

interface HeaderProps {
  view: View;
  onAddTask: () => void;
  onOpenSidebar: () => void;
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

export function Header({ view, onAddTask, onOpenSidebar }: HeaderProps) {
  const copy = COPY[view];

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
      <button className="btn btn-accent" onClick={onAddTask}>
        + Add task <span className="app-header__shortcut">N</span>
      </button>
    </header>
  );
}
