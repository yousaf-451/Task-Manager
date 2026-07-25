interface BulkActionBarProps {
  selectedCount: number;
  onComplete: () => void;
  onDelete: () => void;
  onClear: () => void;
}

export function BulkActionBar({ selectedCount, onComplete, onDelete, onClear }: BulkActionBarProps) {
  if (selectedCount === 0) return null;

  return (
    <div className="bulk-bar" role="status">
      <span className="bulk-bar__count">{selectedCount} selected</span>
      <div className="bulk-bar__actions">
        <button className="btn btn-ghost btn-sm" onClick={onComplete}>
          Mark complete
        </button>
        <button className="btn btn-danger btn-sm" onClick={onDelete}>
          Delete
        </button>
        <button className="btn btn-ghost btn-sm" onClick={onClear}>
          Clear
        </button>
      </div>
    </div>
  );
}
