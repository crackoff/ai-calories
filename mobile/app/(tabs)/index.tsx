import { useState } from 'react';
import { ScrollView, View, Text, StyleSheet, RefreshControl, ActivityIndicator } from 'react-native';
import { useTranslation } from 'react-i18next';
import MacroPieChart from '../../components/MacroPieChart';
import NutritionHistoryChart from '../../components/NutritionHistoryChart';
import PeriodToggle from '../../components/PeriodToggle';
import MealGroupComponent from '../../components/MealGroup';
import { useTodaySummary } from '../../hooks/useFood';
import { useFoodHistory } from '../../hooks/useFoodHistory';
import { useDeleteFood } from '../../hooks/useFood';

type Period = 'week' | 'month' | 'year';

export default function DashboardScreen() {
  const { t } = useTranslation();
  const [period, setPeriod] = useState<Period>('week');

  const { data: summary, isLoading, refetch, isRefetching } = useTodaySummary();
  const { data: history } = useFoodHistory(period);
  const { mutate: deleteFood } = useDeleteFood();

  if (isLoading) {
    return (
      <View style={styles.center}>
        <ActivityIndicator size="large" color="#2d6a4f" />
      </View>
    );
  }

  const macros = summary?.macros_breakdown ?? { protein_pct: 0, fat_pct: 0, carbs_pct: 0 };
  const totalCalories = summary?.total_calories ?? 0;

  return (
    <ScrollView
      style={styles.scroll}
      contentContainerStyle={styles.content}
      refreshControl={<RefreshControl refreshing={isRefetching} onRefresh={refetch} colors={['#2d6a4f']} />}
    >
      {/* Macros pie chart */}
      <MacroPieChart macros={macros} totalCalories={totalCalories} />

      {/* Period toggle + stacked area chart */}
      <PeriodToggle value={period} onChange={setPeriod} />
      <NutritionHistoryChart data={history?.data ?? []} period={period} />

      {/* Today's meals grouped */}
      <Text style={styles.sectionLabel}>{t('todaysMeals')}</Text>

      {summary && summary.meals.every((m) => m.entries.length === 0) ? (
        <Text style={styles.emptyText}>{t('noEntriesToday')}</Text>
      ) : (
        <View style={styles.meals}>
          {summary?.meals.map((group) => (
            <MealGroupComponent
              key={group.period}
              group={group}
              onDelete={deleteFood}
            />
          ))}
        </View>
      )}
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  scroll: {
    flex: 1,
    backgroundColor: '#f8f9fa',
  },
  content: {
    paddingTop: 16,
    paddingBottom: 32,
  },
  center: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    backgroundColor: '#f8f9fa',
  },
  sectionLabel: {
    fontSize: 15,
    fontWeight: '600',
    color: '#495057',
    marginHorizontal: 16,
    marginTop: 8,
    marginBottom: 10,
    textTransform: 'uppercase',
    letterSpacing: 0.5,
  },
  meals: {
    paddingHorizontal: 16,
  },
  emptyText: {
    textAlign: 'center',
    color: '#adb5bd',
    fontSize: 14,
    marginTop: 24,
  },
});
