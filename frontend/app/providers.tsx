"use client";

import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { useState } from "react";
import { WebSocketProvider } from "@/hooks/useWebSocket";
import { ReactNode, createContext, useContext } from 'react';
import { useToast } from '@/hooks/useToast';
import { ToastContainer } from '@/components/Toast';

const ToastContext = createContext<ReturnType<typeof useToast> | null>(null);

export function useToastContext() {
  const context = useContext(ToastContext);
  if (!context) {
    throw new Error('useToastContext must be used within ToastProvider');
  }
  return context;
}

export default function Providers({ children }: { children: React.ReactNode }) {
  const toast = useToast();

  const [client] = useState(
    () =>
      new QueryClient({
        defaultOptions: {
          queries: {
            retry: 1,
            refetchOnWindowFocus: false,
          },
        },
      })
  );

  return (
    <QueryClientProvider client={client}>
      <WebSocketProvider>
        <ToastContext.Provider value={toast}>
          {children}
          <ToastContainer toasts={toast.toasts} onClose={toast.hideToast} />
        </ToastContext.Provider>
      </WebSocketProvider>
    </QueryClientProvider>
  );
}
