import type { Task, TaskStats } from "../types/task";
import { PRIORITY_LABELS, STATUS_LABELS } from "../types/task";

interface DashboardProps {
  stats: TaskStats | null;
  statsLoading: boolean;
  statsError: string | null;
  tasks: Task[];
  onViewTasks: () => void;
}

interface StatCardDef {
  label: string;
  value: number;
  accent: "ink" | "accent" | "amber" | "danger" | "slate";
  hint?: string;
}

function StatCard({ label, value, accent, hint }: StatCardDef) {
  return (
    <div className={`stat-card stat-card--${accent}`}>
      <p className="stat-card__label">{label}</p>
      <p className="stat-card__value">{value}</p>
      {hint && <p className="stat-card__hint">{hint}</p>}
    </div>
  );
}

/** A dependency-free horizontal bar chart built from plain divs. */
function BreakdownBar({
  segments,
}: {
  segments: { label: string; value: number; color: string }[];
}) {
  const total = segments.reduce((sum, s) => sum + s.value, 0);

  return (
    <div className="breakdown">
      <div className="breakdown__track">
        {total === 0 ? (
          <div className="breakdown__empty" />
        ) : (
          segments.map((s) => (
            <div
              key={s.label}
              className="breakdown__segment"
              style={{ width: `${(s.value / total) * 100}%`, background: s.color }}
              title={`${s.label}: ${s.value}`}
            />
          ))
        )}
      </div>
      <ul className="breakdown__legend">
        {segments.map((s) => (
          <li key={s.label}>
            <span className="breakdown__dot" style={{ background: s.color }} aria-hidden="true" />
            {s.label} <strong>{s.value}</strong>
          </li>
        ))}
      </ul>
    </div>
  );
}

function formatDueDate(dueDate: string): string {
  const [year, month, day] = dueDate.split("-").map(Number);
  const date = new Date(year, month - 1, day);
  return date.toLocaleDateString(undefined, { month: "short", day: "numeric" });
}

export function Dashboard({ stats, statsLoading, statsError, tasks, onViewTasks }: DashboardProps) {
  const upcoming = tasks
    .filter((t) => t.status !== "completed")
    .slice()
    .sort((a, b) => a.dueDate.localeCompare(b.dueDate))
    .slice(0, 5);

  return (
    <div className="dashboard">
      {statsError && <div className="banner-error">{statsError}</div>}

      <div className="stat-grid" aria-busy={statsLoading}>
        <StatCard label="Total tasks" value={stats?.total ?? 0} accent="ink" />
        <StatCard label="Completed" value={stats?.completed ?? 0} accent="accent" />
        <StatCard label="Pending" value={stats?.pending ?? 0} accent="slate" />
        <StatCard label="In progress" value={stats?.inProgress ?? 0} accent="amber" />
        <StatCard label="High priority" value={stats?.highPriority ?? 0} accent="danger" hint="not yet done" />
        <StatCard label="Overdue" value={stats?.overdue ?? 0} accent="danger" />
        <StatCard label="Due in 7 days" value={stats?.upcomingWeek ?? 0} accent="amber" />
        <StatCard label="Favorites" value={stats?.favorites ?? 0} accent="slate" />
      </div>

      <div className="dashboard__panels">
        <section className="panel">
          <h2 className="panel__title">Status breakdown</h2>
          <BreakdownBar
            segments={[
              { label: STATUS_LABELS.completed, value: stats?.completed ?? 0, color: "var(--color-accent)" },
              { label: STATUS_LABELS.in_progress, value: stats?.inProgress ?? 0, color: "var(--color-amber)" },
              { label: STATUS_LABELS.pending, value: stats?.pending ?? 0, color: "var(--color-slate)" },
            ]}
          />
        </section>

        <section className="panel">
          <h2 className="panel__title">Upcoming deadlines</h2>
          {upcoming.length === 0 ? (
            <p className="panel__empty">Nothing due soon — you're all caught up.</p>
          ) : (
            <ul className="upcoming-list">
              {upcoming.map((t) => (
                <li key={t.id} className="upcoming-list__item">
                  <span className="upcoming-list__dot" style={{ background: t.color }} aria-hidden="true" />
                  <span className="upcoming-list__title">{t.title}</span>
                  <span className="upcoming-list__meta">
                    {PRIORITY_LABELS[t.priority]} · {formatDueDate(t.dueDate)}
                  </span>
                </li>
              ))}
            </ul>
          )}
          <button className="btn btn-ghost btn-sm panel__cta" onClick={onViewTasks}>
            View all tasks →
          </button>
        </section>
      </div>
    </div>
  );
}
