interface CategoryTagProps {
  category: string;
}

/** Renders nothing when there's no category, so callers can use it unconditionally. */
export function CategoryTag({ category }: CategoryTagProps) {
  if (!category) return null;
  return <span className="category-tag">{category}</span>;
}
