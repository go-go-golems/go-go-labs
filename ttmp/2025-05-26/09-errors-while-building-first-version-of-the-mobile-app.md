# TypeScript Compilation Errors - Mobile Film Development App

## Current Issues

### 1. Expo Notifications API Type Mismatches

**Error:** 
```
Type '{ shouldShowAlert: true; shouldPlaySound: true; shouldSetBadge: false; }' is missing the following properties from type 'NotificationBehavior': shouldShowBanner, shouldShowList
```

**Location:** `src/services/notifications.ts:7`

**Issue:** The `setNotificationHandler` return type doesn't match expected `NotificationBehavior` interface in current Expo SDK version.

### 2. Notification Trigger Type Missing

**Error:**
```
Property 'type' is missing in type '{ seconds: number; }' but required in type 'TimeIntervalTriggerInput'
```

**Location:** `src/services/notifications.ts:63, 93`

**Issue:** Notification scheduling requires explicit `type` property with `Notifications.SchedulableTriggerInputTypes.TIME_INTERVAL`.

### 3. React Navigation Screen Component Types

**Error:**
```
Type 'FC<TimerScreenProps>' is not assignable to type 'ScreenComponentType<ParamListBase, "Timer">'
```

**Location:** `App.tsx:25`

**Issue:** Custom screen props interface conflicts with React Navigation's expected screen component signature.

## Potential Solutions

1. **Notifications:** Use `as any` type assertion or update to compatible Expo SDK version
2. **Navigation:** Simplify screen components to use `({ route, navigation }: any)` signature
3. **Trigger Types:** Explicitly specify `type: Notifications.SchedulableTriggerInputTypes.TIME_INTERVAL`

## Workaround Applied

- Changed screen components to use `any` types for navigation props
- Added explicit trigger types for notifications
- Need to resolve notification handler interface mismatch

## Status

App structure is complete but needs TypeScript compilation fixes before testing on device.

---

# Debug Playbook for Intern

Below is a concise "debug playbook" you can hand to the intern.
Each section shows 1) **why** the error appears (versions & typings that changed), 2) the **exact code fix**, and 3) links to the authoritative docs that show the new types so they can double-check.

---

## 1  `NotificationBehavior` shape changed in SDK ≥ 53

| Old field         | New fields (SDK 52/53+)              |
| ----------------- | ------------------------------------ |
| `shouldShowAlert` | `shouldShowBanner`, `shouldShowList` |

**Fix**

```ts
// src/services/notifications.ts
import * as Notifications from 'expo-notifications';

Notifications.setNotificationHandler({
  handleNotification: async (): Notifications.NotificationBehavior => ({
    // NEW field names ↓
    shouldShowBanner: true,
    shouldShowList : true,
    // still valid
    shouldPlaySound: true,
    shouldSetBadge : false,
  }),
});
```

The Expo docs' sample now lists the new keys – see the snippet in the reference page ([Expo Documentation][1])

---

## 2  `TimeIntervalTriggerInput` now requires `type`

The trigger union was tightened; every branch must declare its discriminator.

**Fix**

```ts
await Notifications.scheduleNotificationAsync({
  content: { title: 'Step complete', body: 'Move to next bath' },
  trigger: {
    // discriminator enum (string literal also works)
    type   : Notifications.SchedulableTriggerInputTypes.TIME_INTERVAL,
    seconds: remainingSec,
  } satisfies Notifications.TimeIntervalTriggerInput,
});
```

The new example in the docs shows the same pattern ([Expo Documentation][1])

---

## 3  React Navigation screen typings

### Why the error?

`ScreenComponentType<ParamListBase, 'Timer'>` (internal helper) expects **any React component** *whose props match the navigator-generated `navigation` and `route` fields*.
Wrapping your component in `React.FC<TimerScreenProps>` *after* you've already pulled `navigation`, `route` out of the generic can trip TypeScript's structural-typing check.

### Quick, safe pattern (works for Stack v6 & Native Stack)

```ts
// navigation/types.ts
import { NativeStackScreenProps } from '@react-navigation/native-stack';

export type RootStackParamList = {
  Home : undefined;
  Timer: { /* optional params */ };
};

// Make these props re-usable
export type TimerScreenProps = NativeStackScreenProps<
  RootStackParamList,
  'Timer'
>;
```

```tsx
// screens/TimerScreen.tsx
import { TimerScreenProps } from '../navigation/types';

export function TimerScreen({ navigation, route }: TimerScreenProps) {
  /* … */
}
```

```tsx
// App.tsx
<Stack.Screen name="Timer" component={TimerScreen} />
```

No more `FC<…>` wrapper is necessary – the function itself is the component.
React Navigation's TypeScript guide shows the same `NativeStackScreenProps` pattern ([React Navigation][2])

---

## 4  Keep your `expo-notifications` typings in sync

The above fixes assume you're already on the Expo SDK 52/53 line (bundled `expo-notifications ~0.31.x`).
If you see mismatched interfaces again after upgrading Expo, run:

```bash
npx expo install expo-notifications  # pins the lib to the SDK-compatible range
```

and commit the updated `package.json`.

---

## 5  Recommended ESLint/TSConfig guard rails

Add these two rules so future SDK migrations surface instantly:

```jsonc
// tsconfig.json
{
  "compilerOptions": {
    "strict": true,
    "exactOptionalPropertyTypes": true
  }
}
```

```jsonc
// .eslintrc
{
  "rules": {
    "@typescript-eslint/consistent-type-definitions": ["error", "type"]
  }
}
```

They force you (and the intern) to update objects when the upstream library adds or renames a field.

---

### Recap for the intern

1. **Update notification handler keys** (`shouldShowBanner`, `shouldShowList`).
2. **Always include `type`** inside `trigger` when scheduling.
3. **Type screens with `NativeStackScreenProps` (or `StackScreenProps`)** and export them as plain functions.
4. **Keep Expo modules and TypeScript strict** to catch the next breaking-type change early.

With these patches the project should compile cleanly and you can resume device testing. Happy coding!

[1]: https://docs.expo.dev/versions/latest/sdk/notifications/ "Notifications - Expo Documentation"
[2]: https://reactnavigation.org/docs/typescript/?utm_source=chatgpt.com "Type checking with TypeScript - React Navigation"