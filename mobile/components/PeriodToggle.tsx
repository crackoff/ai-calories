import { View, Text, TouchableOpacity, StyleSheet } from 'react-native';
import { useTranslation } from 'react-i18next';

type Period = 'week' | 'month' | 'year';

interface Props {
  value: Period;
  onChange: (p: Period) => void;
}

const PERIODS: Period[] = ['week', 'month', 'year'];

export default function PeriodToggle({ value, onChange }: Props) {
  const { t } = useTranslation();

  return (
    <View style={styles.container}>
      {PERIODS.map((p) => (
        <TouchableOpacity
          key={p}
          style={[styles.btn, value === p && styles.active]}
          onPress={() => onChange(p)}
        >
          <Text style={[styles.label, value === p && styles.activeLabel]}>
            {t(p)}
          </Text>
        </TouchableOpacity>
      ))}
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flexDirection: 'row',
    backgroundColor: '#e9ecef',
    borderRadius: 10,
    marginHorizontal: 16,
    padding: 3,
    marginBottom: 8,
  },
  btn: {
    flex: 1,
    paddingVertical: 8,
    borderRadius: 8,
    alignItems: 'center',
  },
  active: {
    backgroundColor: '#2d6a4f',
  },
  label: {
    fontSize: 13,
    fontWeight: '500',
    color: '#6c757d',
  },
  activeLabel: {
    color: '#fff',
  },
});
