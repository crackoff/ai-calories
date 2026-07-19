import { View, Text, StyleSheet, Dimensions } from 'react-native';
import Svg, { Path, G, Line, Text as SvgText } from 'react-native-svg';
import { useTranslation } from 'react-i18next';
import { HistoryDataPoint } from '../services/api';

interface Props {
  data: HistoryDataPoint[];
  period: 'week' | 'month' | 'year';
}

const COLORS = { protein: '#52b788', fat: '#f4a261', carbs: '#e9c46a' };

const SVG_W = Dimensions.get('window').width - 32;
const SVG_H = 200;
const PAD_L = 40;
const PAD_B = 30;
const PAD_T = 10;
const PAD_R = 12;
const PLOT_W = SVG_W - PAD_L - PAD_R;
const PLOT_H = SVG_H - PAD_T - PAD_B;

function shortLabel(date: string, period: 'week' | 'month' | 'year'): string {
  if (period === 'year') {
    const m = parseInt(date.split('-')[1] ?? '1', 10) - 1;
    return ['Jan','Feb','Mar','Apr','May','Jun','Jul','Aug','Sep','Oct','Nov','Dec'][m] ?? '';
  }
  if (period === 'week') {
    const d = new Date(date);
    return ['Su','Mo','Tu','We','Th','Fr','Sa'][d.getDay()] ?? '';
  }
  return date.split('-')[2] ?? '';
}

function buildStackedPaths(
  data: HistoryDataPoint[],
  maxVal: number
): { protein: string; fat: string; carbs: string } {
  const n = data.length;
  if (n === 0 || maxVal === 0) return { protein: '', fat: '', carbs: '' };

  const xs = data.map((_, i) => PAD_L + (i / Math.max(n - 1, 1)) * PLOT_W);
  const ys = (v: number) => PAD_T + PLOT_H - (v / maxVal) * PLOT_H;

  // Stacked: protein first, then protein+fat, then protein+fat+carbs
  const p  = data.map((d) => d.protein);
  const pf = data.map((d, i) => p[i] + d.fat);
  const pfc= data.map((d, i) => pf[i] + d.carbs);

  const linePath = (vals: number[]) =>
    vals.map((v, i) => `${i === 0 ? 'M' : 'L'} ${xs[i]} ${ys(v)}`).join(' ');

  const areaPath = (top: number[], bot: number[]) => {
    const forward = top.map((v, i) => `${i === 0 ? 'M' : 'L'} ${xs[i]} ${ys(v)}`).join(' ');
    const backward = [...bot].reverse().map((v, i, arr) => {
      const idx = arr.length - 1 - i;
      return `L ${xs[idx]} ${ys(v)}`;
    }).join(' ');
    return `${forward} ${backward} Z`;
  };

  const baseline = data.map(() => 0);

  return {
    protein: areaPath(p, baseline),
    fat:     areaPath(pf, p),
    carbs:   areaPath(pfc, pf),
  };
}

export default function NutritionHistoryChart({ data, period }: Props) {
  const { t } = useTranslation();

  if (!data || data.length === 0) {
    return (
      <View style={[styles.container, styles.empty]}>
        <Text style={styles.emptyText}>{t('loading')}</Text>
      </View>
    );
  }

  const maxVal = Math.max(...data.map((d) => d.protein + d.fat + d.carbs), 1);
  const paths = buildStackedPaths(data, maxVal);

  // Y-axis ticks
  const yTicks = [0, 0.25, 0.5, 0.75, 1].map((f) => ({
    y: PAD_T + PLOT_H - f * PLOT_H,
    label: Math.round(f * maxVal) + 'g',
  }));

  return (
    <View style={styles.container}>
      <Svg width={SVG_W} height={SVG_H}>
        {/* Y grid lines */}
        {yTicks.map((tick) => (
          <G key={tick.y}>
            <Line
              x1={PAD_L}
              y1={tick.y}
              x2={SVG_W - PAD_R}
              y2={tick.y}
              stroke="#f1f3f5"
              strokeWidth={1}
            />
            <SvgText
              x={PAD_L - 4}
              y={tick.y + 4}
              fontSize={9}
              fill="#adb5bd"
              textAnchor="end"
            >
              {tick.label}
            </SvgText>
          </G>
        ))}

        {/* Stacked areas */}
        {paths.protein && <Path d={paths.protein} fill={COLORS.protein} opacity={0.85} />}
        {paths.fat     && <Path d={paths.fat}     fill={COLORS.fat}     opacity={0.85} />}
        {paths.carbs   && <Path d={paths.carbs}   fill={COLORS.carbs}   opacity={0.85} />}

        {/* X-axis labels */}
        {data.map((d, i) => {
          const x = PAD_L + (i / Math.max(data.length - 1, 1)) * PLOT_W;
          const skip = data.length > 12 ? Math.ceil(data.length / 8) : 1;
          if (i % skip !== 0) return null;
          return (
            <SvgText
              key={d.date}
              x={x}
              y={SVG_H - 6}
              fontSize={9}
              fill="#6c757d"
              textAnchor="middle"
            >
              {shortLabel(d.date, period)}
            </SvgText>
          );
        })}
      </Svg>

      {/* Legend */}
      <View style={styles.legend}>
        {[
          { key: 'protein', color: COLORS.protein, label: t('protein') },
          { key: 'fat',     color: COLORS.fat,     label: t('fat') },
          { key: 'carbs',   color: COLORS.carbs,   label: t('carbs') },
        ].map((l) => (
          <View key={l.key} style={styles.legendItem}>
            <View style={[styles.dot, { backgroundColor: l.color }]} />
            <Text style={styles.legendText}>{l.label}</Text>
          </View>
        ))}
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    backgroundColor: '#fff',
    borderRadius: 16,
    marginHorizontal: 16,
    paddingTop: 8,
    paddingBottom: 12,
    marginBottom: 12,
    elevation: 2,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.08,
    shadowRadius: 4,
  },
  empty: { height: 200, justifyContent: 'center', alignItems: 'center' },
  emptyText: { color: '#adb5bd', fontSize: 14 },
  legend: {
    flexDirection: 'row',
    justifyContent: 'center',
    gap: 16,
    paddingHorizontal: 16,
  },
  legendItem: { flexDirection: 'row', alignItems: 'center', gap: 4 },
  dot: { width: 10, height: 10, borderRadius: 5 },
  legendText: { fontSize: 12, color: '#495057' },
});
