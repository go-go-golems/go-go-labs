import { TypedUseSelectorHook, useDispatch, useSelector } from 'react-redux/alternate-renderers';
import type { RootState, AppDispatch } from './store.js';

export const useAppDispatch: () => AppDispatch = useDispatch;
export const useAppSelector: TypedUseSelectorHook<RootState> = useSelector; 