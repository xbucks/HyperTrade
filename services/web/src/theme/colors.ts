import { Indicators } from '../events/types';

export enum Colors {
  primary = '#1890ff',
  gray = '#9ca3af',
  lightGray = '#ddd',
  black = '#202020',
}

export const LineColors: Record<keyof Indicators, string> = {
  rsi: '#8884d8',
  adx: '#f6bd15',
  macd: '#5d7092',
  macd_hist: '#5ad8a6',
  macd_signal: '#5b8ff9',
};
