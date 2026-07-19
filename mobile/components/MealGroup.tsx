import { View, Text, TouchableOpacity, StyleSheet } from 'react-native';
import { useTranslation } from 'react-i18next';
import { FoodEntryResponse, MealGroup as MealGroupType } from '../services/api';

interface Props {
  group: MealGroupType;
  onDelete?: (id: number) => void;
}

export default function MealGroup({ group, onDelete }: Props) {
  const { t } = useTranslation();

  if (group.entries.length === 0) return null;

  const periodKey = group.period.toLowerCase() as 'morning' | 'afternoon' | 'evening';
  const label = t(periodKey);

  return (
    <View style={styles.container}>
      <Text style={styles.periodLabel}>{label}</Text>
      {group.entries.map((entry) => (
        <FoodRow key={entry.id} entry={entry} onDelete={onDelete} />
      ))}
    </View>
  );
}

function FoodRow({
  entry,
  onDelete,
}: {
  entry: FoodEntryResponse;
  onDelete?: (id: number) => void;
}) {
  return (
    <View style={styles.row}>
      <View style={styles.rowInfo}>
        <Text style={styles.foodName} numberOfLines={1}>
          {entry.food_item}
        </Text>
        <Text style={styles.macros}>
          {Math.round(entry.weight)}g · P{Math.round(entry.protein)}g F{Math.round(entry.fat)}g C{Math.round(entry.carbohydrates)}g
        </Text>
      </View>
      <View style={styles.rowRight}>
        <Text style={styles.calories}>{Math.round(entry.calories)} kcal</Text>
        {onDelete && (
          <TouchableOpacity onPress={() => onDelete(entry.id)} style={styles.deleteBtn}>
            <Text style={styles.deleteText}>✕</Text>
          </TouchableOpacity>
        )}
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    marginBottom: 8,
  },
  periodLabel: {
    fontSize: 13,
    fontWeight: '600',
    color: '#6c757d',
    marginBottom: 4,
    textTransform: 'uppercase',
    letterSpacing: 0.5,
  },
  row: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: '#fff',
    borderRadius: 10,
    paddingHorizontal: 14,
    paddingVertical: 10,
    marginBottom: 4,
    elevation: 1,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.05,
    shadowRadius: 2,
  },
  rowInfo: {
    flex: 1,
    marginRight: 8,
  },
  foodName: {
    fontSize: 15,
    fontWeight: '500',
    color: '#212529',
  },
  macros: {
    fontSize: 12,
    color: '#adb5bd',
    marginTop: 2,
  },
  rowRight: {
    alignItems: 'flex-end',
  },
  calories: {
    fontSize: 15,
    fontWeight: '600',
    color: '#2d6a4f',
  },
  deleteBtn: {
    marginTop: 4,
    padding: 2,
  },
  deleteText: {
    fontSize: 14,
    color: '#adb5bd',
  },
});
