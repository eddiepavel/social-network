import { useEffect, useRef } from "react";

export default function useInfiniteScroll(cb: () => void, hasMore: boolean, isLoading: boolean) {
  const observer = useRef<IntersectionObserver | null>(null);
  const lastElementRef = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    if (isLoading) return;
    if (observer.current) observer.current.disconnect();
    observer.current = new window.IntersectionObserver((entries) => {
      if (entries[0].isIntersecting && hasMore) {
        cb();
      }
    });
    if (lastElementRef.current) {
      observer.current.observe(lastElementRef.current);
    }
    return () => {
      if (observer.current) observer.current.disconnect();
    };
  }, [isLoading, hasMore, cb]);

  return lastElementRef;
}
