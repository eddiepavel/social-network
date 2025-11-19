interface ErrorMessageProps {
  message: string;
  onDismiss?: () => void;
}

export function ErrorMessage({ message, onDismiss }: ErrorMessageProps) {
  if (!message) return null;

  return (
    <div className="error-message">
      <p>{message}</p>
      {onDismiss && (
        <button onClick={onDismiss} className="dismiss-btn">
          ×
        </button>
      )}
    </div>
  );
}
