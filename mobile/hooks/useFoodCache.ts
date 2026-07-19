import { useState, useEffect, useRef } from 'react';
import { foodCacheApi, FoodCacheSearchResult } from '../services/api';

export function useFoodCacheSearch(query: string, debounceMs = 300) {
  const [results, setResults] = useState<FoodCacheSearchResult[]>([]);
  const [loading, setLoading] = useState(false);
  const timer = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    if (timer.current) clearTimeout(timer.current);

    const trimmed = query.trim();
    if (trimmed.length < 2) {
      setResults([]);
      return;
    }

    timer.current = setTimeout(async () => {
      setLoading(true);
      try {
        const data = await foodCacheApi.search(trimmed);
        setResults(data);
      } catch {
        setResults([]);
      } finally {
        setLoading(false);
      }
    }, debounceMs);

    return () => {
      if (timer.current) clearTimeout(timer.current);
    };
  }, [query, debounceMs]);

  return { results, loading };
}
