import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { foodApi, LogFoodRequest } from '../services/api';

export function useTodaySummary() {
  return useQuery({
    queryKey: ['food', 'summary', 'today'],
    queryFn: foodApi.getTodaySummary,
  });
}

export function useDateSummary(date: string) {
  return useQuery({
    queryKey: ['food', 'summary', date],
    queryFn: () => foodApi.getDateSummary(date),
    enabled: !!date,
  });
}

export function useLogFood() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (req: LogFoodRequest) => foodApi.log(req),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['food', 'summary', 'today'] });
      qc.invalidateQueries({ queryKey: ['food', 'history'] });
    },
  });
}

export function useDeleteFood() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => foodApi.deleteById(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['food', 'summary', 'today'] });
    },
  });
}

export function useDeleteLastFood() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: foodApi.deleteLast,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['food', 'summary', 'today'] });
    },
  });
}
