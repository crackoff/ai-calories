import { View, Text, TouchableOpacity, StyleSheet } from 'react-native';

interface Language {
  code: string;
  label: string;
  flag: string;
}

const LANGUAGES: Language[] = [
  { code: 'en',    label: 'English (US)',            flag: '🇺🇸' },
  { code: 'es-419',label: 'Español (Latinoamérica)', flag: '🇲🇽' },
  { code: 'pt-BR', label: 'Português (Brasil)',       flag: '🇧🇷' },
  { code: 'ru',    label: 'Русский',                  flag: '🇷🇺' },
  { code: 'de',    label: 'Deutsch',                  flag: '🇩🇪' },
  { code: 'fr',    label: 'Français',                 flag: '🇫🇷' },
];

interface Props {
  value: string;
  onChange: (code: string) => void;
}

export default function LanguagePicker({ value, onChange }: Props) {
  return (
    <View style={styles.container}>
      {LANGUAGES.map((lang) => {
        const selected = lang.code === value;
        return (
          <TouchableOpacity
            key={lang.code}
            style={[styles.option, selected && styles.selected]}
            onPress={() => onChange(lang.code)}
          >
            <Text style={styles.flag}>{lang.flag}</Text>
            <Text style={[styles.label, selected && styles.selectedLabel]}>{lang.label}</Text>
            {selected && <Text style={styles.check}>✓</Text>}
          </TouchableOpacity>
        );
      })}
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    backgroundColor: '#fff',
    borderRadius: 12,
    overflow: 'hidden',
    borderWidth: 1,
    borderColor: '#dee2e6',
  },
  option: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingHorizontal: 16,
    paddingVertical: 14,
    borderBottomWidth: 1,
    borderBottomColor: '#f1f3f5',
  },
  selected: {
    backgroundColor: '#f0faf4',
  },
  flag: {
    fontSize: 20,
    marginRight: 12,
  },
  label: {
    flex: 1,
    fontSize: 15,
    color: '#212529',
  },
  selectedLabel: {
    fontWeight: '600',
    color: '#2d6a4f',
  },
  check: {
    fontSize: 16,
    color: '#2d6a4f',
    fontWeight: '700',
  },
});
