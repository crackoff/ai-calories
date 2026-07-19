import { useState } from 'react';
import {
  ScrollView,
  View,
  Text,
  TouchableOpacity,
  StyleSheet,
  Alert,
  ActivityIndicator,
} from 'react-native';
import { useTranslation } from 'react-i18next';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useAuthStore } from '../../stores/authStore';
import { useSettingsStore } from '../../stores/settingsStore';
import LanguagePicker from '../../components/LanguagePicker';
import { userApi, paymentApi } from '../../services/api';

export default function SettingsScreen() {
  const { t } = useTranslation();
  const { logout } = useAuthStore();
  const { language, setLanguage } = useSettingsStore();
  const qc = useQueryClient();

  const [showLanguagePicker, setShowLanguagePicker] = useState(false);

  const { data: profile } = useQuery({
    queryKey: ['user', 'profile'],
    queryFn: userApi.getProfile,
  });

  const { data: payment } = useQuery({
    queryKey: ['payments', 'current'],
    queryFn: paymentApi.getCurrent,
  });

  const langMutation = useMutation({
    mutationFn: (lang: string) => userApi.updateLanguage(lang),
    onSuccess: (_, lang) => {
      setLanguage(lang);
      setShowLanguagePicker(false);
      qc.invalidateQueries({ queryKey: ['user', 'profile'] });
    },
    onError: () => Alert.alert('Error', t('error')),
  });

  const handleLogout = () => {
    Alert.alert(t('logout'), 'Are you sure?', [
      { text: t('cancel'), style: 'cancel' },
      { text: t('logout'), style: 'destructive', onPress: logout },
    ]);
  };

  const planLabel = payment ? `${payment.sku} (until ${payment.expiration_date.split('T')[0]})` : t('freePlan');

  return (
    <ScrollView style={styles.scroll} contentContainerStyle={styles.content}>
      {/* Profile section */}
      <Text style={styles.sectionLabel}>{t('profile')}</Text>
      <View style={styles.card}>
        <InfoRow label={t('email')} value={profile?.email ?? '—'} />
        <InfoRow label={t('currentPlan')} value={planLabel} />
      </View>

      {/* Language section */}
      <Text style={[styles.sectionLabel, { marginTop: 20 }]}>{t('language')}</Text>
      <TouchableOpacity
        style={styles.card}
        onPress={() => setShowLanguagePicker(!showLanguagePicker)}
      >
        <InfoRow label={t('language')} value={language} chevron />
      </TouchableOpacity>

      {showLanguagePicker && (
        <View style={styles.pickerWrapper}>
          {langMutation.isPending ? (
            <ActivityIndicator color="#2d6a4f" style={{ padding: 20 }} />
          ) : (
            <LanguagePicker value={language} onChange={(lang) => langMutation.mutate(lang)} />
          )}
        </View>
      )}

      {/* Logout */}
      <TouchableOpacity style={styles.logoutBtn} onPress={handleLogout}>
        <Text style={styles.logoutText}>{t('logout')}</Text>
      </TouchableOpacity>
    </ScrollView>
  );
}

function InfoRow({
  label,
  value,
  chevron,
}: {
  label: string;
  value: string;
  chevron?: boolean;
}) {
  return (
    <View style={styles.infoRow}>
      <Text style={styles.infoLabel}>{label}</Text>
      <View style={styles.infoRight}>
        <Text style={styles.infoValue} numberOfLines={1}>
          {value}
        </Text>
        {chevron && <Text style={styles.chevron}>›</Text>}
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  scroll: { flex: 1, backgroundColor: '#f8f9fa' },
  content: { padding: 20, paddingBottom: 40 },
  sectionLabel: {
    fontSize: 12,
    fontWeight: '600',
    color: '#6c757d',
    marginBottom: 8,
    textTransform: 'uppercase',
    letterSpacing: 0.8,
  },
  card: {
    backgroundColor: '#fff',
    borderRadius: 12,
    borderWidth: 1,
    borderColor: '#dee2e6',
    overflow: 'hidden',
  },
  pickerWrapper: {
    marginTop: 8,
    borderRadius: 12,
    overflow: 'hidden',
  },
  infoRow: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    paddingHorizontal: 16,
    paddingVertical: 14,
    borderBottomWidth: 1,
    borderBottomColor: '#f1f3f5',
  },
  infoLabel: {
    fontSize: 15,
    color: '#495057',
  },
  infoRight: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 4,
    maxWidth: '60%',
  },
  infoValue: {
    fontSize: 15,
    color: '#212529',
    fontWeight: '500',
    textAlign: 'right',
  },
  chevron: {
    fontSize: 20,
    color: '#adb5bd',
  },
  logoutBtn: {
    marginTop: 32,
    backgroundColor: '#fff',
    borderRadius: 12,
    paddingVertical: 16,
    alignItems: 'center',
    borderWidth: 1,
    borderColor: '#f4a261',
  },
  logoutText: {
    fontSize: 16,
    fontWeight: '600',
    color: '#e76f51',
  },
});
