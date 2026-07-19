import { useState, useMemo } from 'react';
import {
  View,
  Text,
  TextInput,
  TouchableOpacity,
  StyleSheet,
  ScrollView,
  KeyboardAvoidingView,
  Platform,
  ActivityIndicator,
  Alert,
} from 'react-native';
import { useRouter } from 'expo-router';
import { useTranslation } from 'react-i18next';
import FoodAutocomplete from '../../components/FoodAutocomplete';
import GramsKcalToggle from '../../components/GramsKcalToggle';
import NutritionPreview from '../../components/NutritionPreview';
import { FoodCacheSearchResult } from '../../services/api';
import { useLogFood } from '../../hooks/useFood';

type InputMode = 'grams' | 'kcal';

function calculatePreview(
  cached: FoodCacheSearchResult | null,
  inputMode: InputMode,
  value: number
) {
  if (!cached || value <= 0) return null;

  let grams: number;
  let calories: number;

  if (inputMode === 'kcal') {
    if (cached.calories_100g <= 0) return null;
    grams = (value / cached.calories_100g) * 100;
    calories = value;
  } else {
    grams = value;
    calories = (value / 100) * cached.calories_100g;
  }

  return {
    calories,
    grams,
    protein: (grams / 100) * cached.protein_100g,
    fat:     (grams / 100) * cached.fat_100g,
    carbs:   (grams / 100) * cached.carbs_100g,
  };
}

export default function AddFoodScreen() {
  const { t } = useTranslation();
  const router = useRouter();
  const { mutateAsync: logFood, isPending } = useLogFood();

  const [foodText, setFoodText] = useState('');
  const [selectedCache, setSelectedCache] = useState<FoodCacheSearchResult | null>(null);
  const [inputMode, setInputMode] = useState<InputMode>('grams');
  const [amountStr, setAmountStr] = useState('');

  const amountValue = parseFloat(amountStr) || 0;
  const preview = useMemo(
    () => calculatePreview(selectedCache, inputMode, amountValue),
    [selectedCache, inputMode, amountValue]
  );

  const handleSelect = (item: FoodCacheSearchResult) => {
    setSelectedCache(item);
    setFoodText(item.food_name);
  };

  const handleFoodTextChange = (text: string) => {
    setFoodText(text);
    // If user edits the text after selecting, deselect cache item
    if (selectedCache && text !== selectedCache.food_name) {
      setSelectedCache(null);
    }
  };

  const handleSave = async () => {
    const trimmed = foodText.trim();
    if (!trimmed) {
      Alert.alert('Validation', 'Please enter a food name');
      return;
    }
    if (amountValue <= 0) {
      Alert.alert('Validation', 'Please enter a valid amount');
      return;
    }

    try {
      if (selectedCache) {
        await logFood({
          food_cache_id: selectedCache.id,
          input_mode: inputMode,
          value: amountValue,
        });
      } else {
        await logFood({
          free_text: trimmed,
          input_mode: inputMode,
          value: amountValue,
        });
      }
      // Reset form and navigate to dashboard
      setFoodText('');
      setSelectedCache(null);
      setAmountStr('');
      router.replace('/(tabs)');
    } catch (err: any) {
      const message = err?.response?.data?.error ?? t('error');
      Alert.alert('Error', message);
    }
  };

  return (
    <KeyboardAvoidingView
      style={styles.flex}
      behavior={Platform.OS === 'ios' ? 'padding' : undefined}
    >
      <ScrollView
        style={styles.scroll}
        contentContainerStyle={styles.content}
        keyboardShouldPersistTaps="handled"
      >
        {/* Food input */}
        <Text style={styles.label}>{t('foodName')}</Text>
        <FoodAutocomplete
          value={foodText}
          onChange={handleFoodTextChange}
          onSelect={handleSelect}
          placeholder={t('searchFood')}
        />

        {/* Amount row */}
        <Text style={[styles.label, { marginTop: 20 }]}>{t('amount')}</Text>
        <View style={styles.amountRow}>
          <TextInput
            style={styles.amountInput}
            value={amountStr}
            onChangeText={setAmountStr}
            keyboardType="decimal-pad"
            placeholder="200"
          />
          <View style={styles.toggleWrapper}>
            <GramsKcalToggle
              value={inputMode}
              onChange={setInputMode}
              disabled={false}
            />
          </View>
        </View>

        {/* Nutrition preview (only for cached foods) */}
        {selectedCache && <NutritionPreview values={preview} />}

        {/* Save button */}
        <TouchableOpacity style={styles.saveBtn} onPress={handleSave} disabled={isPending}>
          {isPending ? (
            <ActivityIndicator color="#fff" />
          ) : (
            <Text style={styles.saveBtnText}>{t('save')}</Text>
          )}
        </TouchableOpacity>
      </ScrollView>
    </KeyboardAvoidingView>
  );
}

const styles = StyleSheet.create({
  flex: { flex: 1, backgroundColor: '#f8f9fa' },
  scroll: { flex: 1 },
  content: {
    padding: 20,
    paddingBottom: 48,
  },
  label: {
    fontSize: 13,
    fontWeight: '600',
    color: '#6c757d',
    marginBottom: 8,
    textTransform: 'uppercase',
    letterSpacing: 0.5,
  },
  amountRow: {
    flexDirection: 'row',
    gap: 10,
    alignItems: 'center',
  },
  amountInput: {
    backgroundColor: '#fff',
    borderWidth: 1,
    borderColor: '#dee2e6',
    borderRadius: 12,
    paddingHorizontal: 16,
    paddingVertical: 14,
    fontSize: 18,
    fontWeight: '600',
    width: 100,
    textAlign: 'center',
  },
  toggleWrapper: {
    flex: 1,
  },
  saveBtn: {
    backgroundColor: '#2d6a4f',
    borderRadius: 12,
    paddingVertical: 16,
    alignItems: 'center',
    marginTop: 28,
  },
  saveBtnText: {
    color: '#fff',
    fontSize: 16,
    fontWeight: '600',
  },
});
