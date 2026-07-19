import { View, Text, TouchableOpacity, StyleSheet } from 'react-native';
import { useTranslation } from 'react-i18next';

type Mode = 'grams' | 'kcal';

interface Props {
  value: Mode;
  onChange: (mode: Mode) => void;
  disabled?: boolean;
}

export default function GramsKcalToggle({ value, onChange, disabled }: Props) {
  const { t } = useTranslation();

  return (
    <View style={[styles.container, disabled && styles.disabled]}>
      {(['grams', 'kcal'] as Mode[]).map((mode) => (
        <TouchableOpacity
          key={mode}
          style={[styles.btn, value === mode && styles.active]}
          onPress={() => !disabled && onChange(mode)}
          disabled={disabled}
        >
          <Text style={[styles.label, value === mode && styles.activeLabel]}>
            {mode === 'grams' ? t('grams') : t('kcal')}
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
    padding: 3,
    flex: 1,
  },
  disabled: {
    opacity: 0.4,
  },
  btn: {
    flex: 1,
    paddingVertical: 10,
    borderRadius: 8,
    alignItems: 'center',
  },
  active: {
    backgroundColor: '#2d6a4f',
  },
  label: {
    fontSize: 14,
    fontWeight: '500',
    color: '#6c757d',
  },
  activeLabel: {
    color: '#fff',
  },
});
