import { useState } from 'react';
import { ScrollView, View, Text, StyleSheet, ActivityIndicator } from 'react-native';
import { useTranslation } from 'react-i18next';
import MacroPieChart from '../../components/MacroPieChart';
import MealGroupComponent from '../../components/MealGroup';
import { useDateSummary } from '../../hooks/useFood';

function formatDate(d: Date): string {
  return d.toISOString().split('T')[0];
}

function DateNav({
  date,
  onPrev,
  onNext,
  isToday,
}: {
  date: Date;
  onPrev: () => void;
  onNext: () => void;
  isToday: boolean;
}) {
  const label = isToday
    ? 'Today'
    : date.toLocaleDateString(undefined, { weekday: 'short', month: 'short', day: 'numeric' });

  return (
    <View style={styles.dateNav}>
      <Text style={styles.dateNavBtn} onPress={onPrev}>{'‹'}</Text>
      <Text style={styles.dateLabel}>{label}</Text>
      <Text style={[styles.dateNavBtn, isToday && styles.disabled]} onPress={isToday ? undefined : onNext}>
        {'›'}
      </Text>
    </View>
  );
}

export default function HistoryScreen() {
  const { t } = useTranslation();
  const today = new Date();
  const [date, setDate] = useState(today);

  const isToday = formatDate(date) === formatDate(today);
  const dateStr = formatDate(date);

  const { data: summary, isLoading } = useDateSummary(dateStr);

  const prev = () => {
    const d = new Date(date);
    d.setDate(d.getDate() - 1);
    setDate(d);
  };
  const next = () => {
    if (isToday) return;
    const d = new Date(date);
    d.setDate(d.getDate() + 1);
    setDate(d);
  };

  return (
    <ScrollView style={styles.scroll} contentContainerStyle={styles.content}>
      <DateNav date={date} onPrev={prev} onNext={next} isToday={isToday} />

      {isLoading ? (
        <ActivityIndicator size="large" color="#2d6a4f" style={{ marginTop: 40 }} />
      ) : summary ? (
        <>
          <MacroPieChart
            macros={summary.macros_breakdown}
            totalCalories={summary.total_calories}
          />
          <Text style={styles.sectionLabel}>{t('todaysMeals')}</Text>
          <View style={styles.meals}>
            {summary.meals.map((group) => (
              <MealGroupComponent key={group.period} group={group} />
            ))}
          </View>
        </>
      ) : (
        <Text style={styles.emptyText}>{t('noEntriesToday')}</Text>
      )}
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  scroll: { flex: 1, backgroundColor: '#f8f9fa' },
  content: { paddingTop: 8, paddingBottom: 32 },
  dateNav: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    paddingHorizontal: 24,
    paddingVertical: 12,
    backgroundColor: '#fff',
    marginBottom: 12,
    borderBottomWidth: 1,
    borderBottomColor: '#f1f3f5',
  },
  dateNavBtn: {
    fontSize: 28,
    color: '#2d6a4f',
    paddingHorizontal: 8,
  },
  disabled: { color: '#dee2e6' },
  dateLabel: { fontSize: 17, fontWeight: '600', color: '#212529' },
  sectionLabel: {
    fontSize: 13,
    fontWeight: '600',
    color: '#6c757d',
    marginHorizontal: 16,
    marginTop: 12,
    marginBottom: 10,
    textTransform: 'uppercase',
    letterSpacing: 0.5,
  },
  meals: { paddingHorizontal: 16 },
  emptyText: {
    textAlign: 'center',
    color: '#adb5bd',
    fontSize: 14,
    marginTop: 48,
  },
});
