'use client';

import Modal from './Modal';
import Button from './Button';

interface ConfirmDialogProps {
  isOpen: boolean;
  onClose: () => void;
  onConfirm: () => void;
  title: string;
  message: string;
  confirmText?: string;
  cancelText?: string;
  type?: 'danger' | 'warning' | 'info';
  isLoading?: boolean;
}

export function ConfirmDialog({
  isOpen,
  onClose,
  onConfirm,
  title,
  message,
  confirmText = 'Confirm',
  cancelText = 'Cancel',
  type = 'info',
  isLoading = false,
}: ConfirmDialogProps) {
  const icon = {
    danger: '⚠',
    warning: '⚠',
    info: 'ℹ',
  }[type];

  const confirmButtonVariant = type === 'danger' ? 'danger' : 'solid';

  return (
    <Modal isOpen={isOpen} onClose={onClose} title={title}>
      <div className="confirm-dialog-content">
        <div className="confirm-dialog-body">
          <span className={`confirm-dialog-icon confirm-dialog-icon-${type}`}>{icon}</span>
          <p className="confirm-dialog-message">{message}</p>
        </div>

        <div className="confirm-dialog-actions">
          <Button
            variant="ghost"
            onClick={onClose}
            disabled={isLoading}
          >
            {cancelText}
          </Button>
          <Button
            variant={confirmButtonVariant as any}
            onClick={onConfirm}
            disabled={isLoading}
          >
            {isLoading ? 'Processing...' : confirmText}
          </Button>
        </div>
      </div>
    </Modal>
  );
}