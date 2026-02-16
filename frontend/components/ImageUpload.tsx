"use client";

import { useRef, useState, useCallback } from "react";
import Button from "@/components/Button";

type ImageUploadProps = {
  onImageSelect: (file: File) => void;
  accept?: string;
  maxSizeMB?: number;
  label?: string;
  compact?: boolean;
};

export default function ImageUpload({
  onImageSelect,
  accept = "image/*",
  maxSizeMB = 5,
  label = "Upload image",
  compact = false,
}: ImageUploadProps) {
  const inputRef = useRef<HTMLInputElement>(null);
  const [isDragging, setIsDragging] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [localPreview, setLocalPreview] = useState<string | null>(null);

  const resetState = useCallback(() => {
    setLocalPreview(null);
    setError(null);
    if (inputRef.current) {
      inputRef.current.value = "";
    }
  }, []);

  const validateAndSelectFile = useCallback(
    (file: File) => {
      setError(null);

      const maxBytes = maxSizeMB * 1024 * 1024;
      if (file.size > maxBytes) {
        setError(`File size must be less than ${maxSizeMB}MB`);
        return;
      }

      if (accept !== "*" && !file.type.match(accept.replace("*", ".*"))) {
        setError("Invalid file type");
        return;
      }

      const reader = new FileReader();
      reader.onload = (e) => {
        setLocalPreview(e.target?.result as string);
        // Reset state after a short delay to allow the preview to be seen
        setTimeout(resetState, 500);
      };
      reader.readAsDataURL(file);

      onImageSelect(file);
    },
    [accept, maxSizeMB, onImageSelect, resetState]
  );

  const handleFileChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (file) {
      validateAndSelectFile(file);
    }
  };

  const handleDrop = useCallback(
    (event: React.DragEvent) => {
      event.preventDefault();
      setIsDragging(false);

      const file = event.dataTransfer.files?.[0];
      if (file) {
        validateAndSelectFile(file);
      }
    },
    [validateAndSelectFile]
  );

  const handleDragOver = (event: React.DragEvent) => {
    event.preventDefault();
    setIsDragging(true);
  };

  const handleDragLeave = () => {
    setIsDragging(false);
  };

  if (compact) {
    return (
      <div className="image-upload-compact">
        <input
          ref={inputRef}
          type="file"
          accept={accept}
          onChange={handleFileChange}
          style={{ display: "none" }}
        />
        <Button
          variant="ghost"
          onClick={() => inputRef.current?.click()}
          type="button"
        >
          📎 {label}
        </Button>
        {error && <span className="upload-error">{error}</span>}
        {localPreview && <span className="upload-success">✓ Image selected</span>}
      </div>
    );
  }

  return (
    <div className="image-upload">
      <input
        ref={inputRef}
        type="file"
        accept={accept}
        onChange={handleFileChange}
        style={{ display: "none" }}
      />

      <div
        className={`upload-dropzone ${isDragging ? "dragging" : ""} ${localPreview ? "uploaded" : ""}`}
        onDrop={handleDrop}
        onDragOver={handleDragOver}
        onDragLeave={handleDragLeave}
        onClick={() => inputRef.current?.click()}
      >
        <div className="dropzone-content">
          {localPreview ? (
            <>
              <span className="dropzone-icon success">✓</span>
              <span className="dropzone-text">Image selected</span>
            </>
          ) : (
            <>
              <span className="dropzone-icon">📷</span>
              <span className="dropzone-text">{label}</span>
              <span className="dropzone-hint">
                Drag & drop or click (max {maxSizeMB}MB)
              </span>
            </>
          )}
        </div>
      </div>

      {error && <p className="upload-error">{error}</p>}
    </div>
  );
}
