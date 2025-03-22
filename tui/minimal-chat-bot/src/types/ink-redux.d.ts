declare module 'ink-redux' {
  import { Store } from 'redux';
  import { TypedUseSelectorHook as ReduxTypedUseSelectorHook } from 'react-redux';

  // Re-export the types we're using from the hooks file
  export type TypedUseSelectorHook<T> = ReduxTypedUseSelectorHook<T>;
  export function useDispatch<T = any>(): T;
  export function useSelector<TState = any, TSelected = any>(
    selector: (state: TState) => TSelected,
    equalityFn?: (left: TSelected, right: TSelected) => boolean
  ): TSelected;
} 