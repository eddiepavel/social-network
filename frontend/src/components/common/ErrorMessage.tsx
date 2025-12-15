interface ErrorMessageProps {
  message: string;
  onDismiss?: () => void;
}

export function ErrorMessage({ message, onDismiss }: ErrorMessageProps) {
  if (!message) return null;

  return (
    <div className="flex items-start justify-between gap-4 rounded-xl border border-red-400/30 bg-red-500/10 px-4 py-3 text-sm text-red-100">
      <p className="leading-relaxed">{message}</p>
      {onDismiss && (
        <button
          onClick={onDismiss}
          className="text-lg leading-none text-red-100 hover:text-white transition"
        >
          ×
        </button>
      )}
    </div>
  );
}
