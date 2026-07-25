export type View = "dashboard" | "tasks";

interface SidebarProps {
  view: View;
  onViewChange: (view: View) => void;
  open: boolean;
  onClose: () => void;
  taskCount: number;
}

interface NavItem {
  view: View;
  label: string;
  icon: JSX.Element;
}

function DashboardIcon() {
  return (
    <svg width="18" height="18" viewBox="0 0 18 18" fill="none" aria-hidden="true">
      <rect x="1.5" y="1.5" width="7" height="7" rx="1.5" stroke="currentColor" strokeWidth="1.4" />
      <rect x="9.5" y="1.5" width="7" height="4.5" rx="1.5" stroke="currentColor" strokeWidth="1.4" />
      <rect x="9.5" y="8" width="7" height="8.5" rx="1.5" stroke="currentColor" strokeWidth="1.4" />
      <rect x="1.5" y="10.5" width="7" height="6" rx="1.5" stroke="currentColor" strokeWidth="1.4" />
    </svg>
  );
}

function TasksIcon() {
  return (
    <svg width="18" height="18" viewBox="0 0 18 18" fill="none" aria-hidden="true">
      <path
        d="M3 5.2l1.4 1.4L6.8 4"
        stroke="currentColor"
        strokeWidth="1.4"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <path d="M9 4.5h6" stroke="currentColor" strokeWidth="1.4" strokeLinecap="round" />
      <path
        d="M3 10.7l1.4 1.4 2.4-2.6"
        stroke="currentColor"
        strokeWidth="1.4"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <path d="M9 10h6" stroke="currentColor" strokeWidth="1.4" strokeLinecap="round" />
      <rect x="3" y="14" width="1.8" height="1.8" rx="0.4" fill="currentColor" />
      <path d="M9 15.5h6" stroke="currentColor" strokeWidth="1.4" strokeLinecap="round" />
    </svg>
  );
}

export function Sidebar({ view, onViewChange, open, onClose, taskCount }: SidebarProps) {
  const items: NavItem[] = [
    { view: "dashboard", label: "Dashboard", icon: <DashboardIcon /> },
    { view: "tasks", label: "Tasks", icon: <TasksIcon /> },
  ];

  return (
    <>
      {open && <div className="sidebar-scrim" onClick={onClose} aria-hidden="true" />}
      <aside className={`sidebar ${open ? "sidebar--open" : ""}`} aria-label="Primary navigation">
        <div className="sidebar__brand">
          <span className="sidebar__brand-mark">GT</span>
          <div>
            <p className="sidebar__brand-name">Granet</p>
            <p className="sidebar__brand-sub">Task Manager</p>
          </div>
        </div>

        <nav className="sidebar__nav">
          {items.map((item) => (
            <button
              key={item.view}
              className={`sidebar__link ${view === item.view ? "sidebar__link--active" : ""}`}
              onClick={() => {
                onViewChange(item.view);
                onClose();
              }}
            >
              {item.icon}
              <span>{item.label}</span>
              {item.view === "tasks" && taskCount > 0 && (
                <span className="sidebar__count">{taskCount}</span>
              )}
            </button>
          ))}
        </nav>

        <div className="sidebar__profile">
          <span className="sidebar__avatar" aria-hidden="true">
            U
          </span>
          <div className="sidebar__profile-text">
            <p className="sidebar__profile-name">Guest User</p>
            <p className="sidebar__profile-role">Developer</p>
          </div>
        </div>
      </aside>
    </>
  );
}
