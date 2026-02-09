export default function EmptyState({
  title,
  body,
}: {
  title: string;
  body: string;
}) {
  return (
    <div className="surface card">
      <strong>{title}</strong>
      <p>{body}</p>
    </div>
  );
}
