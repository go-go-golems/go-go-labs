// Film data extracted from documentation
export const filmData = {
  // Agfa films
  'agfa_apx100': { name: 'Agfa APX100', chartLetter: 'N' },
  'agfa_apx400': { name: 'Agfa APX400', chartLetter: 'R' },
  
  // Fuji films
  'fuji_acros100': { name: 'Fuji Acros 100', chartLetter: 'O' },
  'fuji_neopan_ss100': { name: 'Fuji Neopan SS 100', chartLetter: 'N' },
  'fuji_neopan400': { name: 'Fuji Neopan 400', chartLetter: 'N' },
  'fuji_neopan1600': { name: 'Fuji Neopan 1600', chartLetter: 'P' },
  
  // Ilford films
  'ilford_panf_plus': { name: 'Ilford PanF+', chartLetter: 'N' },
  'ilford_fp4_plus': { name: 'Ilford FP4+', chartLetter: 'N' },
  'ilford_hp5_plus': { name: 'Ilford HP5+', chartLetter: 'O' },
  'ilford_delta100': { name: 'Ilford Delta 100', chartLetter: 'N' },
  'ilford_delta400': { name: 'Ilford Delta 400', chartLetter: 'P' },
  'ilford_delta3200': { name: 'Ilford Delta 3200', chartLetter: 'S' },
  'ilford_sfx200': { name: 'Ilford SFX200', chartLetter: 'O' },
  
  // Kodak films
  'kodak_125px': { name: 'Kodak 125PX', chartLetter: 'N' },
  'kodak_400tx': { name: 'Kodak 400TX', chartLetter: 'O' },
  'kodak_tmax100': { name: 'Kodak T-MAX 100', chartLetter: 'O' },
  'kodak_400tmy': { name: 'Kodak 400TMY', chartLetter: 'P' },
  'kodak_3200tmz': { name: 'Kodak 3200TMZ', chartLetter: 'S' }
};

// Development times in seconds for each chart letter at different temperatures
export const chartTimes = {
  'L': { '18': 480, '20': 390, '22': 315, '24': 255 },
  'M': { '18': 555, '20': 450, '22': 360, '24': 300 },
  'N': { '18': 630, '20': 510, '22': 420, '24': 330 },
  'O': { '18': 750, '20': 600, '22': 480, '24': 390 },
  'P': { '18': 840, '20': 690, '22': 555, '24': 450 },
  'Q': { '18': 960, '20': 780, '22': 630, '24': 510 },
  'R': { '18': 1080, '20': 900, '22': 750, '24': 600 },
  'S': { '18': 1260, '20': 1020, '22': 840, '24': 675 },
  'T': { '18': 1500, '20': 1200, '22': 960, '24': 780 }
};

// Default process steps with their durations in seconds
export const defaultProcessSteps = [
  { id: 'prewet', name: 'Water Pre-wet', duration: 60, optional: true },
  { id: 'develop', name: 'Develop', duration: 0, optional: false }, // Duration will be calculated based on film and temperature
  { id: 'stop', name: 'Stop Bath', duration: 60, optional: false },
  { id: 'fix', name: 'Fix', duration: 180, optional: false },
  { id: 'prewash', name: 'Water Pre-wash', duration: 60, optional: true },
  { id: 'fixerRemover', name: 'Remove Fixer', duration: 180, optional: true },
  { id: 'wash', name: 'Water Wash', duration: 300, optional: false },
  { id: 'stabilize', name: 'Stabilize', duration: 60, optional: true }
];

// Push/pull adjustment logic
// Each push/pull stop typically moves 1-2 chart letters
export const letterOrder = ['L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T'];

export function adjustChartLetterForPushPull(chartLetter, pushPullValue) {
  const baseIndex = letterOrder.indexOf(chartLetter);
  const adjustedIndex = Math.max(0, Math.min(letterOrder.length - 1, baseIndex + pushPullValue));
  return letterOrder[adjustedIndex];
}

export function calculateDevelopmentTime(filmId, temperature, pushPullValue) {
  if (!filmId) return 0;
  
  const film = filmData[filmId];
  if (!film) return 0;
  
  let chartLetter = film.chartLetter;
  
  // Adjust chart letter based on push/pull value
  chartLetter = adjustChartLetterForPushPull(chartLetter, pushPullValue);
  
  // Get development time from chart
  return chartTimes[chartLetter][temperature.toString()] || 0;
} 