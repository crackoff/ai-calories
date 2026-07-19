import { View, Text, StyleSheet } from 'react-native';
import { useTranslation } from 'react-i18next';

interface NutritionValues {
  calories: number;
  protein: number;
  fat: number;
  carbs: number;
  grams: number;
}

interface Props {
  values: NutritionValues | null;
}

export default function NutritionPreview({ values }: Props) {
  const { t } = useTranslation();

  if (!values) return null;

  const rows: Array<{ label: string; value: string }> = [
    { label: t('calories'), value: `${Math.round(values.calories)} kcal` },
    { label: t('protein'),  value: `${values.protein.toFixed(1)}g` },
    { label: t('fat'),      value: `${values.fat.toFixed(1)}g` },
    { label: t('carbs'),    value: `${values.carbs.toFixed(1)}g` },
    { label: t('weight'),   value: `${Math.round(values.grams)}g` },
  ];

  return (
    <View style={styles.container}>
      <Text style={styles.title}>{t('nutritionPreview')}</Text>
      {rows.map((r) => (
        <View key={r.label} style={styles.row}>
          <Text style={styles.label}>{r.label}</Text>
          <Text style={styles.value}>{r.value}</Text>
        </View>
      ))}
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    backgroundColor: '#f0faf4',
    borderRadius: 12,
    padding: 14,
    marginTop: 12,
    borderWidth: 1,
    borderColor: '#b7e4c7',
  },
  title: {
    fontSize: 13,
    fontWeight: '600',
    color: '#2d6a4f',
    marginBottom: 8,
    textTransform: 'uppercase',
    letterSpacing: 0.5,
  },
  row: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    paddingVertical: 4,
  },
  label: {
    fontSize: 14,
    color: '#495057',
  },
  value: {
    fontSize: 14,
    fontWeight: '600',
    color: '#212529',
  },
});
