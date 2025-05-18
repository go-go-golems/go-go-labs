import { http, HttpResponse } from 'msw';
import type { Widget } from '../services/widgetsApi';

const widgets: Widget[] = [
  { id: 1, name: 'Foo' },
  { id: 2, name: 'Bar' },
];

export const handlers = [
  http.get('/api/widgets', () => {
    return HttpResponse.json(widgets);
  }),
]; 