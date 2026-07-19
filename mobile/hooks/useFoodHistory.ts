import { useQuery } from '@tanstack/react-query';
import { foodApi } from '../services/api';

type Period = 'week' | 'month' | 'year';

export function useFoodHistory(period: Period) {
  return useQuery({
    queryKey: ['food', 'history', period],
    queryFn: () => foodApi.getHistory(period),
    staleTime: 5 * 60_000,
  });
}
