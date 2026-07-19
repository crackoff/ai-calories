import {
  View,
  Text,
  TextInput,
  FlatList,
  TouchableOpacity,
  StyleSheet,
  ActivityIndicator,
} from 'react-native';
import { FoodCacheSearchResult } from '../services/api';
import { useFoodCacheSearch } from '../hooks/useFoodCache';

interface Props {
  value: string;
  onChange: (text: string) => void;
  onSelect: (item: FoodCacheSearchResult) => void;
  placeholder?: string;
}

export default function FoodAutocomplete({ value, onChange, onSelect, placeholder }: Props) {
  const { results, loading } = useFoodCacheSearch(value);

  const showDropdown = value.trim().length >= 2 && (results.length > 0 || loading);

  return (
    <View style={styles.wrapper}>
      <View style={styles.inputRow}>
        <TextInput
          style={styles.input}
          value={value}
          onChangeText={onChange}
          placeholder={placeholder ?? 'Search food...'}
          autoCapitalize="none"
          autoCorrect={false}
        />
        {loading && <ActivityIndicator size="small" color="#2d6a4f" style={styles.spinner} />}
        {value.length > 0 && (
          <TouchableOpacity onPress={() => onChange('')} style={styles.clear}>
            <Text style={styles.clearText}>✕</Text>
          </TouchableOpacity>
        )}
      </View>

      {showDropdown && (
        <View style={styles.dropdown}>
          <FlatList
            data={results}
            keyExtractor={(item) => String(item.id)}
            scrollEnabled={false}
            renderItem={({ item }) => (
              <TouchableOpacity style={styles.option} onPress={() => onSelect(item)}>
                <Text style={styles.optionName}>{item.food_name}</Text>
                <Text style={styles.optionCal}>{Math.round(item.calories_100g)} kcal/100g</Text>
              </TouchableOpacity>
            )}
          />
        </View>
      )}
    </View>
  );
}

const styles = StyleSheet.create({
  wrapper: {
    zIndex: 10,
  },
  inputRow: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: '#fff',
    borderWidth: 1,
    borderColor: '#dee2e6',
    borderRadius: 12,
    paddingHorizontal: 14,
  },
  input: {
    flex: 1,
    paddingVertical: 14,
    fontSize: 16,
  },
  spinner: {
    marginLeft: 8,
  },
  clear: {
    padding: 6,
    marginLeft: 4,
  },
  clearText: {
    fontSize: 14,
    color: '#adb5bd',
  },
  dropdown: {
    backgroundColor: '#fff',
    borderWidth: 1,
    borderColor: '#dee2e6',
    borderRadius: 12,
    marginTop: 4,
    overflow: 'hidden',
  },
  option: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingHorizontal: 14,
    paddingVertical: 12,
    borderBottomWidth: 1,
    borderBottomColor: '#f1f3f5',
  },
  optionName: {
    fontSize: 15,
    color: '#212529',
    flex: 1,
  },
  optionCal: {
    fontSize: 13,
    color: '#6c757d',
    marginLeft: 8,
  },
});
