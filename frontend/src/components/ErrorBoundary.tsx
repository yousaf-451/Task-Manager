import { Component } from "react";
import type { ErrorInfo, ReactNode } from "react";

interface ErrorBoundaryProps {
  children: ReactNode;
}

interface ErrorBoundaryState {
  error: Error | null;
}

/**
 * Catches unexpected render-time errors anywhere below it in the tree and
 * shows a friendly full-page fallback instead of a blank white screen.
 * This is distinct from the inline `error` state in useTasks (which
 * handles *expected* API failures) — this is the last line of defense
 * for genuine bugs.
 */
export class ErrorBoundary extends Component<ErrorBoundaryProps, ErrorBoundaryState> {
  state: ErrorBoundaryState = { error: null };

  static getDerivedStateFromError(error: Error): ErrorBoundaryState {
    return { error };
  }

  componentDidCatch(error: Error, info: ErrorInfo) {
    // In a real production deployment this is where you'd forward the
    // error to a monitoring service (Sentry, etc).
    console.error("Unhandled error caught by ErrorBoundary:", error, info.componentStack);
  }

  handleReload = () => {
    window.location.reload();
  };

  render() {
    if (this.state.error) {
      return (
        <div className="error-page">
          <div className="error-page__panel">
            <p className="error-page__eyebrow">Something went wrong</p>
            <h1 className="error-page__title">The dashboard hit an unexpected error</h1>
            <p className="error-page__message">
              Try reloading the page. If the problem keeps happening, please report it.
            </p>
            <button type="button" className="btn btn-accent" onClick={this.handleReload}>
              Reload page
            </button>
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}
