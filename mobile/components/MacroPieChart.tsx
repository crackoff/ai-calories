import { View, Text, StyleSheet } from 'react-native';
import Svg, { G, Path, Circle, Text as SvgText } from 'react-native-svg';
import { useTranslation } from 'react-i18next';
import { MacrosBreakdown } from '../services/api';

interface Props {
  macros: MacrosBreakdown;
  totalCalories: number;
}

const COLORS = { protein: '#52b788', fat: '#f4a261', carbs: '#e9c46a' };
const R = 80;  // outer radius
const IR = 50; // inner radius
const CX = 120;
const CY = 100;

function polarToCartesian(cx: number, cy: number, r: number, deg: number) {
  const rad = ((deg - 90) * Math.PI) / 180;
  return { x: cx + r * Math.cos(rad), y: cy + r * Math.sin(rad) };
}

function arcPath(cx: number, cy: number, r: number, ir: number, startDeg: number, endDeg: number) {
  const s = polarToCartesian(cx, cy, r, startDeg);
  const e = polarToCartesian(cx, cy, r, endDeg);
  const si = polarToCartesian(cx, cy, ir, startDeg);
  const ei = polarToCartesian(cx, cy, ir, endDeg);
  const large = endDeg - startDeg > 180 ? 1 : 0;
  return [
    `M ${s.x} ${s.y}`,
    `A ${r} ${r} 0 ${large} 1 ${e.x} ${e.y}`,
    `L ${ei.x} ${ei.y}`,
    `A ${ir} ${ir} 0 ${large} 0 ${si.x} ${si.y}`,
    'Z',
  ].join(' ');
}

export default function MacroPieChart({ macros, totalCalories }: Props) {
  const { t } = useTranslation();

  const slices = [
    { key: 'protein' as const, pct: macros.protein_pct, color: COLORS.protein, label: t('protein') },
    { key: 'fat'     as const, pct: macros.fat_pct,     color: COLORS.fat,     label: t('fat') },
    { key: 'carbs'   as const, pct: macros.carbs_pct,   color: COLORS.carbs,   label: t('carbs') },
  ];

  // Ensure we have something to draw even if all 0
  const total = slices.reduce((s, d) => s + (d.pct || 0), 0) || 1;
  let cursor = 0;
  const arcs = slices.map((s) => {
    const sweep = (s.pct / total) * 360;
    const path = sweep > 0.5 ? arcPath(CX, CY, R, IR, cursor, cursor + sweep) : null;
    cursor += sweep;
    return { ...s, path };
  });

  return (
    <View style={styles.container}>
      <Text style={styles.totalLabel}>
        {Math.round(totalCalories)} {t('kcal')}
      </Text>
      <Svg width={240} height={200}>
        <G>
          {arcs.map((a) =>
            a.path ? (
              <Path key={a.key} d={a.path} fill={a.color} />
            ) : null
          )}
        </G>
      </Svg>
      {/* Legend */}
      <View style={styles.legend}>
        {arcs.map((a) => (
          <View key={a.key} style={styles.legendItem}>
            <View style={[styles.legendDot, { backgroundColor: a.color }]} />
            <Text style={styles.legendText}>
              {a.label} {Math.round(a.pct)}%
            </Text>
          </View>
        ))}
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    alignItems: 'center',
    backgroundColor: '#fff',
    borderRadius: 16,
    paddingVertical: 12,
    marginHorizontal: 16,
    marginBottom: 12,
    elevation: 2,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.08,
    shadowRadius: 4,
  },
  totalLabel: {
    fontSize: 22,
    fontWeight: '700',
    color: '#2d6a4f',
    marginBottom: 4,
  },
  legend: {
    flexDirection: 'row',
    justifyContent: 'center',
    gap: 16,
    flexWrap: 'wrap',
    marginTop: 4,
    paddingHorizontal: 16,
    paddingBottom: 8,
  },
  legendItem: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 6,
  },
  legendDot: {
    width: 10,
    height: 10,
    borderRadius: 5,
  },
  legendText: {
    fontSize: 13,
    color: '#495057',
  },
});
